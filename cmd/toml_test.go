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
	assert.Equal(t, got.Libraries, map[string]Library{
		"guava": {
			Group:   "com.google.guava",
			Name:    "guava",
			Version: "32.0.0-jre",
		},
	})
}
