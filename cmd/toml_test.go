package cmd

import (
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestReadCatalog(t *testing.T) {
	got, err := ReadCatalog("../test/minimum.libs.version.toml")
	assert.NoError(t, err)
	assert.Empty(t, got.Versions)
	assert.Empty(t, got.Plugins)
	assert.Empty(t, got.Bundles)
	assert.Equal(t, map[string]map[string]any{
		"guava": {
			"group":   "com.google.guava",
			"name":    "guava",
			"version": "32.0.0-jre",
		},
		"foo-bar": {
			"group": "org.example",
			"name":  "foo-bar",
			"version": map[string]any{
				"ref": "bar",
			},
		},
		"awsJavaSdkDynamodb": {
			"module": "com.amazonaws:aws-java-sdk-dynamodb",
			"version": map[string]any{
				"ref": "awsJavaSdk",
			},
		},
		"commons-lang3": {
			"group": "org.apache.commons",
			"name":  "commons-lang3",
			"version": map[string]any{
				"strictly": "[3.8, 4.0[",
				"prefer":   "3.9",
			},
		},
		"mylib-full-format": {
			"group": "com.mycompany",
			"name":  "alternate",
			"version": map[string]any{
				"require":   "1.4",
				"reject":    "1.4.0",
				"rejectAll": false,
			},
		},
	}, got.Libraries)
}

func TestWriteCatalog(t *testing.T) {
	tempdir := t.TempDir()
	targetPath := filepath.Join(tempdir, "libs.version.toml")
	tempfile, err := os.OpenFile(targetPath, os.O_TRUNC|os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	assert.NoError(t, err)

	got, err := ReadCatalog("../test/writer.libs.version.toml")
	assert.NoError(t, err)

	err = WriteCatalog(tempfile, *got)
	assert.NoError(t, err)
	assert.NoError(t, tempfile.Close())

	srcFile, err := os.OpenFile("../test/writer.libs.version.toml", os.O_RDONLY, 0644)
	assert.NoError(t, err)
	srcContent, err := io.ReadAll(srcFile)
	assert.NoError(t, err)
	tempfile, err = os.OpenFile(targetPath, os.O_RDONLY, 0644)
	assert.NoError(t, err)
	generatedContent, err := io.ReadAll(tempfile)
	assert.NoError(t, err)
	assert.Equal(t, string(srcContent), string(generatedContent))
}
