package respond

import (
	"fmt"
	"strings"
	"time"

	"github.com/OptLTD/library/search/consts"
	"github.com/OptLTD/library/search/request"
	"github.com/OptLTD/library/search/schema"
	"github.com/OptLTD/library/search/source"
	"github.com/OptLTD/library/search/support"

	"github.com/duke-git/lancet/v2/maputil"
	"github.com/duke-git/lancet/v2/slice"
	jsoniter "github.com/json-iterator/go"
)

type Record struct {
	UUKey  string `json:"uukey"`
	Model  string `json:"model"`
	LogID  string `json:"logid"`
	OpType string `json:"opType,omitempty"`
	Action string `json:"action,omitempty"`

	Exists bool   `json:"exists"` // 是否存在
	System object `json:"system"` // 系统数据
	Unique object `json:"unique"` // 唯一数据

	Default object `json:"default"` // 存储时的默认数据
	Request object `json:"request"` // 请求过来的数据
	Storage object `json:"storage"` // 数据库存储的数据
	Prepare object `json:"prepare"` // 数据处理合并数据
	// 以下数据都是基于 Storage 数据计算的
	Current object `json:"current"` // 数据库对象准换成扁平数据
	Changed object `json:"changed"` // 合并后对象最终扁平数据
	Changes object `json:"changes"` // 数据处理改动数据
	Objects object `json:"objects"` // 多层嵌套的对象数据
}

func (self *Record) parse(skma schema.ISchema, i int, data object) object {
	var parsed = object{}
	fields := skma.GetFields()
	group := skma.GetGroups()[i]
	for i := 0; i < len(fields); i++ {
		var field = fields[i]
		if field.Group != group.UUKey {
			continue
		}

		// 解析字段
		if get, ok := data[field.Field]; ok {
			parsed[field.Field] = field.Parse(get)
		}
	}
	return parsed
}

// mysql json string 字段需要解析成 json object
func (self *Record) Decode(skma schema.ISchema, data object) object {
	result := object{}
	groups := skma.GetGroups()
	fields := skma.GetFields()
	for _, group := range groups {
		gtype := strings.ToUpper(group.GType)
		// 没匹配到数据
		if gtype != consts.GTYPE_FLATTEN {
			value, ok := data[group.UUKey]
			if !ok || value == nil {
				continue
			}
			switch gtype {
			case consts.GTYPE_GROUPED:
				object := object{}
				if jsoniter.UnmarshalFromString(value.(string), &object) == nil {
					result[group.UUKey] = object
				}
			default:
				array := []any{}
				if jsoniter.UnmarshalFromString(value.(string), &array) == nil {
					result[group.UUKey] = array
				}
			}
			continue
		}
		object := object{}
		for _, field := range fields {
			if field.Group != group.UUKey {
				continue
			}
			// 优先 field, 取不到取 index
			value, ok := data[field.Field]
			if !ok {
				value, ok = data[field.Index]
			}
			if ok && value != nil {
				// 格式化时间
				if field.FType == consts.FTYPE_DATETIME {
					if val, ok := value.(time.Time); ok {
						value = val.String()
					}
				}
				object[field.Field] = value
				if field.Extra.Embedded {
					result[field.Field] = value
				}
			}
		}
		result[group.UUKey] = object
	}
	return result
}

// mysql json object 字段需要编码成 json string
func (self *Record) Encode(skma schema.ISchema, data object) object {
	result := object{}
	groups := skma.GetGroups()
	fields := skma.GetFields()
	for _, group := range groups {
		gtype := strings.ToUpper(group.GType)
		// 没匹配到数据
		if gtype != consts.GTYPE_FLATTEN {
			if value, ok := data[group.UUKey]; ok {
				json, err := jsoniter.MarshalToString(value)
				if err == nil {
					result[group.UUKey] = json
				}
				continue
			}
			continue
		}

		for _, field := range fields {
			if field.Group != group.UUKey {
				continue
			}
			if value, ok := data[field.Field]; ok {
				result[field.Field] = value
			}
		}
	}
	return result
}

