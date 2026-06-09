package loader

import (
	"context"
	"testing"

	"github.com/OptLTD/library/search/source"
	"github.com/OptLTD/library/search/support"
)

type loaderFunc func(ctx context.Context, model string) (*source.Value, error)

func (f loaderFunc) Load(ctx context.Context, model string) (*source.Value, error) {
	return f(ctx, model)
}

func TestMergeExtendedValue(t *testing.T) {
	base := &source.Value{
		Model: source.Model{
			UUKey:  "tms.waybill",
			Title:  "运单管理",
			Source: "app_tms_waybill",
			Search: "app_tms_waybill",
			Driver: "MONGODB",
		},
		Groups: map[string]source.Group{
			"basic": {UUKey: "basic", Model: "tms.waybill", Title: "基础信息"},
		},
		Fields: map[string]source.Field{
			"basic.uukey": {UUKey: "basic.uukey", Group: "basic"},
		},
	}
	overlay := []byte(`{"model":{"title":"运单（财务）","extends":"tms.waybill"}}`)
	merged, err := mergeExtendedValue(base, "fms.waybill", overlay)
	if err != nil {
		t.Fatal(err)
	}
	if merged.Model.UUKey != "fms.waybill" {
		t.Fatalf("uukey=%s", merged.Model.UUKey)
	}
	if merged.Model.Source != "app_tms_waybill" {
		t.Fatalf("source=%s", merged.Model.Source)
	}
	if merged.Model.Title != "运单（财务）" {
		t.Fatalf("title=%s", merged.Model.Title)
	}
	if merged.Groups["basic"].Model != "fms.waybill" {
		t.Fatalf("group model=%s", merged.Groups["basic"].Model)
	}
	if len(merged.Clicks) != 0 {
		t.Fatalf("clicks should not inherit from parent: %d", len(merged.Clicks))
	}
	if merged.Model.Extra["extends"] != "tms.waybill" {
		t.Fatalf("extends extra=%v", merged.Model.Extra["extends"])
	}
}

func TestResolveExtendsPhysical(t *testing.T) {
	child := &source.Value{
		Model: source.Model{
			UUKey:  "fms.waybill",
			Source: "app_fms_waybill",
			Search: "app_fms_waybill",
			Extra:  map[string]any{"extends": "tms.waybill"},
		},
	}
	parent := &source.Value{
		Model: source.Model{
			UUKey:   "tms.waybill",
			Source:  "app_tms_waybill",
			Search:  "app_tms_waybill",
			Driver:  "MONGODB",
		},
	}
	patchExtendsPhysical(child, parent)
	if child.Model.Source != "app_tms_waybill" {
		t.Fatalf("source=%s", child.Model.Source)
	}
	if child.Model.Search != "app_tms_waybill" {
		t.Fatalf("search=%s", child.Model.Search)
	}
	if child.Model.Driver != "MONGODB" {
		t.Fatalf("driver=%s", child.Model.Driver)
	}
}

func TestLoadExtendsParentCycle(t *testing.T) {
	ctx := context.Background()
	ctx, ok := markExtendsVisiting(ctx, "a")
	if !ok {
		t.Fatal("first visit should succeed")
	}
	ctx, ok = markExtendsVisiting(ctx, "b")
	if !ok {
		t.Fatal("second visit should succeed")
	}
	_, ok = markExtendsVisiting(ctx, "a")
	if ok {
		t.Fatal("cycle a->b->a should be rejected")
	}
	_, err := loadExtendsParent(ctx, nil, "a")
	if err != support.ExtendsCycle {
		t.Fatalf("err=%v", err)
	}
}
