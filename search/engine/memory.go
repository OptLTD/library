package engine

import (
	"fmt"
	"log"
	"github.com/OptLTD/library/search/consts"
	"github.com/OptLTD/library/search/respond"
	"github.com/OptLTD/library/search/schema"
	"github.com/OptLTD/library/search/source"
	"github.com/OptLTD/library/search/support"
	"sort"
	"strings"
	"sync"

	"github.com/duke-git/lancet/v2/maputil"
	"github.com/duke-git/lancet/v2/slice"
)

// MemoryEngine 内存存储引擎，使用 map 存储数据
type MemoryEngine struct {
	debug bool

	// 存储结构: table -> uukey -> data
	storage map[string]map[string]map[string]any
	mu      sync.RWMutex
	handles []ICallable
}

// NewMemoryEngine 创建新的内存引擎实例
func NewMemoryEngine() *MemoryEngine {
	return &MemoryEngine{
		storage: make(map[string]map[string]map[string]any),
		handles: []ICallable{},
	}
}

func (self *MemoryEngine) Using(handle ICallable) IEngine {
	if self.handles == nil {
		self.handles = []ICallable{}
	}
	self.handles = append(self.handles, handle)
	return self
}

func (self *MemoryEngine) First(skma *schema.Input, record *respond.Record) error {
	table := skma.Model.Source
	where := map[string]any{
		consts.FIELD_UUKEY: record.UUKey,
	}
	where = maputil.Merge(skma.Scope, where)
	preset, _ := schema.BuildQuery(where)

	self.mu.RLock()
	defer self.mu.RUnlock()

	tableData, exists := self.storage[table]
	if !exists {
		record.Storage = nil
		record.Objects = record.ToObjects(skma, record.Request)
		record.Changes = record.ToChanges(skma, record.Request)
		record.Changed = record.ToFlatten(skma, record.Objects)
		record.Prepare = record.ToPrepare(skma, record.Objects)
		return nil
	}

	// 查找匹配的记录
	for uukey, data := range tableData {
		if self.matchQuery(data, &preset, &skma.Fields) {
			record.Storage = data
			record.UUKey = uukey
			break
		}
	}

	record.Objects = record.ToObjects(skma, record.Request)
	record.Changes = record.ToChanges(skma, record.Request)
	record.Prepare = record.ToPrepare(skma, record.Objects)
	record.Changed = record.ToFlatten(skma, record.Objects)
	return nil
}

func (self *MemoryEngine) Store(skma *schema.Input, record *respond.Record) error {
	table := skma.Model.Source

	// 处理回调
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

	upsert := record.Format(skma, record.Prepare)
	self.mu.Lock()
	defer self.mu.Unlock()

	if self.storage[table] == nil {
		self.storage[table] = make(map[string]map[string]any)
	}

	if record.Event == "" || record.Event == "INSERT" {
		merged := maputil.Merge(record.Default, upsert)
		self.storage[table][record.UUKey] = merged
	} else {
		existing := self.storage[table][record.UUKey]
		if existing == nil {
			existing = make(map[string]any)
		}
		merged := maputil.Merge(existing, upsert)
		self.storage[table][record.UUKey] = merged
	}

	// 处理回调
	for _, handle := range self.handles {
		err := handle.HandleUpsert(skma, record)
		if err != nil {
			break
		}
	}
	return nil
}

func (self *MemoryEngine) Select(skma *schema.Input, records []*respond.Record) error {
	table := skma.Model.Source
	uukeys, size := []any{}, len(records)
	for i := 0; i < size; i++ {
		uukeys = append(uukeys, records[i].UUKey)
	}

	self.mu.RLock()
	defer self.mu.RUnlock()

	tableData, exists := self.storage[table]
	if !exists {
		tableData = make(map[string]map[string]any)
	}

	mapped := make(map[string]map[string]any)
	for uukeyStr, data := range tableData {
		// 检查是否在 uukeys 列表中
		contained := false
		for _, uukey := range uukeys {
			if uukey == uukeyStr {
				contained = true
				break
			}
		}
		if contained {
			// 对于 Select，直接匹配 uukey，不需要复杂的查询匹配
			mapped[uukeyStr] = data
		}
	}

	for i := 0; i < size; i++ {
		if value, ok := mapped[records[i].UUKey]; ok {
			records[i].Storage = value
		}
	}
	return nil
}

