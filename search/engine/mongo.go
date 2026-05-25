package engine

import (
	"context"
	"errors"
	"fmt"
	"math"
	"search/consts"
	"search/request"
	"search/respond"
	"search/schema"
	"search/source"
	"search/support"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/maputil"
	"github.com/duke-git/lancet/v2/slice"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoEngine struct {
	debug bool

	client  *mongo.Database
	handles []ICallable
}

// aggrStage 封装 MongoDB 聚合管道的各个阶段配置
type aggrStage struct {
	Project bson.M // $project 阶段的字段映射
	Group   bson.M // $group 阶段的聚合操作
	Unique  bson.M // $addFields 阶段的去重计数计算（用于 VALUE_UNQ）
}

func (self *MongoEngine) Using(handle ICallable) IEngine {
	if self.handles == nil {
		self.handles = []ICallable{}
	}

	self.handles = append(self.handles, handle)
	return self
}

func (self *MongoEngine) First(skma *schema.Input, record *respond.Record) error {
	logID := skma.Request.LogID
	table := skma.Model.Source
	where := map[string]any{
		consts.FIELD_UUKEY: record.UUKey,
	}
	where = maputil.Merge(skma.Scope, where)
	preset, _ := schema.BuildQuery(where)
	tblSkm := self.inputToTable(skma)
	parsed := self.buildQuery(consts.LOGIC_SUBAND, &preset, tblSkm)
	t0 := time.Now()
	result := self.client.Collection(table).FindOne(context.TODO(), parsed)
	if err := result.Decode(&record.Storage); err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return err
		}
	}
	if record.Storage != nil {
		record.Exists = true
	}
	cost := time.Since(t0).Milliseconds()
	support.LogKVByCostMs(logID, "mongo query first", cost,
		"table", table, "query", preset, "filter", parsed,
		"exists", record.Exists, "uukey", record.UUKey,
	)
	record.Current = record.ToCurrent(skma, record.Storage)
	record.Objects = record.ToObjects(skma, record.Current)
	return nil
}

func (self *MongoEngine) Store(skma *schema.Input, record *respond.Record) error {
	table := skma.Model.Source
	// 处理回掉
	for _, handle := range self.handles {
		err := handle.BeforeUpsert(skma, record)
		if err != nil {
			return err
		}
	}
	where := map[string]any{
		consts.FIELD_UUKEY: record.UUKey,
	}
	where = maputil.Merge(skma.Scope, where)
	preset, _ := schema.BuildQuery(where)
	tblSkm := self.inputToTable(skma)
	parsed := self.buildQuery(consts.LOGIC_SUBAND, &preset, tblSkm)

	var logID = ""
	upsert := record.Format(skma, record.Prepare)
	if skma.Request != nil {
		logID = skma.Request.LogID
	}
	var err error
	collect := self.client.Collection(table)
	t0 := time.Now()
	if record.Event == "" || record.Event == "INSERT" {
		merged := maputil.Merge(record.Default, upsert)
		_, err = collect.InsertOne(
			context.TODO(), merged,
		)
	} else {
		_, err = collect.UpdateOne(
			context.TODO(), parsed,
			bson.M{"$set": upsert},
		)
	}
	// handle error
	if err != nil {
		return err
	}
	cost := time.Since(t0).Milliseconds()
	support.LogKVByCostMs(logID, "mongo store record", cost,
		"table", table, "query", preset, "filter", parsed,
		"event", record.Event, "uukey", record.UUKey,
	)
	// 处理回掉
	for _, handle := range self.handles {
		err := handle.HandleUpsert(skma, record)
		if err != nil {
			break
		}
	}
	return nil
}

