package cmd

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseCatalogMinimum(t *testing.T) {
	got, err := ParseCatalog("../test/minimum.libs.version.toml")
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
