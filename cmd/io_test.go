package cmd

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, parent string, fileName string, content string) {
	filePath := filepath.Join(parent, fileName)

	enclosingDir := filepath.Dir(filePath)
	assert.NoError(t, os.MkdirAll(enclosingDir, 0750))

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0750)
	assert.NoError(t, err)
	_, err = file.WriteString(content)
	assert.NoError(t, err)
	assert.NoError(t, file.Close())
}

func TestFindBuildGradle(t *testing.T) {
	tempdir := t.TempDir()

	writeFile(t, tempdir, "build.gradle", "")
	writeFile(t, tempdir, "foo/build.gradle", "")
	writeFile(t, tempdir, "bar/build.gradle", "")
	writeFile(t, tempdir, "bar/bar/build.gradle", "")
	writeFile(t, tempdir, "bar/bar/bar/build.gradle", "")
	writeFile(t, tempdir, "bar/bar/bar/bar/build.gradle", "")

	actual, err := findBuildGradle(tempdir, 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(actual))

	actual, err = findBuildGradle(tempdir, 1, 0)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(actual))

	actual, err = findBuildGradle(tempdir, 2, 0)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(actual))

	actual, err = findBuildGradle(tempdir, 3, 0)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(actual))

	actual, err = findBuildGradle(tempdir, 4, 0)
	assert.NoError(t, err)
	assert.Equal(t, 6, len(actual))

	actual, err = findBuildGradle(tempdir, 5, 0)
	assert.NoError(t, err)
	assert.Equal(t, 6, len(actual))
}

func TestVersionExtractor(t *testing.T) {
	ext := compieLibraryVersionExtractor()
	match := extractVersion(ext, `

		implementation "foo:bar:1.2.3"
		implementation "foo:no-version"
		implementation "foo:variable:$myVar"
		api 'foo-bar:quax:4.5.6-b'
		testImplementation('a.b.c:foo-bar:1.2')

	`)
	assert.Equal(t, []Library{
		{Group: "foo", Name: "bar", Version: "1.2.3"},
		{Group: "foo", Name: "no-version", Version: "FIXME"},
		{Group: "foo", Name: "variable", Version: "$myVar"},
		{Group: "foo-bar", Name: "quax", Version: "4.5.6-b"},
		{Group: "a.b.c", Name: "foo-bar", Version: "1.2"},
	}, match)
}

func TestUpdateCatalog(t *testing.T) {
	catalog := initVersionCatalog()
	assert.Empty(t, catalog.Libraries)

	updateCatalog(catalog, []Library{})
	assert.Empty(t, catalog.Libraries)

	updateCatalog(catalog, []Library{
		{Group: "foo", Name: "bar", Version: "1.1"},
		{Group: "com.example.a123", Name: "d_A_S_h", Version: "1.2.3-M4"},
	})
	assert.Empty(t, catalog.Versions)
	// key is kebab-case
	assert.Equal(t, Libraries{
		"foo-bar": {
			"group":   "foo",
			"name":    "bar",
			"version": "1.1",
		},
		"com-example-a123-d-a-s-h": {
			"group":   "com.example.a123",
			"name":    "d_A_S_h",
			"version": "1.2.3-M4",
		},
	}, catalog.Libraries)

	updateCatalog(catalog, []Library{
		{Group: "foo", Name: "variable", Version: "$myVar"},
		{Group: "foo", Name: "variable-2", Version: "$myVar"},
	})
	assert.Equal(t, Versions{
		"myVar": "FIXME",
	}, catalog.Versions)
	assert.Equal(t, Libraries{
		"foo-bar": {
			"group":   "foo",
			"name":    "bar",
			"version": "1.1",
		},
		"foo-variable": {
			"group": "foo",
			"name":  "variable",
			"version": map[string]string{
				"ref": "myVar",
			},
		},
		"foo-variable-2": {
			"group": "foo",
			"name":  "variable-2",
			"version": map[string]string{
				"ref": "myVar",
			},
		},
		"com-example-a123-d-a-s-h": {
			"group":   "com.example.a123",
			"name":    "d_A_S_h",
			"version": "1.2.3-M4",
		},
	}, catalog.Libraries)
}