func (self *MongoEngine) Select(skma *schema.Input, records []*respond.Record) error {
	logID := skma.Request.LogID
	table := skma.Model.Source
	uukeys, size := []any{}, len(records)
	for i := 0; i < size; i++ {
		uukeys = append(uukeys, records[i].UUKey)
	}

	where := bson.M{
		consts.FIELD_UUKEY: bson.M{"$in": uukeys},
	}
	where = maputil.Merge(skma.Scope, where)
	preset, _ := schema.BuildQuery(where)

	tbl := self.inputToTable(skma)
	values, collect := []map[string]any{}, self.client.Collection(table)
	parsed := self.buildQuery(consts.LOGIC_SUBAND, &preset, tbl)
	t0 := time.Now()
	if cursor, err := collect.Find(context.TODO(), parsed); err != nil {
		return err
	} else if err = cursor.All(context.TODO(), &values); err != nil {
		return err
	}
	cost := time.Since(t0).Milliseconds()
	support.LogKVByCostMs(logID, "mongo query select", cost,
		"table", table, "query", preset, "filter", parsed,
		"batch_size", size, "rows", len(values),
	)
	mapped := slice.KeyBy(values, func(item map[string]any) string {
		return item["uukey"].(string)
	})
	for i := 0; i < size; i++ {
		if value, ok := mapped[records[i].UUKey]; ok {
			records[i].Storage = value
			records[i].Exists = true
		} else {
			records[i].Exists = false
		}
	}
	return nil
}

func (self *MongoEngine) Upsert(skma *schema.Input, records []*respond.Record) error {
	table, size := skma.Model.Source, len(records)
	// 处理回掉
	for _, handle := range self.handles {
		for i := 0; i < size; i++ {
			err := handle.BeforeUpsert(skma, records[i])
			if err != nil {
				return err
			}
		}
	}

	// insert, update := []any{}, []any{}
	collect := self.client.Collection(table)
	var upsert []mongo.WriteModel
	for i := 0; i < size; i++ {
		record := records[i]
		combine := maputil.OmitByKeys(record.Prepare, maputil.Keys(record.Default))
		values, filter := record.Format(skma, combine), record.GetUnique(skma)

		preset := maputil.OmitByKeys(record.Default, skma.Unique)
		update := bson.M{"$set": values, "$setOnInsert": preset}
		ops := mongo.NewUpdateOneModel().SetFilter(filter).
			SetUpdate(update).SetUpsert(true)
		upsert = append(upsert, ops)
	}
	t0, logID := time.Now(), skma.Request.LogID
	res, err := collect.BulkWrite(context.Background(), upsert)
	if err != nil {
		support.LogError(logID, res.InsertedCount, res.UpsertedIDs)
		return err
	}
	cost := time.Since(t0).Milliseconds()
	support.LogKVByCostMs(logID, "mongo upsert bulk write", cost,
		"table", table, "writes", len(upsert), "matched", res.MatchedCount,
		"modified", res.ModifiedCount, "upserted", res.UpsertedCount,
	)
	// 处理回掉
	for _, handle := range self.handles {
		for i := 0; i < size; i++ {
			err := handle.HandleUpsert(skma, records[i])
			if err != nil {
				break
			}
		}
	}
	return nil
}

func (self *MongoEngine) Update(skma *schema.Input, data map[string]any) error {
	// 检查 Scope 是否为空
	if len(skma.Scope) == 0 {
		return fmt.Errorf("update scope cannot be empty")
	}

	// 检查是否包含有效的 corp_id
	id, has := skma.Scope[consts.FIELD_CORP_ID]
	if !has || support.Bool(id) == false {
		return fmt.Errorf("update scope must contain corp_id")
	}

	preset, _ := schema.BuildQuery(skma.Scope)
	table, tblSkm := skma.Model.Source, self.inputToTable(skma)
	parsed := self.buildQuery(consts.LOGIC_SUBAND, &preset, tblSkm)

	t0, logid := time.Now(), skma.Request.LogID // 记录开始时间
	// 更新前先 count，超过 1000 条则阻止
	collect := self.client.Collection(table)
	count, err := collect.CountDocuments(context.TODO(), parsed)
	if err != nil {
		support.LogErrorKV(logid, "mongo update count error",
			"table", table, "query", preset, "filter", parsed, "err", err,
		)
		return err
	}
	if count > 1000 {
		cost := time.Since(t0).Milliseconds()
		support.LogKVByCostMs(logid, "mongo update", cost,
			"table", table, "query", preset, "filter", parsed,
			"count", count, "blocked", true,
		)
		return fmt.Errorf("update would affect %d records, exceeds limit of 1000", count)
	}

	upRes, err := collect.UpdateMany(
		context.TODO(), parsed,
		bson.M{"$set": data},
	)
	cost := time.Since(t0).Milliseconds()
	var mod, mat int64
	if err == nil {
		mod, mat = upRes.ModifiedCount, upRes.MatchedCount
	}
	support.LogKVByCostMs(logid, "mongo update", cost,
		"table", table, "query", preset, "filter", parsed,
		"count", count, "modified", mod, "matched", mat,
	)
	return err
}

