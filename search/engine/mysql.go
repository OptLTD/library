package engine

import (
	"fmt"
	"log"
	"github.com/OptLTD/library/search/consts"
	"github.com/OptLTD/library/search/respond"
	"github.com/OptLTD/library/search/schema"
	"github.com/OptLTD/library/search/source"
	"github.com/OptLTD/library/search/support"
	"strings"

	"github.com/duke-git/lancet/v2/maputil"
	"github.com/duke-git/lancet/v2/slice"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MySQLEngine struct {
	client  *gorm.DB
	handles []ICallable
}

func (self *MySQLEngine) Using(handle ICallable) IEngine {
	if self.handles == nil {
		self.handles = []ICallable{}
	}

	self.handles = append(self.handles, handle)
	return self
}

func (self *MySQLEngine) First(skma *schema.Input, record *respond.Record) error {
	table := skma.Model.Source
	where := map[string]any{
		"uukey": record.UUKey,
	}
	value := []map[string]any{}

	where = maputil.Merge(skma.Scope, where)
	mergedQuery, _ := schema.BuildQuery(where)
	parsedQuery := self.buildQuery(consts.LOGIC_SUBAND, &mergedQuery)
	query := self.client.Table(table).Clauses(parsedQuery...).Find(&value)
	if query.RowsAffected >= 2 {
		return support.RecordsFoundWhenFirst
	}
	if query.RowsAffected != 0 {
		record.Storage = record.Decode(skma, value[0])
		record.Changed = record.ToFlatten(skma, record.Storage)
	}
	record.Objects = record.ToObjects(skma, record.Request)
	record.Changes = record.ToChanges(skma, record.Request)
	record.Prepare = record.ToPrepare(skma, record.Objects)
	record.Changed = record.ToFlatten(skma, record.Objects)
	return nil
}

func (self *MySQLEngine) Store(skma *schema.Input, record *respond.Record) error {
	table := skma.Model.Source
	// 处理回掉
	for _, handle := range self.handles {
		err := handle.BeforeUpsert(skma, record)
		if err != nil {
			return err
		}
	}
	upsert := record.Encode(skma, record.Prepare)
	where := map[string]any{
		"model": record.Model,
		"uukey": record.UUKey,
	}
	where = maputil.Merge(skma.Scope, where)
	dftQuery, _ := schema.BuildQuery(where)
	parsedQuery := self.buildQuery(consts.LOGIC_SUBAND, &dftQuery)
	query := self.client.Table(table).Clauses(parsedQuery...)
	if record.Event == "" || record.Event == "INSERT" {
		merged := maputil.Merge(record.Default, upsert)
		query.Create(merged)
	} else {
		query.Updates(upsert)
	}
	// 处理回掉
	for _, handle := range self.handles {
		err := handle.HandleUpsert(skma, record)
		if err != nil {
			break
		}
	}
	return nil
}

func (self *MySQLEngine) Select(skma *schema.Input, records []*respond.Record) error {
	table := skma.Model.Source
	uukeys, size := []any{}, len(records)
	for i := 0; i < size; i++ {
		uukeys = append(uukeys, records[i].UUKey)
	}
	where := map[string]any{
		"uukey": uukeys,
	}
	values := []map[string]any{}
	if skma.Scope != nil {
		where = maputil.Merge(skma.Scope, where)
	}
	usingQuery, _ := schema.BuildQuery(where)
	parsedQuery := self.buildQuery(consts.LOGIC_SUBAND, &usingQuery)
	query := self.client.Table(table).Clauses(parsedQuery...).Find(&values)
	if query.Error != nil {
		return query.Error
	}
	result := slice.KeyBy(values, func(item map[string]any) string {
		return item["uukey"].(string)
	})
	for i := 0; i < size; i++ {
		record := records[i]
		if value, ok := result[record.UUKey]; ok {
			record.Event = "UPDATE"
			record.Storage = record.Decode(skma, value)
		} else {
			record.Event = "INSERT"
			record.Storage = record.Default
		}
		record.Objects = record.ToObjects(skma, record.Request)
		record.Changes = record.ToChanges(skma, record.Request)
		record.Changed = record.ToFlatten(skma, record.Objects)
		record.Prepare = record.ToPrepare(skma, record.Objects)
	}
	return nil
}

func (self *MySQLEngine) Upsert(skma *schema.Input, records []*respond.Record) error {
	table := skma.Model.Source
	values := []map[string]any{}
	size := len(records)

	// 处理回掉
	for _, handle := range self.handles {
		for i := 0; i < size; i++ {
			err := handle.BeforeUpsert(skma, records[i])
			if err != nil {
				return err
			}
		}
	}

	fields := []string{} // "utime"
	for i := 0; i < size; i++ {
		record := records[i]
		value := record.Encode(skma, record.Prepare)
		values = append(values, value)

		change := record.Changes
		fields = append(fields, maputil.Keys(change)...)
	}
	fields = slice.Unique(fields)
	query := self.client.Table(table).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "uukey"}},
			DoUpdates: clause.AssignmentColumns(fields),
		}).
		CreateInBatches(values, 100)
		// 处理回掉
	for _, handle := range self.handles {
		for i := 0; i < size; i++ {
			err := handle.HandleUpsert(skma, records[i])
			if err != nil {
				break
			}
		}
	}
	return query.Error
}