/**
 * storage data format
 * not nested data
 * not flatten data
 */
func (self *Record) Format(skma schema.ISchema, data object) object {
	result := object{}
	groups := skma.GetGroups()
	for i, group := range groups {
		gtype := strings.ToUpper(group.GType)
		if gtype == consts.GTYPE_FLATTEN {
			parsed := self.parse(skma, i, data)
			result = maputil.Merge(result, parsed)
			continue
		}

		// 没匹配到数据
		value, ok := data[group.UUKey]
		if !ok || value == nil {
			continue
		}
		if group.Extra.Multiple == false {
			if object, ok := value.(object); ok {
				parsed := self.parse(skma, i, object)
				result[group.UUKey] = parsed
			}
			continue
		}
		// collection
		if array, ok := value.([]any); ok {
			result[group.UUKey] = slice.Map(array, func(idx int, item any) object {
				if val, ok := item.(object); ok {
					return self.parse(skma, i, val)
				}
				return nil
			})
		}
	}
	return result
}

// FillNilData 在扁平请求（如 basic.name）上，
// 为 schema 中应有但未出现的字段补上 nil，
// 便于 parse/后续合并把「未传」与「不传」区分开。
func (self *Record) FillNilData(skma schema.ISchema, request object) object {
	if request == nil {
		request = object{}
	}
	groups := skma.GetGroups()
	groupByUU := make(map[string]source.Group, len(groups))
	for _, g := range groups {
		groupByUU[g.UUKey] = g
	}
	for _, field := range skma.GetFields() {
		group, ok := groupByUU[field.Group]
		if !ok {
			continue
		}
		if group.Extra.Relation != "" {
			continue
		}
		if field.Extra.Disabled == consts.DISABLED_ALWAYS {
			continue
		}
		gtype := strings.ToUpper(group.GType)
		if gtype == consts.GTYPE_GROUPED && group.Extra.Multiple {
			request[field.Group] = nil
			continue
		}
		key := field.GetKey()
		if _, ok := request[key]; !ok {
			request[key] = nil
		}
	}
	return request
}

func (self *Record) PickObject(group *source.Group, data object) object {
	result := object{}
	for key, val := range data {
		suffix := strings.Replace(key, group.UUKey+".", "", 1)
		if suffix != key { // replace success
			result[suffix] = val
		}
	}
	return result
}

func (self *Record) GetRelated(group *source.Group) *request.Record {
	if group.Extra.Relation == "" {
		return nil
	}

	// 从 record.Changed 中获取该分组的变化数据
	changed, ok := self.Changes[group.UUKey]
	if !ok || changed == nil {
		return nil // 该分组没有变化，返回 nil
	}

	request := &request.Record{
		Model: group.Extra.Relation,
		LogID: self.LogID, Scene: self.OpType,
	}

	if group.Extra.Multiple {
		// Multiple=true：使用 Batch 批量处理
		if batch, ok := changed.([]any); !ok {
			return nil
		} else {
			var wrap = func(idx int, item any) any {
				if data, ok := item.(object); ok {
					data[consts.REFER_MODEL] = self.Model
					data[consts.REFER_VALUE] = self.UUKey
					return object{consts.GROUP_BASIC: data}
				}
				return object{consts.GROUP_BASIC: item}
			}
			request.Batch = slice.Map(batch, wrap)
		}
	} else {
		// Multiple=false：使用 Value 单个处理
		if data, ok := changed.(object); !ok {
			return nil
		} else {
			data[consts.REFER_MODEL] = self.Model
			data[consts.REFER_VALUE] = self.UUKey
			request.Value = object{
				consts.GROUP_BASIC: data,
			}
		}
	}

	return request
}

func (self *Record) ToObjects(skma schema.ISchema, data object) object {
	result := object{}
	for i, group := range skma.GetGroups() { // handle group
		if value, ok := data[group.UUKey]; ok {
			result[group.UUKey] = value
			continue
		}
		picked := self.PickObject(&group, data)
		parsed := self.parse(skma, i, picked)
		result[group.UUKey] = parsed
	}
	return result
}