func (self *MongoEngine) Search(skma *schema.Table) (*respond.Result, error) {
	skma = self.resetTable(skma)
	request, table := skma.Request, skma.Model.Source
	logID := support.Or(request.LogID, "unknown")
	query, logic := skma.BuildQuery(), consts.LOGIC_SUBAND
	parsed := self.buildQuery(logic, &query, skma)

	record, t0 := respond.Record{}, time.Now()
	collect := self.client.Collection(table)

	values, count := []map[string]any{}, int64(0)
	offset := int64((request.Page - 1) * request.Size)
	result := &respond.Result{Page: request.Page, Size: request.Size}
	count, err := collect.CountDocuments(context.TODO(), parsed)
	if err != nil || count == 0 {
		cost := time.Since(t0).Milliseconds()
		support.LogKVByCostMs(logID, "mongo query search", cost,
			"table", table, "query", query, "filter", parsed,
			"count", count, "err", err,
		)
		if err != nil {
			support.LogErrorKV(logID, "mongo query error",
				"stage", "count", "table", table,
				"query", query, "filter", parsed,
				"count", count, "err", err,
			)
		}
		return result, nil
	}
	size := int64(request.Size)
	result.Count = uint64(count)
	option := &options.FindOptions{
		Limit: &size, Skip: &offset,
	}
	if order := request.Order; order != nil {
		var sort int32 = 1
		switch strings.ToUpper(order.Order) {
		case "ASC", "ASCENDING":
			sort = 1
		case "DESC", "DESCENDING":
			sort = -1
		default:
			sort = 1
		}
		// 如果排序字段不存在，则不排序
		field := skma.GetField(order.Field)
		if field != nil && field.Index != "" {
			option.SetSort(bson.M{field.Index: sort})
		}
	}
	if cursor, err := collect.Find(context.TODO(), parsed, option); err != nil {
		return result, err
	} else if err = cursor.All(context.TODO(), &values); err != nil {
		return result, err
	}
	cost := time.Since(t0).Milliseconds()
	support.LogKVByCostMs(logID, "mongo query search", cost,
		"table", table, "filter", parsed, "query", query,
		"count", count, "size", size, "rows", len(values),
	)

	result.Values = slice.Map(values, func(idx int, item map[string]any) map[string]any {
		return record.ToCurrent(skma, item)
	})

	// 处理回掉
	for _, h := range self.handles {
		err := h.SearchResult(skma, result)
		if err != nil {
			support.LogError(request.LogID, "handle result err:", err)
		}
	}
	return result, nil
}

