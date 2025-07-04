package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/stoewer/go-strcase"
	"io"
	"maps"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
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
		name := entry.Name()
		path := filepath.Join(root, name)
		if !entry.IsDir() && (strings.HasSuffix(name, ".gradle") || strings.HasSuffix(name, ".gradle.kts")) {
			baseName := filepath.Base(path)
			if currentDepth == 0 && (baseName == "settings.gradle" || baseName == "settings.gradle.kts") {
				// Skip the root settings.gradle file
				// https://discuss.gradle.org/t/how-to-use-version-catalog-in-the-root-settings-gradle-kts-file/44603/2
				continue
			}
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
		"annotationProcessor",
		"classpath",
		"implementation",
		"compileOnly",
		"compileOnlyApi",
		"platform",
		"integrationTestImplementation",
		"integrationTestRuntimeOnly",
		"runtimeOnly",
		"testImplementation",
		"testCompileOnly",
		"testRuntimeOnly",
	}
}

type StaticExtractors struct {
	plugin, library regexp.Regexp
}

func getStaticExtractors() StaticExtractors {
	return StaticExtractors{
		plugin:  compilePluginExtractor(),
		library: compieLibraryVersionExtractor(),
	}
}

func compieLibraryVersionExtractor() regexp.Regexp {
	configPattern := strings.Join(getConfigurations(), "|")
	libraryPattern := "(?P<group>[^:\"']+):(?P<name>[^:\"']+)(?::(?P<version>[^:\"']+)(?::(?P<classifier>[a-zA-Z0-9_-]+))?)?"
	return *regexp.MustCompile(fmt.Sprintf(`(?P<config>%s)\s*\(?["']%s["']\)?`, configPattern, libraryPattern))
}

func compilePluginExtractor() regexp.Regexp {
	return *regexp.MustCompile(`(\W)id\W+(?P<id>[\w.-]+)\W+version[ ("']+(?P<version>[\w.${}-]+)["')]+`)
}

func compileVersionVariableExtractor(keys []string) regexp.Regexp {
	combinedKeys := strings.Join(keys, "|")
	return *regexp.MustCompile(fmt.Sprintf(`\W(%s)\W?=\W*["']([^"']+)["']`, combinedKeys))
}

func compileVersionVariableInPropertyFileExtractor(keys []string) regexp.Regexp {
	combinedKeys := strings.Join(keys, "|")
	return *regexp.MustCompile(fmt.Sprintf(`\s*(%s)\W?=\W*([^"'\r\n]+)`, combinedKeys))
}

func extractTemp(extractor StaticExtractors, text string) (Versions, []Plugin, []StrictLibrary) {
	versions := make(Versions, 0)

	allMatchedLibs := extractor.library.FindAllStringSubmatch(text, -1)
	libs := make([]StrictLibrary, len(allMatchedLibs))
	for i, match := range allMatchedLibs {
		var version string
		if match[4] == "" {
			version = "FIXME"
		} else {
			version = match[4]
		}
		libs[i] = StrictLibrary{
			Group:   match[2],
			Name:    match[3],
			Version: version,
		}

		if strings.HasPrefix(version, "$") {
			key := extractVariableName(version)
			versions[key] = "FIXME"
			libs[i].Version = "$" + key
		}
	}

	allMatchedPlugins := extractor.plugin.FindAllStringSubmatch(text, -1)
	plugins := make([]Plugin, len(allMatchedPlugins))
	for i, match := range allMatchedPlugins {
		plugins[i] = Plugin{
			Id:      match[2],
			Version: match[3],
		}

		if strings.HasPrefix(match[3], "$") {
			key := extractVariableName(match[3])
			versions[key] = "FIXME"
			plugins[i].Version = LooseLibrary{
				"ref": key,
			}
		}
	}

	return versions, plugins, libs
}

func extractVersionInPropertyFile(extractor regexp.Regexp, text string) Versions {
	versions := make(Versions, 0)
	allMatchedLibs := extractor.FindAllStringSubmatch(text, -1)
	for _, match := range allMatchedLibs {
		key := match[1]
		version := match[2]
		versions[key] = version
	}
	return versions
}

var variableNameExtractor = regexp.MustCompile(`^\$(?:\{(.+)}|([^{}]+))$`)

func extractVariableName(name string) string {
	submatch := variableNameExtractor.FindStringSubmatch(name)
	if len(submatch[1]) > 0 {
		return submatch[1]
	}
	if len(submatch[2]) > 0 {
		return submatch[2]
	}
	return name
}

func extractVersioVariables(versions Versions, extractor regexp.Regexp, text string) {
	allMatches := extractor.FindAllStringSubmatch(text, -1)
	for _, match := range allMatches {
		key := match[1]
		version := match[2]
		versions[key] = version
	}
}

var nonIdChars = regexp.MustCompile("[^a-zA-Z0-9_-]+")
var numericFollowingSeparator = regexp.MustCompile("-[0-9]+")

func catalogSafeKey(lib StrictLibrary) string {
	combined := fmt.Sprintf("%s.%s", lib.Group, lib.Name)
	return safeKey(combined)
}

func catalogSafeKeyPlugin(lib Plugin) string {
	return safeKey(lib.Id)
}

func safeKey(src string) string {
	hyphenated := nonIdChars.ReplaceAllString(src, "-")
	kebab := strcase.KebabCase(hyphenated)
	return numericFollowingSeparator.ReplaceAllStringFunc(kebab, func(matched string) string {
		return strings.ReplaceAll(matched, "-", "")
	})
}

