package loader

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/OptLTD/library/search/source"
	"github.com/OptLTD/library/search/support"

	parser "github.com/buger/jsonparser"
	"github.com/duke-git/lancet/v2/convertor"
)

func configExtendsModel(configJSON string) string {
	extends, err := parser.GetString([]byte(configJSON), "model", "extends")
	if err != nil || extends == "" {
		return ""
	}
	return strings.TrimSpace(extends)
}

func modelExtendsFromExtra(extra map[string]any) string {
	if extra == nil {
		return ""
	}
	extends, _ := extra["extends"].(string)
	return strings.TrimSpace(extends)
}

func mergeOverlayClicks(overlayJSON []byte) map[string]source.Click {
	clicks := map[string]source.Click{}
	if _, _, _, err := parser.Get(overlayJSON, "clicks"); err != nil {
		return clicks
	}
	parser.ObjectEach(overlayJSON, func(key []byte, value []byte, dataType parser.ValueType, offset int) error {
		click := source.Click{UUKey: string(key)}
		if json.Unmarshal(value, &click) == nil {
			clicks[string(key)] = click
		}
		return nil
	}, "clicks")
	return clicks
}

func mergeExtendedValue(base *source.Value, childName string, overlayJSON []byte) (*source.Value, error) {
	if base == nil {
		return nil, support.ConfigNotExsit
	}
	child := convertor.DeepClone(base)
	child.Model.UUKey = childName
	if title, err := parser.GetString(overlayJSON, "model", "title"); err == nil && title != "" {
		child.Model.Title = title
	}
	if brief, err := parser.GetString(overlayJSON, "model", "brief"); err == nil {
		child.Model.Brief = brief
	}
	if sourceName, err := parser.GetString(overlayJSON, "model", "source"); err == nil && sourceName != "" {
		child.Model.Source = sourceName
	}
	if searchName, err := parser.GetString(overlayJSON, "model", "search"); err == nil && searchName != "" {
		child.Model.Search = searchName
	}
	if driver, err := parser.GetString(overlayJSON, "model", "driver"); err == nil && driver != "" {
		child.Model.Driver = strings.ToUpper(driver)
	}
	extra, vtype, _, err := parser.Get(overlayJSON, "model", "extra")
	if err == nil && vtype.String() == "object" {
		overlayExtra := map[string]any{}
		if json.Unmarshal(extra, &overlayExtra) == nil {
			if child.Model.Extra == nil {
				child.Model.Extra = map[string]any{}
			}
			for key, val := range overlayExtra {
				child.Model.Extra[key] = val
			}
		}
	}
	if extends := configExtendsModel(string(overlayJSON)); extends != "" {
		if child.Model.Extra == nil {
			child.Model.Extra = map[string]any{}
		}
		child.Model.Extra["extends"] = extends
	}
	for key, group := range child.Groups {
		group.Model = childName
		child.Groups[key] = group
	}
	child.Clicks = mergeOverlayClicks(overlayJSON)
	return child, nil
}

// loadExtendsParent 跨 loader 加载父模型：优先 ctx 注入，否则回退当前 loader.Load。
func loadExtendsParent(ctx context.Context, ldr ILoader, extendsModel string) (*source.Value, error) {
	ctx, ok := markExtendsVisiting(ctx, extendsModel)
	if !ok {
		return nil, support.ExtendsCycle
	}
	if fn := GetExtendsLoader(ctx); fn != nil {
		return fn(ctx, extendsModel)
	}
	if ldr != nil {
		return ldr.Load(ctx, extendsModel)
	}
	return nil, support.ConfigNotExsit
}

func patchExtendsPhysical(child, parent *source.Value) {
	if child == nil || parent == nil {
		return
	}
	if parent.Model.Source != "" {
		child.Model.Source = parent.Model.Source
	}
	if parent.Model.Search != "" {
		child.Model.Search = parent.Model.Search
	}
	if parent.Model.Driver != "" {
		child.Model.Driver = parent.Model.Driver
	}
}

// ResolveExtends 运行时（DB loader）按 extra.extends 从父模型补齐物理表/索引。
func ResolveExtends(ctx context.Context, ldr ILoader, name string, value *source.Value) (*source.Value, error) {
	if value == nil {
		return value, nil
	}
	extends := modelExtendsFromExtra(value.Model.Extra)
	if extends == "" || extends == name {
		return value, nil
	}
	parent, err := loadExtendsParent(ctx, ldr, extends)
	if err != nil || parent == nil {
		return value, nil
	}
	patchExtendsPhysical(value, parent)
	if value.Model.Extra == nil {
		value.Model.Extra = map[string]any{}
	}
	value.Model.Extra["extends"] = extends
	return value, nil
}

func (self *JSONLoader) loadSchemaWithExtends(ctx context.Context, ldr ILoader, name, configJSON string) (*source.Value, error) {
	extends := configExtendsModel(configJSON)
	if extends == "" || extends == name {
		return self.parseSource(name, configJSON)
	}
	parent, err := loadExtendsParent(ctx, ldr, extends)
	if err != nil {
		return nil, err
	}
	if parent == nil {
		return nil, support.ConfigNotExsit
	}
	return mergeExtendedValue(parent, name, []byte(configJSON))
}
