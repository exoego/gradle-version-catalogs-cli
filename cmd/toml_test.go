package cmd

import (
	"testing"
)

func TestParseCatalog(t *testing.T) {
	got, err := ParseCatalog("../test/libs.version.toml")
	if err != nil {
		t.Fatalf("unexpected err: %s", err)
	}
	if guava, ok := got.Libraries["guava"]; !ok {
		t.Fatalf("got: %v", got)
	} else if !(guava.Name == "guava" && guava.Group == "com.google.guava") || guava.Version != "32.0.0-jre" {
		t.Fatalf("Failed ")
	}
}
