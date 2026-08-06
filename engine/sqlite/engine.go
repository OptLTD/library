package sqlite

import (
	"fmt"
	"log"
	"strings"

	"github.com/OptLTD/library/search/consts"
	"github.com/OptLTD/library/search/storage"
	"github.com/OptLTD/library/search/respond"
	"github.com/OptLTD/library/search/schema"
	"github.com/OptLTD/library/search/support"

	"github.com/duke-git/lancet/v2/maputil"
	"github.com/duke-git/lancet/v2/slice"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Engine struct {
	client  *gorm.DB
	handles []storage.ICallable
}

func NewEngine(db *gorm.DB) storage.IEngine {
	session := &gorm.Session{QueryFields: true}
	return &Engine{
		client: db.Session(session),
	}
}

func (self *Engine) Using(handle storage.ICallable) storage.IEngine {
	if self.handles == nil {
		self.handles = []storage.ICallable{}
	}
	self.handles = append(self.handles, handle)
	return self
}

func (self *Engine) First(skma *schema.Input, record *respond.Record) error {
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
		record.Exists = true
		record.Current = record.ToFlatten(skma, record.Storage)
	}
	return nil
}

func (self *Engine) Store(skma *schema.Input, record *respond.Record) error {
	table := skma.Model.Source
	for _, handle := range self.handles {
		if err := handle.BeforeUpsert(skma, record); err != nil {
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
	if record.OpType == "" || record.OpType == "INSERT" {
		merged := maputil.Merge(record.Default, upsert)
		query.Create(merged)
	} else {
		query.Updates(upsert)
	}
	for _, handle := range self.handles {
		if err := handle.HandleUpsert(skma, record); err != nil {
			break
		}
	}
	return nil
}

func (self *Engine) Select(skma *schema.Input, records []*respond.Record) error {
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
			record.OpType = "UPDATE"
			record.Storage = record.Decode(skma, value)
		} else {
			record.OpType = "INSERT"
			record.Storage = record.Default
		}
		record.Objects = record.ToObjects(skma, record.Request)
		record.Changes = record.ToChanges(skma, record.Request)
		record.Changed = record.ToFlatten(skma, record.Objects)
		record.Prepare = record.ToPrepare(skma, record.Objects)
	}
	return nil
}

func (self *Engine) Upsert(skma *schema.Input, records []*respond.Record) error {
	table := skma.Model.Source
	values := []map[string]any{}
	size := len(records)

	for _, handle := range self.handles {
		for i := 0; i < size; i++ {
			if err := handle.BeforeUpsert(skma, records[i]); err != nil {
				return err
			}
		}
	}

	fields := []string{}
	for i := 0; i < size; i++ {
		record := records[i]
		value := record.Encode(skma, record.Prepare)
		values = append(values, value)
		fields = append(fields, maputil.Keys(record.Changes)...)
	}
	fields = slice.Unique(fields)
	query := self.client.Table(table).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "uukey"}},
			DoUpdates: clause.AssignmentColumns(fields),
		}).
		CreateInBatches(values, 100)
	for _, handle := range self.handles {
		for i := 0; i < size; i++ {
			if err := handle.HandleUpsert(skma, records[i]); err != nil {
				break
			}
		}
	}
	return query.Error
}

func (self *Engine) Update(skma *schema.Input, data map[string]any) error {
	if len(skma.Scope) == 0 {
		return fmt.Errorf("update scope cannot be empty")
	}
	id, has := skma.Scope[consts.FIELD_CORP_ID]
	if !has || support.Bool(id) == false {
		return fmt.Errorf("update scope must contain corp_id")
	}

	table := skma.Model.Source
	mergedQuery, _ := schema.BuildQuery(skma.Scope)
	parsedQuery := self.buildQuery(consts.LOGIC_SUBAND, &mergedQuery)

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

func (self *Engine) Search(skma *schema.Table) (*respond.Result, error) {
	skma = self.resetTable(skma)
	request := skma.Request
	merged := skma.BuildQuery()
	bindQueryIndexes(skma, &merged)
	queries := self.buildQuery(consts.LOGIC_SUBAND, &merged)
	offset := int((request.Page - 1) * request.Size)
	values, count := []map[string]any{}, int64(0)
	query := self.client.Table(skma.Model.Search).Clauses(queries...)
	if request.Order != nil && request.Order.Field != "" && skma.GetField(request.Order.Field) != nil {
		expr := fieldExpr(skma, request.Order.Field)
		dir := "ASC"
		if strings.EqualFold(request.Order.Order, "desc") {
			dir = "DESC"
		}
		query = query.Order(fmt.Sprintf("%s %s", expr, dir))
	}
	query.Count(&count).Limit(int(request.Size)).Offset(offset).Find(&values)
	record := respond.Record{}
	values = slice.Map(values, func(idx int, item map[string]any) map[string]any {
		decoded := record.Decode(skma, item)
		return record.ToFlatten(skma, decoded)
	})
	result := &respond.Result{
		Page: request.Page, Count: uint64(count),
		Size: request.Size, Values: values,
	}
	for _, h := range self.handles {
		if err := h.SearchResult(skma, result); err != nil {
			log.Println("handle result err:", err)
		}
	}
	return result, nil
}


func (self *Engine) buildQuery(logic string, queries *[]schema.Query) []clause.Expression {
	result := []clause.Expression{}
	for _, query := range *queries {
		col := querySQLExpr(query)
		switch strings.ToUpper(query.Logic) {
		case consts.LOGIC_EQUALSTO:
			result = append(result, eqClause(col, query.Value))
		case consts.LOGIC_NOTEQUAL:
			result = append(result, neClause(col, query.Value))
		case consts.LOGIC_VAL_NULL:
			result = append(result, nilClause(col))
		case consts.LOGIC_NOT_NULL:
			result = append(result, nnlClause(col))
		case consts.LOGIC_STR_LIKE:
			result = append(result, likeClause(col, query.Value))
		case consts.LOGIC_INCLUDES:
			result = append(result, inClause(col, query.Value.([]any)))
		case consts.LOGIC_CONTAINS:
			subsql := fmt.Sprintf(`instr(',' || %s || ',', ',' || ? || ',') > 0`, quoteCol(col))
			result = append(result, clause.Expr{SQL: subsql, Vars: []any{query.Value}})
		case consts.LOGIC_LESTHAN:
			result = append(result, cmpClause(col, "<", query.Value))
		case consts.LOGIC_GREATER:
			result = append(result, cmpClause(col, ">", query.Value))
		case consts.LOGIC_LESS_EQ:
			result = append(result, cmpClause(col, "<=", query.Value))
		case consts.LOGIC_GRAT_EQ:
			result = append(result, cmpClause(col, ">=", query.Value))
		case consts.LOGIC_BETWEEN:
			subsql := fmt.Sprintf(`%s between ? and ?`, quoteCol(col))
			result = append(result, clause.Expr{SQL: subsql, Vars: query.Value.([]any)})
		case consts.LOGIC_SUBRAW:
			result = append(result, clause.Expr{SQL: query.Value.(string), Vars: []any{}})
		case consts.LOGIC_SUBOR:
			result = append(result, clause.Or(self.buildQuery(consts.LOGIC_SUBOR, query.Items)...))
		case consts.LOGIC_SUBAND:
			result = append(result, clause.And(self.buildQuery(consts.LOGIC_SUBAND, query.Items)...))
		case consts.LOGIC_SUBNOT:
			result = append(result, clause.Not(self.buildQuery(consts.LOGIC_SUBNOT, query.Items)...))
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