func (self *Record) ToCurrent(skma schema.ISchema, data object) object {
	var objects = object{}
	for i, group := range skma.GetGroups() { // handle group
		gtype := strings.ToUpper(group.GType)
		if gtype == consts.GTYPE_FLATTEN {
			parsed := self.parse(skma, i, data)
			objects[group.UUKey] = parsed
			continue
		}
		if value, ok := data[group.UUKey]; ok {
			objects[group.UUKey] = value
		}
	}
	return self.ToFlatten(skma, objects)
}

func (self *Record) ToFlatten(skma schema.ISchema, objects object) object {
	groups := skma.GetGroups()
	fields := skma.GetFields()

	var result = object{}
	var expand *schema.Field
	for _, group := range groups {
		value, ok := objects[group.UUKey]
		if !ok || value == nil {
			continue
		}
		if group.Extra.Multiple {
			if support.GetType(value) == "array" {
				result[group.UUKey] = value
			}
			if support.GetType(value) == "object" {
				result[group.UUKey] = []any{value}
			}
			continue
		}
		var child, ok1 = value.(object)
		if !ok1 || len(child) == 0 {
			continue
		}
		for _, field := range fields {
			if field.Group != group.UUKey {
				continue
			}
			if field.Extra.Disabled == consts.DISABLED_ALWAYS {
				continue
			}
			if field.FType == consts.FTYPE_DATETIME {
				if field.HasTimeTerm() {
					expand = &field
				}
			}
			if value, ok := child[field.Field]; ok {
				result[field.UUKey] = value
			}
		}
	}

	// 按期,按周字段fill to flatten
	if expand == nil || !expand.HasTimeTerm() {
		return result
	}
	date := self.asTime(result[expand.UUKey])
	expands := self.ExpandTime(skma, date)
	for key, value := range expands {
		result["basic."+key] = value
	}
	return result
}

func (self *Record) ToPrepare(skma schema.ISchema, flatten object) object {
	groups := skma.GetGroups()
	fields := skma.GetFields()

	var result = object{}
	var current = self.Current
	var expand *schema.Field
	var dateTime *time.Time
	for _, group := range groups {
		// 关联数据不参与合并
		if group.Extra.Relation != "" {
			continue
		}

		// 处理collection
		if group.Extra.Multiple {
			source, ok1 := flatten[group.UUKey]
			target, ok2 := current[group.UUKey]
			if !ok1 && !ok2 {
				continue
			}

			if ok1 && ok2 && false {
				toMap := func(item any) string {
					object, ok := item.(object)
					if !ok {
						return "TO_DISCARD"
					}
					uukey, ok := object["uukey"]
					if !ok {
						return "TO_DISCARD"
					}
					return uukey.(string)
				}
				sourceMap := slice.KeyBy(source.([]any), toMap)
				targetMap := slice.KeyBy(target.([]any), toMap)
				mergedMap := maputil.Merge(targetMap, sourceMap)
				delete(mergedMap, "TO_DISCARD")
				result[group.UUKey] = maputil.Values(mergedMap)
				continue
			}
			if ok1 {
				result[group.UUKey] = source
			} else {
				result[group.UUKey] = target
			}
			continue
		}

		var parts = object{}
		for _, field := range fields {
			if field.Group != group.UUKey {
				continue
			}
			if v, ok := current[field.Field]; ok {
				parts[field.Field] = v
			} else if v, ok := current[field.GetKey()]; ok {
				parts[field.Field] = v
			}

		}
		// requestValue need be clean data
		if val, ok := flatten[group.UUKey]; ok {
			request := map[string]any{}
			if get, _ := val.(object); ok {
				request = get
			}
			for _, field := range fields {
				if field.Group != group.UUKey {
					continue
				}
				// 处理日期时间字段, 用于按期/周字段fill to prepare
				if field.FType == consts.FTYPE_DATETIME {
					find, ok := request[field.Field]
					if ok && field.HasTimeTerm() {
						expand = &field
						dateTime = self.asTime(find)
					}
				}
				// 禁用字段不能修改
				switch field.Extra.Disabled {
				case consts.DISABLED_ALWAYS:
					continue
				case consts.DISABLED_UPSERT:
					continue
				}
				if v, ok := request[field.Field]; ok {
					parts[field.Field] = v
				}
			}
		}
		// 合并group data into result objects
		if group.GType == consts.GTYPE_GROUPED {
			result[group.UUKey] = parts
		} else {
			result = maputil.Merge(result, parts)
		}
	}

	// 按期,按周字段fill to prepare
	if expand == nil || !expand.HasTimeTerm() {
		return result
	}
	expands := self.ExpandTime(skma, dateTime)
	for key, value := range expands {
		result[key] = value
	}
	return result
}

