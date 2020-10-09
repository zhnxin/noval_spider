package main

import (
	"testing"

	"github.com/BurntSushi/toml"
)

func TestTocConfig(t *testing.T) {
	CONFIG := ConfigPackage{}
	_, err := toml.DecodeFile("../conf.toml", &CONFIG)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range CONFIG.TocConfig {
		t.Log(c.Base)
		t.Log(c.Start)
		t.Log(c.Selector.Title)
		t.Log(c.Selector.Content)
	}
}
