package mongo

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/OptLTD/library/search/consts"
	"github.com/OptLTD/library/search/loader"
	"github.com/OptLTD/library/search/source"
	"github.com/OptLTD/library/search/support"

	"github.com/duke-git/lancet/v2/maputil"
	"github.com/duke-git/lancet/v2/slice"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewLoader(db *mongo.Database, tables *loader.SchemaTableName) loader.ILoader {
	return &Loader{client: db, tables: tables}
}

type Loader struct {
	client *mongo.Database
	tables *loader.SchemaTableName
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
	tModel := time.Now()
	detail, err := self.getModel(ctx, where1)
	support.LogWatchCostMs(tModel, logID, "Engine Stage",
		"stage", "schema_loader_biz_model", "model", model,
	)
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
	// reset default group, seqno to 0
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

	tTbl := time.Now()
	tables, _ := self.getTables(ctx, where2)
	support.LogWatchCostMs(tTbl, logID, "Engine Stage",
		"stage", "schema_loader_biz_tables", "model", model,
	)
	tInp := time.Now()
	inputs, _ := self.getInputs(ctx, where2)
	support.LogWatchCostMs(tInp, logID, "Engine Stage",
		"stage", "schema_loader_biz_inputs", "model", model,
	)

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

func (self *Loader) toBson(where map[string]any) bson.D {
	filter := bson.D{}
	for key, val := range where {
		switch support.GetType(val) {
		case "array":
			child := bson.D{{Key: "$in", Value: val}}
			where := bson.E{Key: key, Value: child}
			filter = append(filter, where)
		case "bool":
			value := []any{val, support.If(val, 1, 0)}
			child := bson.D{{Key: "$in", Value: value}}
			where := bson.E{Key: key, Value: child}
			filter = append(filter, where)
		default:
			filter = append(filter, bson.E{Key: key, Value: val})
		}
	}
	return filter
}

func (self *Loader) getModel(ctx context.Context, where map[string]any) (*source.Model, error) {
	logID := loader.GetLogID(ctx)
	result := &source.Model{}
	collect := self.client.Collection(self.tables.Model)
	find := collect.FindOne(ctx, self.toBson(where))
	if err := find.Decode(&result); err != nil {
		support.LogError(logID, "Get Model Error", err)
		return nil, err
	}

	// reset index
	slice.ForEach(result.Groups, func(idx int, item source.Group) {
		result.Groups[idx].SeqNo = support.Or(item.SeqNo, uint16(idx+1))
	})
	slice.ForEach(result.Fields, func(idx int, item source.Field) {
		result.Fields[idx].UUKey = item.GetKey()
		result.Fields[idx].SeqNo = support.Or(
			item.SeqNo, uint16(idx+1),
		)
	})
	// rename driver
	result.Driver = strings.ToUpper(result.Driver)
	return result, nil
}

func (self *Loader) getTables(ctx context.Context, where map[string]any) (map[string]source.Table, error) {
	logID := loader.GetLogID(ctx)
	result, collect := []source.Table{}, self.client.Collection(self.tables.Table)
	if cursor, err := collect.Find(ctx, self.toBson(where)); err != nil {
		support.LogError(logID, "Get Tables Error", err)
		return nil, err
	} else if err := cursor.All(context.TODO(), &result); err != nil {
		support.LogError(logID, "Get Tables Error", err)
		return nil, err
	} else if len(result) == 0 {
		result = append(result, source.Table{
			Title: "default", UUKey: "default",
		})
	}

	final := map[string]source.Table{}
	for _, table := range result {
		if len(table.Query) > 0 {
			table.Query = support.NormalizeQueryObject(table.Query)
		}
		if len(table.Extra) > 0 {
			var extra = map[string]any{}
			if data, err := json.Marshal(table.Extra); err == nil {
				json.Unmarshal(data, &extra)
			}
			if len(extra) > 0 {
				table.Extra = extra
			}
		}
		final[table.UUKey] = table
	}

	// key by uukey
	return final, nil
}

func (self *Loader) getInputs(ctx context.Context, where map[string]any) (map[string]source.Input, error) {
	logID := loader.GetLogID(ctx)
	collect := self.client.Collection(self.tables.Input)
	result, collect := []source.Input{}, self.client.Collection(self.tables.Input)
	if cursor, err := collect.Find(ctx, self.toBson(where)); err != nil {
		support.LogError(logID, "Get Inputs Error", err)
		return nil, err
	} else if err := cursor.All(context.TODO(), &result); err != nil {
		support.LogError(logID, "Get Inputs Error", err)
		return nil, err
	} else if len(result) == 0 {
		result = append(result, source.Input{
			Title: "default", UUKey: "default",
		})
	}
	return slice.KeyBy(result, func(item source.Input) string {
		return item.UUKey
	}), nil
}
