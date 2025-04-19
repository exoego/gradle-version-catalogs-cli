package cmd

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestReadCatalog(t *testing.T) {
	got, err := ReadCatalog("../test/minimum.libs.versions.toml")
	assert.NoError(t, err)
	assert.Empty(t, got.Versions)
	assert.Empty(t, got.Plugins)
	assert.Empty(t, got.Bundles)
	assert.Equal(t, Libraries{
		"guava": {
			"group":   "com.google.guava",
			"name":    "guava",
			"version": "32.0.0-jre",
		},
		"foo-bar": {
			"group": "org.example",
			"name":  "foo-bar",
			"version": LooseLibrary{
				"ref": "bar",
			},
		},
		"awsJavaSdkDynamodb": {
			"module": "com.amazonaws:aws-java-sdk-dynamodb",
			"version": LooseLibrary{
				"ref": "awsJavaSdk",
			},
		},
		"commons-lang3": {
			"group": "org.apache.commons",
			"name":  "commons-lang3",
			"version": LooseLibrary{
				"strictly": "[3.8, 4.0[",
				"prefer":   "3.9",
			},
		},
		"mylib-full-format": {
			"group": "com.mycompany",
			"name":  "alternate",
			"version": LooseLibrary{
				"require":   "1.4",
				"reject":    "1.4.0",
				"rejectAll": false,
			},
		},
	}, got.Libraries)
}

func TestWriteCatalog(t *testing.T) {
	tempdir := t.TempDir()
	targetPath := filepath.Join(tempdir, "libs.versions.toml")

	got, err := ReadCatalog("../test/writer.libs.versions.toml")
	assert.NoError(t, err)

	err = WriteCatalog(targetPath, *got)
	assert.NoError(t, err)

	srcContent, err := os.ReadFile("../test/writer.libs.versions.toml")
	assert.NoError(t, err)
	generatedContent, err := os.ReadFile(targetPath)
	assert.NoError(t, err)
	assert.NoError(t, err)
	assert.Equal(t, string(srcContent), string(generatedContent))
}
