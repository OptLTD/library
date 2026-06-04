package loader

import (
	"testing"
)

func TestParseInputsPresetRename(t *testing.T) {
	raw := `{
		"energy-fuel": {
			"uukey": "energy-fuel",
			"title": "加油",
			"preset": { "basic.energyType": "fuel" },
			"rename": { "basic.volume": "加油升数" },
			"replace": { "basic.energyType": { "editable": "NEVER" } },
			"fields": [".*"],
			"clicks": ["[SETUP][*]"]
		}
	}`
	loader := &JSONLoader{}
	inputs := loader.parseInputs(raw)
	item, ok := inputs["energy-fuel"]
	if !ok {
		t.Fatal("energy-fuel not found")
	}
	if item.Preset["basic.energyType"] != "fuel" {
		t.Fatalf("preset = %#v", item.Preset)
	}
	if item.Rename["basic.volume"] != "加油升数" {
		t.Fatalf("rename = %#v", item.Rename)
	}
	extra, ok := item.Replace["basic.energyType"].(map[string]any)
	if !ok || extra["editable"] != "NEVER" {
		t.Fatalf("replace = %#v", item.Replace)
	}
}

func TestParseTablesRename(t *testing.T) {
	raw := `{
		"energy-fuel": {
			"uukey": "energy-fuel",
			"title": "加油",
			"query": { "basic.energyType": "fuel" },
			"rename": { "basic.volume": "加油升数" }
		}
	}`
	loader := &JSONLoader{}
	tables := loader.parseTables(raw)
	item, ok := tables["energy-fuel"]
	if !ok {
		t.Fatal("energy-fuel not found")
	}
	if item.Rename["basic.volume"] != "加油升数" {
		t.Fatalf("rename = %#v", item.Rename)
	}
	if item.Query["basic.energyType"] != "fuel" {
		t.Fatalf("query = %#v", item.Query)
	}
}
