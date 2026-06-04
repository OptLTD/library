package parser

import (
	"encoding/json"

	"github.com/OptLTD/library/search/source"

	"github.com/duke-git/lancet/v2/convertor"
)

type viewObject = map[string]any

// applyFieldRename 按 view 配置覆盖字段展示名（tables.json / inputs.json 的 rename）。
func applyFieldRename(fields []source.Field, rename viewObject) {
	if len(rename) == 0 || len(fields) == 0 {
		return
	}
	for i := range fields {
		raw, ok := rename[fields[i].UUKey]
		if !ok {
			continue
		}
		label, ok := raw.(string)
		if !ok || label == "" {
			continue
		}
		fields[i].Label = label
	}
}

// applyFieldReplace 按 view 配置覆盖 field.extra（replace 值为 FExtra 字段，不含 label/shown 等基础属性）。
func applyFieldReplace(fields []source.Field, replace viewObject) {
	if len(replace) == 0 || len(fields) == 0 {
		return
	}
	for i := range fields {
		raw, ok := replace[fields[i].UUKey]
		if !ok {
			continue
		}
		extraPatch, ok := raw.(map[string]any)
		if !ok || len(extraPatch) == 0 {
			continue
		}
		mergeFieldExtra(&fields[i].Extra, extraPatch)
	}
}

func mergeFieldExtra(dst *source.FExtra, patch map[string]any) {
	if dst == nil || len(patch) == 0 {
		return
	}
	base, err := convertor.StructToMap(*dst)
	if err != nil {
		return
	}
	for key, val := range patch {
		base[key] = val
	}
	data, err := json.Marshal(base)
	if err != nil {
		return
	}
	_ = json.Unmarshal(data, dst)
}

func applyFieldViewPatch(fields []source.Field, rename, replace viewObject) {
	applyFieldRename(fields, rename)
	applyFieldReplace(fields, replace)
}
