package cmd

import (
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
)

type (
	VersionCatalog struct {
		Versions  map[string]string
		Libraries map[string]Library
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
	meta, err := toml.DecodeFile(path, &catalog)
	if err != nil {
		return nil, err
	}
	if len(meta.Undecoded()) > 0 {
		return nil, errors.New(fmt.Sprintf("Undecoded keys: %v", meta.Undecoded()))
	}
	return catalog, nil
}