func (self *Record) asTime(value any) *time.Time {
	switch v := value.(type) {
	case *time.Time:
		return v
	case time.Time:
		t := v
		return &t
	case string:
		return support.ParseDate(v)
	default:
		return nil
	}
}

func (self *Record) ToChanges(skma schema.ISchema, flatten object) object {
	changes, ignores := object{}, []string{
		"basic.updated_at", "basic.updated_by",
	}
	// flatten := self.ToFlatten(skma, self.Storage)
	for key, get := range flatten {
		if field := skma.GetField(key); field == nil {
			continue
		} else if slice.Contain(ignores, key) {
			continue
		} else if old, ok := self.Current[key]; !ok {
			changes[key] = get
		} else if !field.Equal(old, get) {
			changes[key] = get
		}
	}
	return changes
}

func (self *Record) ToIndexed(skma schema.ISchema) object {
	data := self.Storage

	result := object{}
	groups := skma.GetGroups()
	fields := skma.GetFields()
	for _, group := range groups {
		for _, field := range fields {
			if field.Index == "" {
				continue
			}
			if field.Group != group.UUKey {
				continue
			}

			value, ok := data[field.Group]
			if !ok || value == nil {
				continue
			}

			object, ok := value.(object)
			if !ok || object == nil {
				continue
			}
			result[field.Index] = object[field.Field]
		}
	}
	return result
}

// ExpandTime 根据 utime 解析出的时间，扩展出：
//   - uweek：ISO 周序号，YYYYWW，如 202552 表示 2025 年第 52 周；
//   - uterm：财务期间，YYYYMM，期间为「上月 start 日 ～ 本月 start-1 日」归属「本月」，
//     例如 start=26 时 12/26～1/25 属 1 月期间（202601）。
func (self *Record) ExpandTime(skma schema.ISchema, date *time.Time) object {
	var expand *schema.Field = nil
	for _, field := range skma.GetFields() {
		if field.FType == consts.FTYPE_DATETIME {
			if field.HasTimeTerm() {
				expand = &field
			}
		}
	}
	if expand == nil || !expand.HasTimeTerm() {
		return object{}
	}
	if date == nil {
		return object{"uweek": nil, "uterm": nil}
	}
	// var group object = data
	// if expand.GType != consts.GTYPE_FLATTEN {
	// 	if get, ok := data[expand.Group]; ok {
	// 		group, _ = get.(map[string]any)
	// 	}
	// }
	// // 获取时间
	// date, ok := group[expand.Field].(*time.Time)
	// if !ok || date == nil {
	// 	return object{"uweek": nil, "uterm": nil}
	// }
	uterm, uweek := expand.GetTermWeek(date)
	return object{"uweek": uweek, "uterm": uterm}
}

func (self *Record) GetUUKey(skma *schema.Input) string {
	parts := []string{}
	for _, key := range skma.Unique {
		if get, ok := self.Default[key]; ok {
			parts = append(parts, fmt.Sprint(get))
		}
	}
	return strings.Join(parts, ":")
}

func (self *Record) GetUnique(skma *schema.Input) object {
	result := object{}
	for _, key := range skma.Unique {
		if get, ok := self.Default[key]; ok {
			result[key] = get
		}
	}
	return result
}
