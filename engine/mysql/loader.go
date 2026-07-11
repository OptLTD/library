package mysql

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/OptLTD/library/search/consts"
	"github.com/OptLTD/library/search/loader"
	"github.com/OptLTD/library/search/source"
	"github.com/OptLTD/library/search/support"

	"github.com/duke-git/lancet/v2/maputil"
	"github.com/duke-git/lancet/v2/slice"
	"gorm.io/gorm"
)

type Loader struct {
	client *gorm.DB
	tables *loader.SchemaTableName
}

func NewLoader(db *gorm.DB, tables *loader.SchemaTableName) loader.ILoader {
	session := &gorm.Session{QueryFields: true}
	return &Loader{
		client: db.Session(session),
		tables: tables,
	}
}

func (self *Loader) Load(ctx context.Context, model string) (*source.Value, error) {
	logID := loader.GetLogID(ctx)
	if self.client == nil || self.tables == nil {
		return nil, support.ErrorConfigClient
	}
	if self.tables.Model == "" || self.tables.Table == "" {
		return nil, support.ErrorConfigTables
	}
	where1 := map[string]any{"uukey": model}
	where2 := map[string]any{"model": model}
	if scope := loader.GetScope(ctx); scope != nil {
		where1 = maputil.Merge(where1, scope)
		where2 = maputil.Merge(where2, scope)
	}
	detail, err := self.getModel(where1)
	if err != nil {
		return nil, support.ErrorModelSource
	}
	clicks := map[string]source.Click{}
	for _, action := range detail.Clicks {
		clicks[action.UUKey] = action
	}

	groups := map[string]source.Group{}
	for _, group := range detail.Groups {
		group.GType = strings.ToUpper(group.GType)
		groups[group.UUKey] = group
	}
	groups["basic"] = loader.ResetDefaultGroup(groups)
	fields := map[string]source.Field{}
	for _, item := range detail.Fields {
		group, _ := groups[item.Group]
		item.GName = group.Title
		item.GType = group.GType
		item.UUKey = item.GetKey()
		item.SeqNo = item.SeqNo + group.SeqNo*1000
		switch strings.ToUpper(group.GType) {
		case consts.GTYPE_FLATTEN:
			item.Index = item.Field
		case consts.GTYPE_GROUPED:
			item.Index = item.UUKey
		}
		fields[item.UUKey] = item
	}
	tables, _ := self.getTables(ctx, where2)
	inputs, _ := self.getInputs(ctx, where2)
	support.LogDebugf(logID, "Load %s Success", model)
	if keys := maputil.Keys(inputs); len(keys) > 0 {
		support.LogDebugf(logID, "Load %s inputs: %s", model, keys)
	}
	if keys := maputil.Keys(tables); len(keys) > 0 {
		support.LogDebugf(logID, "Load %s tables: %s", model, keys)
	}

	value := &source.Value{
		Model: *detail, Fields: fields, Groups: groups,
		Tables: tables, Inputs: inputs, Clicks: clicks,
	}
	return loader.ResolveExtends(ctx, self, model, value)
}

func (self *Loader) getModel(where map[string]any) (*source.Model, error) {
	result := &source.Model{}
	query := self.client.Table(self.tables.Model)
	count := query.Where(where).First(result).RowsAffected
	if count == 0 {
		return nil, query.Error
	}
	slice.ForEach(result.Groups, func(idx int, item source.Group) {
		result.Groups[idx].SeqNo = support.Or(item.SeqNo, uint16(idx+1))
	})
	slice.ForEach(result.Fields, func(idx int, item source.Field) {
		result.Fields[idx].UUKey = item.GetKey()
		result.Fields[idx].SeqNo = support.Or(
			item.SeqNo, uint16(idx+1),
		)
	})
	result.Driver = strings.ToUpper(result.Driver)
	return result, nil
}

