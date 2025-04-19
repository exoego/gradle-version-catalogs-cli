package cmd

import (
	"github.com/stretchr/testify/assert"
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