func (self *MemoryEngine) Upsert(skma *schema.Input, records []*respond.Record) error {
	table, size := skma.Model.Source, len(records)

	// 处理回调
	for _, handle := range self.handles {
		for i := 0; i < size; i++ {
			err := handle.BeforeUpsert(skma, records[i])
			if err != nil {
				return err
			}
		}
	}

	self.mu.Lock()
	defer self.mu.Unlock()

	if self.storage[table] == nil {
		self.storage[table] = make(map[string]map[string]any)
	}

	for i := 0; i < size; i++ {
		record := records[i]
		combine := maputil.OmitByKeys(record.Prepare, maputil.Keys(record.Default))
		values := record.Format(skma, combine)

		existing := self.storage[table][record.UUKey]
		if existing == nil {
			existing = make(map[string]any)
		}

		// 合并数据：$set 更新，$setOnInsert 仅在不存在时设置
		merged := maputil.Merge(record.Default, existing)
		merged = maputil.Merge(merged, values)
		self.storage[table][record.UUKey] = merged
	}

	// 处理回调
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

func (self *MemoryEngine) Search(skma *schema.Table) (*respond.Result, error) {
	self.resetIndex(skma)
	request, table := skma.Request, skma.Model.Source
	query := skma.BuildQuery()

	self.mu.RLock()
	defer self.mu.RUnlock()

	tableData, exists := self.storage[table]
	if !exists {
		return &respond.Result{
			Page:   request.Page,
			Size:   request.Size,
			Count:  0,
			Values: []map[string]any{},
		}, nil
	}

	// 过滤数据
	matched := []map[string]any{}
	for _, data := range tableData {
		if self.matchQuery(data, &query, &skma.Fields) {
			matched = append(matched, data)
		}
	}

	count := int64(len(matched))
	offset := int((request.Page - 1) * request.Size)
	limit := int(request.Size)

	// 排序
	if order := request.Order; order != nil {
		self.sortData(matched, order.Field, order.Order)
	}

	// 分页
	values := []map[string]any{}
	if offset < len(matched) {
		end := offset + limit
		if end > len(matched) {
			end = len(matched)
		}
		values = matched[offset:end]
	}

	record := respond.Record{}
	result := &respond.Result{
		Page:  request.Page,
		Size:  request.Size,
		Count: uint64(count),
		Values: slice.Map(values, func(idx int, item map[string]any) map[string]any {
			record.Storage = item
			value := record.ToObjects(skma, item)
			// handle plain data
			for _, field := range skma.GetFields() {
				switch strings.ToUpper(field.GType) {
				case consts.GTYPE_GROUPED:
					continue
				}
				group, ok := map[string]any{}, false
				if group, ok = value[field.Group].(map[string]any); !ok {
					group = map[string]any{}
				}
				if val, ok := item[field.Field]; ok {
					group[field.Field] = val
				}
				value[field.Group] = group
			}
			return record.ToFlatten(skma, value)
		}),
	}

	// 添加合计行（简化实现，不调用 Pivot 避免递归）
	// 可以在这里实现简单的聚合逻辑
	result.Totals = map[string]any{}

	// 处理回调
	for _, h := range self.handles {
		err := h.SearchResult(skma, result)
		if err != nil {
			log.Println("handle result err:", err)
		}
	}
	return result, nil
}

func (self *MemoryEngine) Update(skma *schema.Input, data map[string]any) error {
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
	where := skma.Scope
	preset, _ := schema.BuildQuery(where)

	self.mu.Lock()
	defer self.mu.Unlock()

	tableData, exists := self.storage[table]
	if !exists {
		return nil
	}

	// 更新前先 count，超过 1000 条则阻止
	count := int64(0)
	for _, record := range tableData {
		if self.matchQuery(record, &preset, &skma.Fields) {
			count++
		}
	}
	if count > 1000 {
		return fmt.Errorf("update would affect %d records, exceeds limit of 1000", count)
	}

	// 查找匹配的记录并更新
	for uukey, record := range tableData {
		if self.matchQuery(record, &preset, &skma.Fields) {
			// 合并更新数据
			updated := maputil.Merge(record, data)
			tableData[uukey] = updated
		}
	}

	return nil
}

func (self *MemoryEngine) Digest(skma *schema.Digest) (*respond.Result, error) {
	// 简化实现，返回空的结果，避免与 Search 的递归调用
	// 在实际使用中，Pivot 应该实现聚合逻辑
	request := skma.Table.Request
	return &respond.Result{
		Page: request.Page,
		Size: request.Size,
		// TODO: 实现聚合逻辑
		Totals: map[string]any{},
		Values: []map[string]any{},
	}, nil
}

func (self *MemoryEngine) resetIndex(skma *schema.Table) {
	if len(skma.Fields) == 0 {
		return
	}
	matched, order := false, skma.Request.Order
	for i := 0; i < len(skma.Fields); i++ {
		theField := skma.Fields[i]
		flatten := consts.GTYPE_FLATTEN == theField.GType
		skma.Fields[i].Index = support.If(
			flatten, theField.Field, theField.UUKey,
		)
		if order != nil && order.Field == theField.UUKey {
			matched = true
			order.Field = skma.Fields[i].Index
		}
	}
	if matched == false && order != nil {
		order.Field = "basic.created_at"
	}
}

// matchQuery 检查数据是否匹配查询条件
func (self *MemoryEngine) matchQuery(data map[string]any, queries *[]schema.Query, fields *[]source.Field) bool {
	if len(*queries) == 0 {
		return true
	}

	for _, query := range *queries {
		field, ok := slice.Find(*fields, func(idx int, item source.Field) bool {
			return item.UUKey == query.Field
		})
		if !ok || field == nil {
			continue
		}

		// 对于 FLATTEN 类型，使用 field.Field（扁平化的字段名）
		// 对于其他类型，可能需要使用 field.Index 或 field.UUKey
		var value any
		var exists bool

		if field.GType == consts.GTYPE_FLATTEN {
			// 扁平化字段，直接使用 field.Field
			value, exists = data[field.Field]
			if !exists && field.Index != "" {
				value, exists = data[field.Index]
			}
		} else {
			// 非扁平化字段，尝试使用 Index 或 UUKey
			if field.Index != "" {
				value, exists = data[field.Index]
			}
			if !exists {
				value, exists = data[field.UUKey]
			}
		}

		if !self.matchField(value, query, *field) {
			return false
		}
	}
	return true
}

// matchField 检查单个字段是否匹配查询条件
func (self *MemoryEngine) matchField(value any, query schema.Query, field source.Field) bool {
	switch strings.ToUpper(query.Logic) {
	case consts.LOGIC_EQUALSTO:
		return self.compareEqual(value, query.Value)
	case consts.LOGIC_NOTEQUAL:
		return !self.compareEqual(value, query.Value)
	case consts.LOGIC_STR_LIKE:
		return self.compareLike(value, query.Value)
	case consts.LOGIC_INCLUDES:
		return self.compareIn(value, query.Value)
	case consts.LOGIC_CONTAINS:
		return self.compareContains(value, query.Value)
	case consts.LOGIC_LESTHAN:
		return self.compareLessThan(value, query.Value)
	case consts.LOGIC_GREATER:
		return self.compareGreaterThan(value, query.Value)
	case consts.LOGIC_LESS_EQ:
		return self.compareLessEqual(value, query.Value)
	case consts.LOGIC_GRAT_EQ:
		return self.compareGreaterEqual(value, query.Value)
	case consts.LOGIC_BETWEEN:
		return self.compareBetween(value, query.Value, field)
	case consts.LOGIC_EXISTS:
		return value != nil
	case consts.LOGIC_VAL_NULL:
		return value == nil || value == ""
	case consts.LOGIC_NOT_NULL:
		return value != nil && value != ""
	case consts.LOGIC_SUBOR:
		if query.Items == nil {
			return true
		}
		for _, item := range *query.Items {
			if self.matchField(value, item, field) {
				return true
			}
		}
		return false
	case consts.LOGIC_SUBAND:
		if query.Items == nil {
			return true
		}
		for _, item := range *query.Items {
			if !self.matchField(value, item, field) {
				return false
			}
		}
		return true
	case consts.LOGIC_SUBNOT:
		if query.Items == nil {
			return true
		}
		for _, item := range *query.Items {
			if self.matchField(value, item, field) {
				return false
			}
		}
		return true
	default:
		return true
	}
}

// 比较函数
func (self *MemoryEngine) compareEqual(a, b any) bool {
	return a == b
}

func (self *MemoryEngine) compareLike(a, b any) bool {
	if a == nil || b == nil {
		return false
	}
	strA := strings.ToLower(fmt.Sprint(a))
	strB := strings.ToLower(fmt.Sprint(b))
	return strings.Contains(strA, strings.Trim(strB, "%"))
}

func (self *MemoryEngine) compareIn(a, b any) bool {
	if b == nil {
		return false
	}
	arr, ok := b.([]any)
	if !ok {
		return false
	}
	for _, item := range arr {
		if a == item {
			return true
		}
		// 存储侧常为 float64，查询字面量常为 int，需数值等价
		if self.compareNumeric(a, item) == 0 {
			return true
		}
	}
	return false
}

func (self *MemoryEngine) compareContains(a, b any) bool {
	return self.compareIn(a, b)
}

func (self *MemoryEngine) compareLessThan(a, b any) bool {
	return self.compareNumeric(a, b) < 0
}

func (self *MemoryEngine) compareGreaterThan(a, b any) bool {
	return self.compareNumeric(a, b) > 0
}

func (self *MemoryEngine) compareLessEqual(a, b any) bool {
	return self.compareNumeric(a, b) <= 0
}

func (self *MemoryEngine) compareGreaterEqual(a, b any) bool {
	return self.compareNumeric(a, b) >= 0
}

func (self *MemoryEngine) compareBetween(a, b any, field source.Field) bool {
	if b == nil {
		return false
	}
	arr, ok := b.([]any)
	if !ok || len(arr) != 2 {
		return false
	}
	val := self.toNumeric(a)
	start := self.toNumeric(arr[0])
	end := self.toNumeric(arr[1])
	return val >= start && val <= end
}

func (self *MemoryEngine) compareNumeric(a, b any) int {
	valA := self.toNumeric(a)
	valB := self.toNumeric(b)
	if valA < valB {
		return -1
	} else if valA > valB {
		return 1
	}
	return 0
}

func (self *MemoryEngine) toNumeric(v any) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case int:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case float32:
		return float64(val)
	case float64:
		return val
	case string:
		return support.ParseNumber(val)
	default:
		return 0
	}
}

// sortData 对数据进行排序
func (self *MemoryEngine) sortData(data []map[string]any, ordBy string, order string) {
	sort.Slice(data, func(i, j int) bool {
		valA := data[i][ordBy]
		valB := data[j][ordBy]
		compare := self.compareNumeric(valA, valB)
		if order == "asc" {
			return compare < 0
		}
		return compare > 0
	})
}