func (self *MongoEngine) Digest(skma *schema.Digest) (*respond.Result, error) {
	self.resetPivot(skma)

	// 分桶聚合
	groups := self.getGroupBy(skma)
	counts := self.getAggrs(skma, "")
	if len(groups)+len(counts) == 0 {
		return nil, nil
	}

	for _, group := range groups {
		option := map[string]any{}
		field := skma.Table.GetField(group.Index)
		if field == nil || field.Index == "" {
			continue
		}
		method := consts.VALUE_TERMS
		switch field.FType {
		case consts.FTYPE_DATETIME:
			method = consts.VALUE_HIST2
			var dateFmt = "%Y-%m-%d"
			if group.Format == "month" {
				dateFmt = "%Y-%m"
			}
			option["$dateToString"] = bson.M{
				"date": "$" + field.Index,
				// date format, @see https://www.mongodb.com/docs/manual/reference/operator/aggregation/dateToString/
				"format": dateFmt,
				// timezone set @config
				"timezone": time.Local.String(),
			}
		}
		label := strings.Replace(group.Index, ".", "#", 1)
		counts = append(counts, source.CountFn{
			Index: field.Index, Label: label,
			Func: method, Option: option, Items: counts,
		})
	}

	// Build the request body.

	req, table := skma.Table.Request, skma.Table.Model.Source
	query, logic := skma.Table.BuildQuery(), consts.LOGIC_SUBAND
	match := self.buildQuery(logic, &query, skma.Table)

	stages := self.buildCountFn(counts)
	groupby := self.buildGroupBy(counts)
	parsed := bson.A{
		bson.M{"$match": match}, bson.M{"$project": stages.Project},
		bson.M{"$group": maputil.Merge(bson.M{"_id": groupby}, stages.Group)},
	}
	// 如果有 VALUE_UNQ 字段，添加额外的 addFields 阶段来计算 $size（保留其他字段）
	if len(stages.Unique) > 0 {
		parsed = append(parsed, bson.M{
			"$addFields": stages.Unique,
		})
	}
	// 聚合结果
	values, result := []bson.M{}, &respond.Result{
		Page: req.Page, Size: req.Size,
	}
	collect := self.client.Collection(table)

	logID := skma.Table.Request.LogID
	t0 := time.Now()
	if cursor, err := collect.Aggregate(context.TODO(), parsed); err != nil {
		support.LogErrorKV(logID, "mongo digest aggregate cursor error",
			"stage", "cursor", "pipeline", parsed, "err", err,
		)
		return result, err
	} else if err = cursor.All(context.TODO(), &values); err != nil {
		support.LogErrorKV(logID, "mongo digest aggregate result error",
			"stage", "result", "pipeline", parsed, "err", err,
		)
		return result, err
	}
	ms := time.Since(t0).Milliseconds()
	support.LogKVByCostMs(logID, "mongo digest aggregate query", ms,
		"table", table, "query", query, "parsed", parsed,
	)
	result.Totals = slice.Reduce(values, func(idx int, item bson.M, carry bson.M) bson.M {
		item = maputil.MapKeys(item, func(key string, val any) string {
			return strings.Replace(key, "#", ".", 1)
		})
		for key, val := range item {
			if key == "_id" {
				continue
			}
			if _, ok := carry[key]; !ok {
				carry[key] = val
				continue
			}

			if get, ok := carry[key].(int32); ok {
				if add, ok := val.(float64); ok {
					carry[key] = float64(get) + add
				} else if add, ok := val.(int64); ok {
					carry[key] = float64(get) + float64(add)
				} else {
					carry[key] = get + val.(int32)
				}
				continue
			}
			if get, ok := carry[key].(int64); ok {
				if add, ok := val.(float64); ok {
					carry[key] = float64(get) + add
				} else if add, ok := val.(int32); ok {
					carry[key] = float64(get) + float64(add)
				} else {
					carry[key] = get + val.(int64)
				}
				continue
			}
			if get, ok := carry[key].(float64); ok {
				if add, ok := val.(int32); ok {
					carry[key] = get + float64(add)
				} else if add, ok := val.(int64); ok {
					carry[key] = get + float64(add)
				} else {
					carry[key] = get + val.(float64)
				}
				continue
			}
		}
		for key, val := range carry { // 四舍五入，保留3位小数
			if get, ok := val.(float64); ok {
				carry[key] = math.Round(get*1000) / 1000
			}
		}
		return carry
	}, bson.M{})
	result.Values = slice.Map(values, func(idx int, item bson.M) map[string]any {
		transform := item["_id"].(bson.M)
		maputil.ForEach(item, func(key string, val any) {
			if !strings.Contains(key, "#") {
				return
			}
			if _, ok := transform[key]; !ok {
				transform[key] = val
			}
		})
		transform = maputil.MapKeys(transform, func(key string, val any) string {
			return strings.Replace(key, "#", ".", 1)
		})

		for key, val := range transform { // 四舍五入，保留3位小数
			if get, ok := val.(float64); ok {
				transform[key] = math.Round(get*1000) / 1000
			}
		}
		return transform
	})
	if req.Order == nil && len(groups) > 0 {
		req.Order = &request.Order{
			Field: groups[0].Index,
			Order: groups[0].SortBy,
		}
	}

	// 处理回掉
	for _, h := range self.handles {
		err := h.DigestResult(skma, result)
		if err != nil {
			support.LogErrorKV(logID, "mongo digest handle result error",
				"stage", "handle",
				"err", err,
			)
		}
	}

	// 对结果进行排序
	result.Sort(req.Order)
	if count := len(result.Values); count > 0 {
		result.Count = uint64(count)
	}
	return result, nil
}

