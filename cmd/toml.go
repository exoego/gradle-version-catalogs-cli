package cmd

import (
	"github.com/BurntSushi/toml"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"
)

type (
	Versions  = map[string]string
	Libraries = map[string]map[string]any
	Plugins   = map[string]Plugin
	Bundles   = map[string][]string

	VersionCatalog struct {
		Versions  Versions
		Libraries Libraries
		Plugins   Plugins
		Bundles   Bundles
	}

	Plugin struct {
		Id      string
		Version any
	}

	Library struct {
		Group   string
		Name    string
		Version string
	}
)

func ReadCatalog(path string) (*VersionCatalog, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	var catalog *VersionCatalog
	_, err := toml.DecodeFile(path, &catalog)
	if err != nil {
		return nil, err
	}
	return catalog, nil
}

func WriteCatalog(path string, catalog VersionCatalog) error {
	var builder strings.Builder
	builder.WriteString(writeVersions(catalog.Versions))
	builder.WriteString(writeLibraries(catalog.Libraries))
	builder.WriteString(writeBundles(catalog.Bundles))
	builder.WriteString(writePlugins(catalog.Plugins))
	return os.WriteFile(path, []byte(builder.String()), 0644)
}

func writeVersions(versions Versions) string {
	if len(versions) == 0 {
		return ""
	}
	var builder strings.Builder
	builder.WriteString("[versions]")
	builder.WriteString(LineBreak)
	for _, k := range slices.Sorted(maps.Keys(versions)) {
		builder.WriteString(k)
		builder.WriteString(" = ")
		builder.WriteString(strconv.Quote(versions[k]))
		builder.WriteString(LineBreak)
	}
	builder.WriteString(LineBreak)
	return builder.String()
}

func writeLibraries(libraries Libraries) string {
	if len(libraries) <= 0 {
		return ""
	}
	var builder strings.Builder
	builder.WriteString("[libraries]")
	builder.WriteString(LineBreak)
	for _, k := range slices.Sorted(maps.Keys(libraries)) {
		v := libraries[k]
		builder.WriteString(k)
		builder.WriteString(" = {")
		if module, ok := v["module"].(string); ok {
			builder.WriteString(" module = ")
			builder.WriteString(strconv.Quote(module))
		} else if group, ok := v["group"].(string); ok {
			builder.WriteString(" group = ")
			builder.WriteString(strconv.Quote(group))
			if name, ok := v["name"].(string); ok {
				builder.WriteString(", name = ")
				builder.WriteString(strconv.Quote(name))
			}
		}
		builder.WriteString(writeVersionEntry(v["version"]))
		builder.WriteString(" }")
		builder.WriteString(LineBreak)
	}
	builder.WriteString(LineBreak)
	return builder.String()
}

func writeVersionEntry(versionContainer any) string {
	if version, ok := versionContainer.(string); ok {
		var builder strings.Builder
		builder.WriteString(", version = ")
		builder.WriteString(strconv.Quote(version))
		return builder.String()
	}
	if version, ok := versionContainer.(map[string]any); ok {
		var builder strings.Builder
		if ref, ok := version["ref"].(string); ok && len(version) == 1 {
			builder.WriteString(", version.ref = ")
			builder.WriteString(strconv.Quote(ref))
			return builder.String()
		}
		builder.WriteString(", version = { ")

		written := false
		for _, vk := range []string{"ref", "strictly", "prefer", "require", "reject"} {
			if v, ok := version[vk].(string); ok {
				if written {
					builder.WriteString(", ")
				}
				written = true
				builder.WriteString(vk)
				builder.WriteString(" = ")
				builder.WriteString(strconv.Quote(v))
			}
		}
		if rejectAll, ok := version["rejectAll"].(bool); ok {
			builder.WriteString(", rejectAll = ")
			builder.WriteString(strconv.FormatBool(rejectAll))
		}
		builder.WriteString(" }")
		return builder.String()
	}
	return ""
}

func writeBundles(bundles Bundles) string {
	if len(bundles) == 0 {
		return ""
	}
	var builder strings.Builder
	builder.WriteString("[bundles]")
	builder.WriteString(LineBreak)
	for _, k := range slices.Sorted(maps.Keys(bundles)) {
		v := bundles[k]
		quoted := make([]string, len(v))
		for i, s := range v {
			quoted[i] = strconv.Quote(s)
		}
		builder.WriteString(k)
		builder.WriteString(" = [")
		builder.WriteString(strings.Join(quoted, ", "))
		builder.WriteString("]")
		builder.WriteString(LineBreak)
	}
	builder.WriteString(LineBreak)
	return builder.String()
}

func writePlugins(plugins Plugins) string {
	if len(plugins) == 0 {
		return ""
	}
	var builder strings.Builder
	builder.WriteString("[plugins]")
	builder.WriteString(LineBreak)
	for _, k := range slices.Sorted(maps.Keys(plugins)) {
		plugin := plugins[k]
		builder.WriteString(k)
		builder.WriteString(" = { id = ")
		builder.WriteString(strconv.Quote(plugin.Id))
		builder.WriteString(writeVersionEntry(plugin.Version))
		builder.WriteString(" }")
		builder.WriteString(LineBreak)
	}
	return builder.String()
}
