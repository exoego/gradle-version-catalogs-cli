package cmd

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestNoArgument(t *testing.T) {
	os.Args = []string{"cli", "generate"}
	assert.ErrorContains(t, generateCommand.Execute(), "not a Gradle project")
}

func TestExplicitPathNotAGradle(t *testing.T) {
	os.Args = []string{"cli", "generate", "./path/to/gradle/project"}
	assert.ErrorContains(t, generateCommand.Execute(), "not a Gradle project")
}

func TestGenerateRejectExtraArg(t *testing.T) {
	tempdir := t.TempDir()
	os.Args = []string{"cli", "generate", tempdir, "extra", "too-much"}
	assert.ErrorContains(t, generateCommand.Execute(), "requires at most two arg")
}

func TestGenerateRejectUnknownFlag(t *testing.T) {
	tempdir := t.TempDir()
	os.Args = []string{"cli", "generate", tempdir, "--huh"}
	assert.ErrorContains(t, generateCommand.Execute(), "unknown flag")
}

func TestGenerateRejectInvalidValue(t *testing.T) {
	tempdir := t.TempDir()
	os.Args = []string{"cli", "generate", tempdir, "--auto-latest=foo"}
	assert.ErrorContains(t, generateCommand.Execute(), `invalid argument "foo"`)
}

func TestNoErrorIfGradleDirectory(t *testing.T) {
	tempdir := t.TempDir()
	writeFile(t, tempdir, "gradle/wrapper/dummy.txt", "")

	os.Args = []string{"cli", "generate", tempdir}
	assert.NoError(t, generateCommand.Execute())

	f, _ := os.ReadFile(filepath.Join(tempdir, "gradle", "libs.versions.toml"))
	assert.Equal(t, string(f), "", "Generates an empty libs.versions.toml")
}

func TestSkipTopLevelSettingsFile(t *testing.T) {
	tempdir := t.TempDir()
	writeFile(t, tempdir, "gradle/wrapper/dummy.txt", "")
	writeFile(t, tempdir, "build.gradle", `
		implementation "foo.sub:No-Version"
        classpath 'software.amazon.awssdk:s3'
		api("foo:foo:1.0-M4")
	    testImplementation("org.apache.flink:flink-runtime_2.12:1.13.2")
        implementation("org.openjfx:javafx-base:11.0.2:win")
	`)
	writeFile(t, tempdir, "settings.gradle.kts", `
		implementation("ignore:ignore:1.0")
	`)
	writeFile(t, tempdir, "foo/build.gradle", `
		testImplementation("bar:bar:0.1")
	`)
	writeFile(t, tempdir, "foo/settings.gradle.kts", `
		implementation("ok:ok:2.0")	
	`)
	writeFile(t, tempdir, "too/much/deep/should/be/ignored/build.gradle", `
		testImplementation("no:no:0.1")
	`)

	os.Args = []string{"cli", "generate", tempdir, "--auto-latest=false"}
	assert.NoError(t, generateCommand.Execute())

	f, _ := os.ReadFile(filepath.Join(tempdir, "gradle", "libs.versions.toml"))
	compareIgnoreLineBreaks(t, `[libraries]
bar-bar = { group = "bar", name = "bar", version = "0.1" }
foo-foo = { group = "foo", name = "foo", version = "1.0-M4" }
foo-sub-no-version = { group = "foo.sub", name = "No-Version", version = "FIXME" }
ok-ok = { group = "ok", name = "ok", version = "2.0" }
org-apache-flink-flink-runtime212 = { group = "org.apache.flink", name = "flink-runtime_2.12", version = "1.13.2" }
org-openjfx-javafx-base = { group = "org.openjfx", name = "javafx-base", version = "11.0.2" }
software-amazon-awssdk-s3 = { group = "software.amazon.awssdk", name = "s3", version = "FIXME" }
`, string(f))

	f2, _ := os.ReadFile(filepath.Join(tempdir, "build.gradle"))
	compareIgnoreLineBreaks(t, `
		implementation(libs.foo.sub.no.version)
		classpath(libs.software.amazon.awssdk.s3)
		api(libs.foo.foo)
		testImplementation(libs.org.apache.flink.flink.runtime212)
		implementation(variantOf(libs.org.openjfx.javafx.base) { classifier("win") })
`, string(f2))

}