func (self *MongoEngine) inputToTable(input *schema.Input) *schema.Table {
	table := &schema.Table{Model: input.Model}
	convertor.CopyProperties(table, input)
	// 复制源数据, 防止字段丢失的问题
	table.Source = input.Source
	return table
}

func (self *MongoEngine) resetTable(skma *schema.Table) *schema.Table {
	if len(skma.Fields) == 0 {
		return skma
	}
	clone := convertor.DeepClone(skma)
	for i := 0; i < len(clone.Fields); i++ {
		theField := clone.Fields[i]
		flatten := consts.GTYPE_FLATTEN == theField.GType
		clone.Fields[i].Index = support.If(
			flatten, theField.Field, theField.UUKey,
		)
	}
	return clone
}

func (self *MongoEngine) resetPivot(skm *schema.Digest) *schema.Digest {
	skm.Table = self.resetTable(skm.Table)
	// 强制显示 digest 取值字段
	if skm.PivotBy == nil {
		skm.PivotBy = &schema.PivotBy{
			State: "none",
		}
	}
	if schema.PivotByActive(skm.PivotBy) {
		var count, shown = 0, []string{
			skm.PivotBy.Value, skm.PivotBy.Pivot,
		}
		for idx, field := range skm.Table.Fields {
			if slice.Contain(shown, field.UUKey) {
				skm.Table.Fields[idx].Shown = true
				if count += 1; count == len(shown) {
					break
				}
			}
		}
	}
	return skm
}

