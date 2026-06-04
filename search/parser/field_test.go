package parser

import (
	"testing"

	"github.com/OptLTD/library/search/consts"
	"github.com/OptLTD/library/search/source"
)

func TestApplyFieldRename(t *testing.T) {
	fields := []source.Field{
		{UUKey: "basic.volume", Label: "用量"},
		{UUKey: "basic.price", Label: "单价"},
	}
	applyFieldRename(fields, viewObject{
		"basic.volume": "加油升数",
	})
	if fields[0].Label != "加油升数" {
		t.Fatalf("volume label = %q", fields[0].Label)
	}
	if fields[1].Label != "单价" {
		t.Fatalf("price label should stay unchanged")
	}
}

func TestApplyFieldReplace(t *testing.T) {
	fields := []source.Field{
		{
			UUKey: "basic.energyType",
			Extra: source.FExtra{
				Required: true,
				Editable: consts.EDITABLE_ALWAYS,
			},
		},
	}
	applyFieldReplace(fields, viewObject{
		"basic.energyType": map[string]any{
			"editable": consts.EDITABLE_NEVER,
			"required": false,
		},
	})
	if fields[0].Extra.Editable != consts.EDITABLE_NEVER {
		t.Fatalf("editable = %q", fields[0].Extra.Editable)
	}
	if fields[0].Extra.Required {
		t.Fatalf("required should be false")
	}
}