func (self *Loader) getTables(ctx context.Context, where map[string]any) (map[string]source.Table, error) {
	logID := loader.GetLogID(ctx)
	result := []map[string]any{}
	query := self.client.Table(self.tables.Table)
	cursor := query.Where(where).Scan(&result)
	if cursor.Error != nil {
		support.LogError(logID, "Get Tables Error", cursor.Error)
		return nil, cursor.Error
	}
	if cursor.RowsAffected == 0 {
		preset := source.Table{
			Title: "default",
			UUKey: "default",
		}
		return map[string]source.Table{
			"default": preset,
		}, nil
	}
	tables := map[string]source.Table{}
	slice.ForEach(result, func(idx int, item map[string]any) {
		table := source.Table{}
		if uukey, ok := item["uukey"]; ok && uukey != "" {
			table.UUKey = uukey.(string)
		}
		if title, ok := item["title"]; ok && title != "" {
			table.Title = title.(string)
		}
		if query, ok := item["query"]; ok && query != nil && query != "" {
			json.Unmarshal([]byte(query.(string)), &table.Query)
		}
		if fields, ok := item["fields"]; ok && fields != nil && fields != "" {
			json.Unmarshal([]byte(fields.(string)), &table.Fields)
		}
		if clicks, ok := item["clicks"]; ok && clicks != nil && clicks != "" {
			json.Unmarshal([]byte(clicks.(string)), &table.Clicks)
		}
		if hidden, ok := item["hidden"]; ok && hidden != nil && hidden != "" {
			json.Unmarshal([]byte(hidden.(string)), &table.Hidden)
		}
		if extra, ok := item["extra"]; ok && extra != nil && extra != "" {
			json.Unmarshal([]byte(extra.(string)), &table.Extra)
		}
		if rename, ok := item["rename"]; ok && rename != nil && rename != "" {
			json.Unmarshal([]byte(rename.(string)), &table.Rename)
		}
		if replace, ok := item["replace"]; ok && replace != nil && replace != "" {
			json.Unmarshal([]byte(replace.(string)), &table.Replace)
		}
		tables[table.UUKey] = table
	})
	return tables, nil
}

func (self *Loader) getInputs(ctx context.Context, where map[string]any) (map[string]source.Input, error) {
	logID := loader.GetLogID(ctx)
	result := []map[string]any{}
	query := self.client.Table(self.tables.Input)
	cursor := query.Where(where).Scan(&result)
	if cursor.Error != nil {
		support.LogError(logID, "Get Inputs Error", cursor.Error)
		return nil, cursor.Error
	}
	if cursor.RowsAffected == 0 {
		support.LogError(logID, "Get Inputs Error", cursor.Error)
		return map[string]source.Input{}, cursor.Error
	}
	inputs := map[string]source.Input{}
	slice.ForEach(result, func(idx int, item map[string]any) {
		input := source.Input{}
		if uukey, ok := item["uukey"]; ok && uukey != "" {
			input.UUKey = uukey.(string)
		}
		if title, ok := item["title"]; ok && title != "" {
			input.Title = title.(string)
		}
		if extra, ok := item["extra"]; ok && extra != nil && extra != "" {
			json.Unmarshal([]byte(extra.(string)), &input.Extra)
		}
		if fields, ok := item["fields"]; ok && fields != nil && fields != "" {
			json.Unmarshal([]byte(fields.(string)), &input.Fields)
		}
		if groups, ok := item["groups"]; ok && groups != nil && groups != "" {
			json.Unmarshal([]byte(groups.(string)), &input.Groups)
		}
		if xrules, ok := item["xrules"]; ok && xrules != nil && xrules != "" {
			json.Unmarshal([]byte(xrules.(string)), &input.XRules)
		}
		if clicks, ok := item["clicks"]; ok && clicks != nil && clicks != "" {
			json.Unmarshal([]byte(clicks.(string)), &input.Clicks)
		}
		if hidden, ok := item["hidden"]; ok && hidden != nil && hidden != "" {
			json.Unmarshal([]byte(hidden.(string)), &input.Hidden)
		}
		if preset, ok := item["preset"]; ok && preset != nil && preset != "" {
			json.Unmarshal([]byte(preset.(string)), &input.Preset)
		}
		if rename, ok := item["rename"]; ok && rename != nil && rename != "" {
			json.Unmarshal([]byte(rename.(string)), &input.Rename)
		}
		if replace, ok := item["replace"]; ok && replace != nil && replace != "" {
			json.Unmarshal([]byte(replace.(string)), &input.Replace)
		}
		inputs[input.UUKey] = input
	})
	return inputs, nil
}