func (self *MongoEngine) buildQuery(logic string, queries *[]schema.Query, skma *schema.Table) bson.M {
	if queries == nil || len(*queries) == 0 {
		return bson.M{}
	}

	clauses := []bson.M{}
	for _, query := range *queries {
		// 子查询(如 ITEMS:OR)不应该作为字段写入过滤条件，
		// 而应作为独立逻辑子句加入当前层的组合条件中。
		if slice.Contain(schema.SubQuery, query.Logic) {
			if query.Items == nil || len(*query.Items) == 0 {
				continue
			}
			child := self.buildQuery(query.Logic, query.Items, skma)
			if len(child) > 0 {
				clauses = append(clauses, child)
			}
			continue
		}

		field := skma.GetField(query.Field)
		if field == nil || field.Index == "" {
			continue
		}
		if field.GType == consts.GTYPE_FLATTEN {
			query.Field = field.Field
		}
		// 数值类型转换为浮点数
		if field.FType == consts.FTYPE_NUMERIC {
			query.Value = field.Parse(query.Value)
		}

		expr := bson.M{}
		switch strings.ToUpper(query.Logic) {
		case consts.LOGIC_EQUALSTO:
			expr[query.Field] = query.Value
		case consts.LOGIC_NOTEQUAL:
			expr[query.Field] = bson.M{"$ne": query.Value}
		case consts.LOGIC_STR_LIKE:
			expr[query.Field] = bson.M{
				"$regex":   query.Value,
				"$options": "i",
			}
		case consts.LOGIC_INCLUDES:
			expr[query.Field] = bson.M{"$in": query.Value}
		case consts.LOGIC_CONTAINS:
			expr[query.Field] = bson.M{"$in": query.Value}
		case consts.LOGIC_LESTHAN:
			expr[query.Field] = bson.M{"$lt": query.Value}
		case consts.LOGIC_GREATER:
			expr[query.Field] = bson.M{"$gt": query.Value}
		case consts.LOGIC_LESS_EQ:
			expr[query.Field] = bson.M{"$lte": query.Value}
		case consts.LOGIC_GRAT_EQ:
			expr[query.Field] = bson.M{"$gte": query.Value}
		case consts.LOGIC_BETWEEN:
			value := query.Value.([]any)
			if field.FType == consts.FTYPE_NUMERIC {
				start := support.ParseNumber(value[0])
				finish := support.ParseNumber(value[1])
				expr[query.Field] = bson.M{
					"$gte": start, "$lte": finish,
				}
			}
			if field.FType == consts.FTYPE_DATETIME {
				start := support.ParseDate(value[0])
				finish := support.ParseDate(value[1])
				expr[query.Field] = bson.M{
					"$gte": start, "$lt": finish,
				}
			}
		case consts.LOGIC_EXISTS:
			expr[query.Field] = bson.M{"$exists": true}
		case consts.LOGIC_VAL_NULL:
			expr[query.Field] = bson.M{"$in": []any{nil, ""}}
		case consts.LOGIC_NOT_NULL:
			expr[query.Field] = bson.M{"$exists": true, "$nin": []any{nil, ""}}
		default:
			// 异常逻辑
		}
		if len(expr) > 0 {
			clauses = append(clauses, expr)
		}
	}

	if len(clauses) == 0 {
		return bson.M{}
	}
	switch logic {
	case consts.LOGIC_SUBOR:
		return bson.M{"$or": clauses}
	case consts.LOGIC_SUBAND:
		return bson.M{"$and": clauses}
	case consts.LOGIC_SUBNOT:
		return bson.M{"$nor": clauses}
	}
	if len(clauses) == 1 {
		return clauses[0]
	}
	return bson.M{"$and": clauses}
}

func (self *MongoEngine) getGroupBy(skma *schema.Digest) []source.GroupBy {
	groups, keys := []source.GroupBy{}, []string{}
	if schema.PivotByActive(skma.PivotBy) {
		pivotIdx := skma.PivotBy.Pivot
		head := source.GroupBy{Index: pivotIdx}
		for _, item := range skma.GroupBy {
			if item.Index == pivotIdx {
				head.SortBy = item.SortBy
				head.Format = item.Format
				break
			}
		}
		keys = append(keys, pivotIdx)
		groups = append(groups, head)
	}
	for _, item := range skma.GroupBy {
		if slice.Contain(keys, item.Index) {
			continue
		}
		keys = append(keys, item.Index)
		groups = append(groups, item)
	}
	return groups
}

