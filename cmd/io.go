package cmd

import (
	"errors"
	"fmt"
	"github.com/stoewer/go-strcase"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func getWorkingDirectory(args []string) (string, error) {
	if len(args) != 0 {
		return args[0], nil
	}
	// Print the current working directory
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return dir, nil
}

func openVersionCatalogFile(root string) (*os.File, error) {
	// Open the sub directory "gradle" under the root
	gradleDirPath := filepath.Join(root, "gradle")
	gradleDir, err := os.Open(gradleDirPath)
	if err != nil {
		return nil, fmt.Errorf("not a Gradle project seemingly: %s", gradleDirPath)
	}

	catalogFile, err := os.OpenFile(filepath.Join(gradleDir.Name(), "libs.versions.toml"), os.O_RDWR|os.O_CREATE, 0644)
	if errors.Is(err, os.ErrNotExist) {
		// empty
		_, err := catalogFile.WriteString("")
		if err != nil {
			return nil, err
		}
	}

	return catalogFile, nil
}

func findBuildGradle(root string, depth int, currentDepth int) ([]string, error) {
	var buildGradleFiles []string

	if currentDepth > depth {
		return buildGradleFiles, nil
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		fmt.Printf("Error reading directory %s: %v\n", root, err)
		return buildGradleFiles, nil // Continue even if one directory fails
	}

	for _, entry := range entries {
		path := filepath.Join(root, entry.Name())
		if !entry.IsDir() && (strings.HasSuffix(entry.Name(), ".gradle") || strings.HasSuffix(entry.Name(), ".gradle.kts")) {
			buildGradleFiles = append(buildGradleFiles, path)
		} else if entry.IsDir() {
			subFiles, err := findBuildGradle(path, depth, currentDepth+1)
			if err != nil {
				return nil, err // Propagate the error if needed
			}
			buildGradleFiles = append(buildGradleFiles, subFiles...)
		}
	}

	return buildGradleFiles, nil
}

func getConfigurations() []string {
	return []string{
		"api",
		"implementation",
		"compileOnly",
		"compileOnlyApi",
		"runtimeOnly",
		"testImplementation",
		"testCompileOnly",
		"testRuntimeOnly",
	}
}

func compieLibraryVersionExtractor() regexp.Regexp {
	configPattern := strings.Join(getConfigurations(), "|")

	// org.apache.httpcomponents:httpclient:4.5.13
	libraryPattern := "(?P<group>[^:\"']+):(?P<name>[^:\"']+)(?::(?P<version>[^:\"']+))?"

	return *regexp.MustCompile(fmt.Sprintf(`(?:%s)\s*\(?["']?%s`, configPattern, libraryPattern))
}

func extractVersion(extractor regexp.Regexp, text string) []Library {
	allMatches := extractor.FindAllStringSubmatch(text, -1)
	result := make([]Library, len(allMatches))
	for i, match := range allMatches {
		var version string
		if match[3] == "" {
			version = "FIXME"
		} else {
			version = match[3]
		}
		result[i] = Library{
			Group:   match[1],
			Name:    match[2],
			Version: version,
		}
	}
	return result
}

var nonIdChars = regexp.MustCompile("[^a-zA-Z0-9_-]+")

func catalogSafeKey(lib Library) string {
	combined := fmt.Sprintf("%s.%s", lib.Group, lib.Name)
	hyphenated := nonIdChars.ReplaceAllString(combined, "-")
	return strcase.KebabCase(hyphenated)
}

func updateCatalog(catalog VersionCatalog, libraries []Library) {
	for _, lib := range libraries {
		key := catalogSafeKey(lib)
		catalog.Libraries[key] = map[string]any{
			"group":   lib.Group,
			"name":    lib.Name,
			"version": lib.Version,
		}
	}
}

func initVersionCatalog() VersionCatalog {
	catalog := VersionCatalog{}
	catalog.Libraries = make(Libraries, 0)
	catalog.Bundles = make(Bundles, 0)
	catalog.Versions = make(Versions, 0)
	catalog.Plugins = make(Plugins, 0)
	return catalog
}

func extractVersionCatalog(buildFilePaths []string) (VersionCatalog, error) {
	catalog := initVersionCatalog()
	extractor := compieLibraryVersionExtractor()

	for _, path := range buildFilePaths {
		file, err := os.OpenFile(path, os.O_RDONLY, 0444)
		if err != nil {
			return catalog, err
		}
		bytes, err := io.ReadAll(file)
		if err != nil {
			return catalog, err
		}
		content := string(bytes)
		libraries := extractVersion(extractor, content)
		updateCatalog(catalog, libraries)
	}

	return catalog, nil
}
