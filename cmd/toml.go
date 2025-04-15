package cmd

import (
	"github.com/BurntSushi/toml"
	"os"
)

type (
	VersionCatalog struct {
		Versions  map[string]string
		Libraries map[string]map[string]any
		Plugins   map[string]Plugin
		Bundles   map[string][]string
	}

	Plugin struct {
		Id      string
		Version string
	}

	Library struct {
		Group   string
		Name    string
		Version string
	}
)

func ParseCatalog(path string) (*VersionCatalog, error) {
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