func (self *MongoEngine) getAggrs(skma *schema.Digest, nested string) []source.CountFn {
	totals := []source.CountFn{}
	scene := skma.Table.Request.Scene
	if scene == consts.SCENE_SEARCH {
		for _, field := range skma.Table.GetFields() {
			if field.Shown == false {
				continue
			}
			// 数据列表显示汇总字段
			skma.CountFn = append(skma.CountFn, source.CountFn{
				Label: field.UUKey, Index: field.UUKey,
			})
		}
	}
	if scene == consts.SCENE_KANBAN {
		for _, field := range skma.CountFn {
			field := skma.Table.GetField(field.Index)
			if field == nil || field.Shown == true {
				continue
			}
			field.Shown = true
		}
	}

	// 构建聚合操作,sql: sum,count,avg,min,max,distinct
	for _, item := range skma.CountFn {
		field := skma.Table.GetField(item.Index)
		if field == nil || field.Shown == false {
			continue
		}
		if nested != "" && field.Group != nested {
			continue
		}

		// 禁用字段不参与聚合
		if field.Extra.Disabled == consts.DISABLED_ALWAYS {
			continue
		}
		aggrkey := strings.Replace(field.UUKey, ".", "#", 1)
		switch dtype := field.GetDataType(); dtype {
		case consts.DTYPE_INTEGER, consts.DTYPE_TINYINT,
			consts.DTYPE_LONGINT, consts.DTYPE_SCALED,
			consts.DTYPE_EXPENSE, consts.DTYPE_DECIMAL:
			if !consts.IsMetric(item.Func) {
				item.Func = consts.VALUE_SUM
			}
			totals = append(totals, source.CountFn{
				Label: aggrkey, Index: field.Index,
				Func: item.Func, Option: nil,
			})
		case consts.DTYPE_SERIALNO, consts.DTYPE_RELATION,
			consts.DTYPE_OPTIONAL, consts.DTYPE_WORKFLOW:
			if !consts.IsUnique(item.Func) {
				item.Func = consts.VALUE_CNT
			}
			totals = append(totals, source.CountFn{
				Label: aggrkey, Index: field.Index,
				Func: item.Func, Option: nil,
			})
		case consts.DTYPE_KEYWORDS, consts.DTYPE_SUBJECT,
			consts.DTYPE_X_EMAIL, consts.DTYPE_X_PHONE:
			if !consts.IsUnique(item.Func) {
				item.Func = consts.VALUE_CNT
			}
			totals = append(totals, source.CountFn{
				Label: aggrkey, Index: field.Index,
				Func: item.Func, Option: nil,
			})
		case consts.DTYPE_DATETIME, consts.DTYPE_ONLYDATE:
			// 日期字段仅支持: MIN / MAX / CNT / UNQ
			switch strings.ToUpper(item.Func) {
			case consts.VALUE_MIN, consts.VALUE_MAX,
				consts.VALUE_CNT, consts.VALUE_UNQ:
			default:
				item.Func = consts.VALUE_CNT
			}
			option := map[string]any(nil)
			// 日期去重按“日”粒度（yyyy-mm-dd）统计，而不是按完整时间戳去重
			if strings.ToUpper(item.Func) == consts.VALUE_UNQ {
				option = bson.M{"$dateToString": bson.M{
					"date":     "$" + field.Index,
					"format":   "%Y-%m-%d",
					"timezone": time.Local.String(),
				}}
			}
			totals = append(totals, source.CountFn{
				Label: aggrkey, Index: field.Index,
				Func: item.Func, Option: option,
			})
		case consts.DTYPE_DOC_FILE, consts.DTYPE_IMG_FILE,
			consts.DTYPE_LONGTEXT, consts.DTYPE_RICHTEXT:
			continue
		case consts.DTYPE_LOCATION:
			support.LogWarn(skma.Table.Request.LogID,
				"cannot aggregate so skip location field",
				field.UUKey, "of model", skma.Table.Model.UUKey,
			)
			continue
		// case consts.FTYPE_DATETIME:
		// 	totals = append(totals, schema.Aggr{
		// 		Label: field.UUKey, Index: field.Index,
		// 		Method: consts.VALUE_MAX, Option: nil,
		// 	})
		default:
			support.LogWarn(skma.Table.Request.LogID,
				"unrecongize:", field.UUKey, dtype,
				"of model", skma.Table.Model.UUKey,
			)
		}
	}
	return totals
}