func updateCatalog(catalog VersionCatalog, libraries []StrictLibrary) {
	for _, lib := range libraries {
		var version any
		if strings.HasPrefix(lib.Version, "$") {
			trimmedVersion := lib.Version[1:]
			version = LooseLibrary{
				"ref": trimmedVersion,
			}
		} else {
			version = lib.Version
		}

		key := catalogSafeKey(lib)
		catalog.Libraries[key] = LooseLibrary{
			"group":   lib.Group,
			"name":    lib.Name,
			"version": version,
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

func embedReferenceToLibs(buildFilePaths []string) error {
	extractor := getStaticExtractors()

	for _, buildFilePath := range buildFilePaths {
		bytes, err := os.ReadFile(buildFilePath)
		if err != nil {
			return err
		}
		content := string(bytes)
		content = extractor.library.ReplaceAllStringFunc(content, func(s string) string {
			match := extractor.library.FindStringSubmatch(s)
			config := match[1]
			key := strings.ReplaceAll(catalogSafeKey(StrictLibrary{
				Group:   match[2],
				Name:    match[3],
				Version: match[4],
			}), "-", ".")
			classifier := match[5]
			if classifier == "" {
				return fmt.Sprintf("%s(libs.%s)", config, key)
			} else {
				return fmt.Sprintf(`%s(variantOf(libs.%s) { classifier("%s") })`, config, key, classifier)
			}
		})

		content = extractor.plugin.ReplaceAllStringFunc(content, func(s string) string {
			match := extractor.plugin.FindStringSubmatch(s)
			key := strings.ReplaceAll(catalogSafeKeyPlugin(Plugin{
				Id:      match[2],
				Version: match[3],
			}), "-", ".")
			leading := match[1]
			return fmt.Sprintf("%salias(libs.plugins.%s)", leading, key)
		})

		err = os.WriteFile(buildFilePath, []byte(content), 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

func searchLatestVersions(catalog VersionCatalog) {
	for _, library := range catalog.Libraries {
		if v, ok := library["version"].(string); ok && v == "FIXME" {
			newVer := searchMaven(library["group"].(string), library["name"].(string))
			library["version"] = newVer
		}
	}

	// Skip plugins since non-core plugins always have version
	// https://docs.gradle.org/current/userguide/plugins.html#sec:binary_plugin_locations
}

var client = &http.Client{
	Timeout: 5 * time.Second,
	Transport: &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   time.Second,
		ResponseHeaderTimeout: time.Second,
		IdleConnTimeout:       time.Second,
	},
}

type MavenResponse struct {
	Response struct {
		NumFound int `json:"numFound"`
		Docs     []struct {
			Version string `json:"v"`
		} `json:"docs"`
	} `json:"response"`
}

func searchMaven(group, name string) string {
	res, err := client.Get(fmt.Sprintf("https://search.maven.org/solrsearch/select?q=g:%s+AND+a:%s&rows=5&core=gav&wt=json", group, name))
	if err != nil {
		return "FIXME"
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return "FIXME"
	}
	data := MavenResponse{}
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return "FIXME"
	}
	if data.Response.NumFound == 0 {
		return "FIXME"
	}
	return data.Response.Docs[0].Version
}

func extractVersionCatalog(catalog VersionCatalog, buildFilePaths []string, variableDefFilePaths []string) (VersionCatalog, error) {
	extractor := getStaticExtractors()

	versionsAggregated := make(Versions, 0)
	librariesAggregated := make([]StrictLibrary, 0)
	pluginsAggregated := make([]Plugin, 0)

	for _, path := range buildFilePaths {
		bytes, err := os.ReadFile(path)
		if err != nil {
			return catalog, err
		}
		content := string(bytes)
		versions, plugins, libraries := extractTemp(extractor, content)
		librariesAggregated = append(librariesAggregated, libraries...)
		pluginsAggregated = append(pluginsAggregated, plugins...)
		maps.Copy(versionsAggregated, versions)
	}

	if len(versionsAggregated) > 0 {
		keys := make([]string, 0, len(versionsAggregated))
		for k := range versionsAggregated {
			keys = append(keys, k)
		}
		versionVariableExtractor := compileVersionVariableExtractor(keys)

		// two-path since version variable may be defined in other files
		for _, path := range buildFilePaths {
			bytes, err := os.ReadFile(path)
			if err != nil {
				return catalog, err
			}
			content := string(bytes)
			extractVersioVariables(versionsAggregated, versionVariableExtractor, content)
		}

		versionVariableInPropertyFileExtractor := compileVersionVariableInPropertyFileExtractor(keys)
		for _, path := range variableDefFilePaths {
			bytes, err := os.ReadFile(path)
			if err != nil {
				// Property file may not exist, so we can skip it
				continue
			}
			content := string(bytes)
			versionsInPropertyFile := extractVersionInPropertyFile(versionVariableInPropertyFileExtractor, content)
			maps.Copy(versionsAggregated, versionsInPropertyFile)
		}
	}

	if len(pluginsAggregated) > 0 {
		for _, plugin := range pluginsAggregated {
			key := catalogSafeKeyPlugin(plugin)
			catalog.Plugins[key] = plugin
		}
	}

	maps.Copy(catalog.Versions, versionsAggregated)
	updateCatalog(catalog, librariesAggregated)

	return catalog, nil
}
