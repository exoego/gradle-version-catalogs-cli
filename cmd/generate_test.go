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
	os.Args = []string{"cli", "generate", tempdir, "extra-"}
	assert.ErrorContains(t, generateCommand.Execute(), "requires at most one arg")
}

func TestNoErrorIfGradleDirectory(t *testing.T) {
	tempdir := t.TempDir()
	writeFile(t, tempdir, "gradle/wrapper/dummy.txt", "")

	os.Args = []string{"cli", "generate", tempdir}
	assert.NoError(t, generateCommand.Execute())

	f, _ := os.ReadFile(filepath.Join(tempdir, "gradle", "libs.versions.toml"))
	assert.Equal(t, string(f), "", "Generates an empty libs.versions.toml")
}

func compareIgnoreLineBreaks(t *testing.T, expected, actual string) {
	// Remove all line breaks and spaces
	re := regexp.MustCompile(`\s+`)
	assert.Equal(t, re.ReplaceAllString(expected, "\n"), re.ReplaceAllString(actual, "\n"))
}

func TestSkipTopLevelSettingsFile(t *testing.T) {
	tempdir := t.TempDir()
	writeFile(t, tempdir, "gradle/wrapper/dummy.txt", "")
	writeFile(t, tempdir, "build.gradle", `
        classpath 'software.amazon.awssdk:s3'
		api("foo:foo:1.0")
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

	os.Args = []string{"cli", "generate", tempdir}
	assert.NoError(t, generateCommand.Execute())

	f, _ := os.ReadFile(filepath.Join(tempdir, "gradle", "libs.versions.toml"))
	compareIgnoreLineBreaks(t, `[libraries]
bar-bar = { group = "bar", name = "bar", version = "0.1" }
foo-foo = { group = "foo", name = "foo", version = "1.0" }
ok-ok = { group = "ok", name = "ok", version = "2.0" }
software-amazon-awssdk-s3 = { group = "software.amazon.awssdk", name = "s3", version = "FIXME" }

`, string(f))
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

	os.Args = []string{"cli", "generate", tempdir}
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
	`)

	os.Args = []string{"cli", "generate", tempdir}
	assert.NoError(t, generateCommand.Execute())

	f, _ := os.ReadFile(filepath.Join(tempdir, "gradle", "libs.versions.toml"))
	compareIgnoreLineBreaks(t, `[versions]
androidPluginVersion = "8.9.0"

[plugins]
com-android-application = { id = "com.android.application", version.ref = "androidPluginVersion" }
com-android-library = { id = "com.android.library", version.ref = "androidPluginVersion" }
org-jetbrains-kotlin-android = { id = "org.jetbrains.kotlin.android", version = "2.1.10" }
`, string(f))
}
