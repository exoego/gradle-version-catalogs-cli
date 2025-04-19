package cmd

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"io"
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
	writer, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	if err := writeVersions(writer, catalog.Versions); err != nil {
		return err
	}
	if err := writeLibraries(writer, catalog.Libraries); err != nil {
		return err
	}
	if err := writeBundles(writer, catalog.Bundles); err != nil {
		return err
	}
	if err := writePlugins(writer, catalog.Plugins); err != nil {
		return err
	}
	return writer.Close()
}

func writeVersions(writer io.StringWriter, versions Versions) error {
	if len(versions) == 0 {
		return nil
	}
	_, err := writer.WriteString(fmt.Sprintf("[versions]%s", LineBreak))
	if err != nil {
		return err
	}
	for _, k := range slices.Sorted(maps.Keys(versions)) {
		_, err := writer.WriteString(fmt.Sprintf("%s = %s%s", k, strconv.Quote(versions[k]), LineBreak))
		if err != nil {
			return err
		}
	}
	_, err = writer.WriteString(LineBreak)
	if err != nil {
		return err
	}
	return nil
}

func writeLibraries(writer io.StringWriter, libraries Libraries) error {
	if len(libraries) <= 0 {
		return nil
	}
	_, err := writer.WriteString(fmt.Sprintf("[libraries]%s", LineBreak))
	if err != nil {
		return err
	}
	for _, k := range slices.Sorted(maps.Keys(libraries)) {
		v := libraries[k]
		_, err := writer.WriteString(fmt.Sprintf("%s = {", k))
		if err != nil {
			return err
		}
		if module, ok := v["module"].(string); ok {
			_, err := writer.WriteString(fmt.Sprintf(" module = %s", strconv.Quote(module)))
			if err != nil {
				return err
			}
		} else if group, ok := v["group"].(string); ok {
			_, err := writer.WriteString(fmt.Sprintf(" group = %s", strconv.Quote(group)))
			if err != nil {
				return err
			}
			if name, ok := v["name"].(string); ok {
				_, err := writer.WriteString(fmt.Sprintf(", name = %s", strconv.Quote(name)))
				if err != nil {
					return err
				}
			}
		}

		err = writeVersionEntry(writer, v["version"])
		if err != nil {
			return err
		}

		_, err = writer.WriteString(fmt.Sprintf(" }%s", LineBreak))
		if err != nil {
			return err
		}
	}
	_, err = writer.WriteString(LineBreak)
	if err != nil {
		return err
	}
	return nil
}

func writeVersionEntry(writer io.StringWriter, versionContainer any) error {
	if version, ok := versionContainer.(string); ok {
		_, err := writer.WriteString(fmt.Sprintf(", version = %s", strconv.Quote(version)))
		if err != nil {
			return err
		}
	} else if version, ok := versionContainer.(map[string]any); ok {
		if ref, ok := version["ref"].(string); ok && len(version) == 1 {
			_, err := writer.WriteString(fmt.Sprintf(", version.ref = %s", strconv.Quote(ref)))
			if err != nil {
				return err
			}

		} else {
			_, err := writer.WriteString(", version = { ")
			if err != nil {
				return err
			}

			written := false
			sep := ""
			for _, vk := range []string{"ref", "strictly", "prefer", "require", "reject"} {
				if v, ok := version[vk].(string); ok {
					if written {
						sep = ", "
					}
					written = true
					_, err := writer.WriteString(fmt.Sprintf("%s%s = %s", sep, vk, strconv.Quote(v)))
					if err != nil {
						return err
					}
				}
			}
			if rejectAll, ok := version["rejectAll"].(bool); ok {
				_, err := writer.WriteString(fmt.Sprintf(", rejectAll = %v", rejectAll))
				if err != nil {
					return err
				}
			}
			_, err = writer.WriteString(" }")
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func writeBundles(writer io.StringWriter, bundles Bundles) error {
	if len(bundles) == 0 {
		return nil
	}
	_, err := writer.WriteString(fmt.Sprintf("[bundles]%s", LineBreak))
	if err != nil {
		return err
	}
	for _, k := range slices.Sorted(maps.Keys(bundles)) {
		v := bundles[k]
		quoted := make([]string, len(v))
		for i, s := range v {
			quoted[i] = strconv.Quote(s)
		}
		_, err := writer.WriteString(fmt.Sprintf("%s = [%s]%s", k, strings.Join(quoted, ", "), LineBreak))
		if err != nil {
			return err
		}
	}
	_, err = writer.WriteString(LineBreak)
	if err != nil {
		return err
	}
	return nil
}

func writePlugins(writer io.StringWriter, plugins Plugins) error {
	if len(plugins) == 0 {
		return nil
	}
	_, err := writer.WriteString(fmt.Sprintf("[plugins]%s", LineBreak))
	if err != nil {
		return err
	}
	for _, k := range slices.Sorted(maps.Keys(plugins)) {
		plugin := plugins[k]
		_, err := writer.WriteString(fmt.Sprintf("%s = { id = %s", k, strconv.Quote(plugin.Id)))
		if err != nil {
			return err
		}

		err = writeVersionEntry(writer, plugin.Version)
		if err != nil {
			return err
		}

		_, err = writer.WriteString(fmt.Sprintf(" }%s", LineBreak))
		if err != nil {
			return err
		}
	}
	return nil
}