// totals to group by and count(1),sum(1)...
func (self *MongoEngine) buildCountFn(totals []source.CountFn) aggrStage {
	stages := aggrStage{
		Group:   bson.M{},
		Unique:  bson.M{},
		Project: bson.M{},
	}
	if len(totals) == 0 {
		return stages
	}
	for _, item := range totals {
		if len(item.Items) > 0 {
			childStages := self.buildCountFn(item.Items)
			stages.Group = maputil.Merge(stages.Group, childStages.Group)
			stages.Unique = maputil.Merge(stages.Unique, childStages.Unique)
			stages.Project = maputil.Merge(stages.Project, childStages.Project)
		}

		item.Label = strings.Replace(item.Label, ".", "#", 1)
		label, index := "$"+item.Label, "$"+item.Index
		switch strings.ToUpper(item.Func) {
		// 取值
		case consts.VALUE_AVG:
			stages.Group[item.Label] = bson.M{"$avg": label}
			stages.Project[item.Label] = bson.M{"$avg": index}
		case consts.VALUE_SUM:
			stages.Group[item.Label] = bson.M{"$sum": label}
			// 显式转换为数值，避免字符串数字或脏数据
			// 在 $sum 中被按 0 处理, 导致结果不准确
			stages.Project[item.Label] = bson.M{
				"$convert": bson.M{
					"input": index, "to": "double",
					"onError": 0, "onNull": 0,
				},
			}
		case consts.VALUE_MAX:
			stages.Group[item.Label] = bson.M{"$max": label}
			stages.Project[item.Label] = bson.M{"$max": index}
		case consts.VALUE_MIN:
			stages.Group[item.Label] = bson.M{"$min": label}
			stages.Project[item.Label] = bson.M{"$min": index}
		case consts.VALUE_CNT:
			stages.Group[item.Label] = bson.M{"$sum": label}
			stages.Project[item.Label] = bson.M{"$sum": 1}
		case consts.VALUE_UNQ:
			// 在 project 阶段映射原始字段（日期去重可通过 Option 先投影到日粒度）
			stages.Project[item.Label] = index
			if support.Bool(item.Option) {
				stages.Project[item.Label] = item.Option
			}
			// 在 group 阶段使用 $addToSet 收集唯一值到数组
			// 注意：$addToSet 在处理大量唯一值时会消耗内存（MongoDB 4.2.3+ 限制约 100MB）
			// 但在实际使用中，VALUE_UNQ 主要用于 optional/workflow/relation 等枚举类型字段
			// 这些字段属于有限枚举，唯一值数量通常只有几十到几百个，最多几千个
			// 因此即使在30万行数据的情况下，每个分组内的唯一值数量也很少，性能压力很小
			stages.Group[item.Label] = bson.M{"$addToSet": label}
			// 在后续的 addFields 阶段使用 $size 计算数组大小（保留其他字段）
			stages.Unique[item.Label] = bson.M{"$size": "$" + item.Label}
		// 分组
		case consts.VALUE_TERMS, consts.VALUE_HIST2:
			//  "basic#title":  "$title",
			stages.Project[item.Label] = index
			// 维度字段使用 $first 保留分组键对应值，避免对字符串执行 $sum
			stages.Group[item.Label] = bson.M{"$first": label}
			if support.Bool(item.Option) {
				stages.Project[item.Label] = item.Option
			}
		default:
			// 异常逻辑
		}
	}

	return stages
}
func (self *MongoEngine) buildGroupBy(totals []source.CountFn) bson.M {
	group := bson.M{}
	if len(totals) == 0 {
		return group
	}
	for _, item := range totals {
		if len(item.Items) > 0 {
			child := self.buildGroupBy(item.Items)
			group = maputil.Merge(group, child)
		}

		label := strings.Replace(item.Label, ".", "#", 1)
		switch strings.ToUpper(item.Func) {
		case consts.VALUE_TERMS, consts.VALUE_HIST2:
			group[label] = fmt.Sprint("$", label)
		default:
			// 异常逻辑
		}
	}

	return group
}