func (self *MySQLEngine) Update(skma *schema.Input, data map[string]any) error {
	// 检查 Scope 是否为空
	if len(skma.Scope) == 0 {
		return fmt.Errorf("update scope cannot be empty")
	}

	// 检查是否包含有效的 corp_id
	id, has := skma.Scope[consts.FIELD_CORP_ID]
	if !has || support.Bool(id) == false {
		return fmt.Errorf("update scope must contain corp_id")
	}

	table := skma.Model.Source
	mergedQuery, _ := schema.BuildQuery(skma.Scope)
	parsedQuery := self.buildQuery(consts.LOGIC_SUBAND, &mergedQuery)

	// 更新前先 count，超过 1000 条则阻止
	var count int64
	countQuery := self.client.Table(table).Clauses(parsedQuery...)
	if err := countQuery.Count(&count).Error; err != nil {
		return err
	}
	if count > 1000 {
		return fmt.Errorf("update would affect %d records, exceeds limit of 1000", count)
	}

	query := self.client.Table(table).Clauses(parsedQuery...).Updates(data)
	return query.Error
}

func (self *MySQLEngine) Search(skma *schema.Table) (*respond.Result, error) {
	request := skma.Request
	merged := skma.BuildQuery()
	queries := self.buildQuery(consts.LOGIC_SUBAND, &merged)
	offset := int((request.Page - 1) * request.Size)
	values, count := []map[string]any{}, int64(0)
	query := self.client.Table(skma.Model.Search).Clauses(queries...)
	query.Count(&count).Limit(int(request.Size)).Offset(offset).Find(&values)
	record := respond.Record{}
	values = slice.Map(values, func(idx int, item map[string]any) map[string]any {
		return record.Decode(skma, item)
	})
	result := &respond.Result{
		Page: request.Page, Count: uint64(count),
		Size: request.Size, Values: values,
	}

	// 处理回掉
	for _, h := range self.handles {
		err := h.SearchResult(skma, result)
		if err != nil {
			log.Println("handle result err:", err)
		}
	}
	return result, nil
}

func (self *MySQLEngine) Digest(skma *schema.Digest) (*respond.Result, error) {
	return self.Search(skma.Table)
}

func (self *MySQLEngine) buildField(fields *[]source.Field) string {
	if fields == nil || len(*fields) == 0 {
		return "*"
	}
	slices := []string{}
	for _, field := range *fields {
		slices = append(slices, field.Field)
	}
	return strings.Join(slices, ",")
}

func (self *MySQLEngine) buildQuery(logic string, queries *[]schema.Query) []clause.Expression {
	result := []clause.Expression{}
	for _, query := range *queries {
		if strings.HasPrefix(query.Field, "basic.") {
			query.Field = strings.Replace(query.Field, "basic.", "", 1)
		}
		switch strings.ToUpper(query.Logic) {
		case consts.LOGIC_EQUALSTO:
			result = append(result, clause.Eq{Column: query.Field, Value: query.Value})
		case consts.LOGIC_STR_LIKE:
			result = append(result, clause.Like{Column: query.Field, Value: query.Value})
		case consts.LOGIC_INCLUDES:
			result = append(result, clause.IN{Column: query.Field, Values: query.Value.([]any)})
		case consts.LOGIC_CONTAINS:
			subsql := fmt.Sprintf("find_in_set(`%s`, ?)", query.Field)
			result = append(result, clause.Expr{SQL: subsql, Vars: []any{query.Value}})
		case consts.LOGIC_LESTHAN:
			result = append(result, clause.Lt{Column: query.Field, Value: query.Value})
		case consts.LOGIC_GREATER:
			result = append(result, clause.Gt{Column: query.Field, Value: query.Value})
		case consts.LOGIC_LESS_EQ:
			result = append(result, clause.Lte{Column: query.Field, Value: query.Value})
		case consts.LOGIC_GRAT_EQ:
			result = append(result, clause.Gte{Column: query.Field, Value: query.Value})
		case consts.LOGIC_BETWEEN:
			subsql := fmt.Sprintf("`%s` between ? and ?", query.Field)
			result = append(result, clause.Expr{SQL: subsql, Vars: query.Value.([]any)})
		case consts.LOGIC_SUBRAW:
			result = append(result, clause.Expr{SQL: query.Value.(string), Vars: []any{}})
		case consts.LOGIC_SUBOR:
			result = append(result, clause.Or(self.buildQuery(consts.LOGIC_SUBOR, query.Items)...))
		case consts.LOGIC_SUBAND:
			result = append(result, clause.And(self.buildQuery(consts.LOGIC_SUBAND, query.Items)...))
		case consts.LOGIC_SUBNOT:
			result = append(result, clause.Not(self.buildQuery(consts.LOGIC_SUBNOT, query.Items)...))
		default:
			// 异常逻辑
		}
	}
	if len(result) == 0 {
		return []clause.Expression{}
	}
	switch logic {
	case consts.LOGIC_SUBOR:
		return []clause.Expression{clause.Or(result...)}
	case consts.LOGIC_SUBAND:
		return []clause.Expression{clause.And(result...)}
	case consts.LOGIC_SUBNOT:
		return []clause.Expression{clause.Not(result...)}
	}
	return []clause.Expression{}
}