func TestVariableSupport(t *testing.T) {
	tempdir := t.TempDir()
	writeFile(t, tempdir, "gradle/wrapper/dummy.txt", "")
	writeFile(t, tempdir, "build.gradle", `
        val fooVersion = "1.0"
		api("foo:foo:$fooVersion")
		api("bar:bar:${barVersion}")
	`)
	writeFile(t, tempdir, "foo/build.gradle", `
		testImplementation("foo:foo-ext:${fooVersion}")
	`)

	os.Args = []string{"cli", "generate", tempdir, "--auto-latest=false"}
	assert.NoError(t, generateCommand.Execute())

	f, _ := os.ReadFile(filepath.Join(tempdir, "gradle", "libs.versions.toml"))
	compareIgnoreLineBreaks(t, `[versions]
barVersion = "FIXME"
fooVersion = "1.0"

[libraries]
bar-bar = { group = "bar", name = "bar", version.ref = "barVersion" }
foo-foo = { group = "foo", name = "foo", version.ref = "fooVersion" }
foo-foo-ext = { group = "foo", name = "foo-ext", version.ref = "fooVersion" }

`, string(f))
}

func TestPluginsSupport(t *testing.T) {
	tempdir := t.TempDir()
	writeFile(t, tempdir, "gradle/wrapper/dummy.txt", "")
	writeFile(t, tempdir, "build.gradle", `
		val androidPluginVersion = "8.9.0"
		id("com.android.application") version "${androidPluginVersion}" apply false
		id("com.android.library") version "$androidPluginVersion"
		id("org.jetbrains.kotlin.android") version "2.1.10" apply false
		id("foo.bar-buz") version "2.2.20-123"
	`)

	os.Args = []string{"cli", "generate", tempdir, "--auto-latest=false"}
	assert.NoError(t, generateCommand.Execute())

	f, _ := os.ReadFile(filepath.Join(tempdir, "gradle", "libs.versions.toml"))
	compareIgnoreLineBreaks(t, `[versions]
androidPluginVersion = "8.9.0"

[plugins]
com-android-application = { id = "com.android.application", version.ref = "androidPluginVersion" }
com-android-library = { id = "com.android.library", version.ref = "androidPluginVersion" }
foo-bar-buz = { id = "foo.bar-buz", version = "2.2.20-123" }
org-jetbrains-kotlin-android = { id = "org.jetbrains.kotlin.android", version = "2.1.10" }
`, string(f))
}

func TestMergeEmpty(t *testing.T) {
	originalBytes, _ := os.ReadFile("../test/writer.libs.versions.toml")
	tempdir := t.TempDir()
	writeFile(t, tempdir, "gradle/libs.versions.toml", string(originalBytes))
	writeFile(t, tempdir, "build.gradle", `
		// empty
	`)

	os.Args = []string{"cli", "generate", tempdir, "--auto-latest=false"}
	assert.NoError(t, generateCommand.Execute())
	f, _ := os.ReadFile(filepath.Join(tempdir, "gradle", "libs.versions.toml"))
	compareIgnoreLineBreaks(t, string(originalBytes), string(f))
}

func TestAutoLatestDependency(t *testing.T) {
	tempdir := t.TempDir()
	writeFile(t, tempdir, "gradle/wrapper/dummy.txt", "")
	writeFile(t, tempdir, "build.gradle", `
		api("org.apache.logging.log4j:log4j-core")
		id("org.gradle.kotlin.embedded-kotlin") version "2.1.10"
	`)

	os.Args = []string{"cli", "generate", tempdir, "--auto-latest=true"}
	assert.NoError(t, generateCommand.Execute())

	f, _ := os.ReadFile(filepath.Join(tempdir, "gradle", "libs.versions.toml"))
	actual := string(f)
	assert.Contains(t, actual, "[libraries]", "Missing [libraries] section")
	assert.NotContains(t, actual, `"FIXME"`, "Should not contain FIXME")
	assert.Regexp(t, regexp.MustCompile(`name = "log4j-core", version = "\d+\.\d+\.\d+`), actual)
}
