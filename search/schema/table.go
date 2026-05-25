package schema

import (
	"fmt"
	"search/consts"
	"search/request"
	"search/source"
	"search/support"
	"strings"

	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/maputil"
	"github.com/duke-git/lancet/v2/slice"
)

type Model = source.Model
type Click = source.Click
type Group = source.Group
type Field = source.Field
type Refer = source.Refer
type XRule = source.XRule
type Option = source.Option
type FExtra = source.FExtra
type GExtra = source.GExtra
type Source = source.Value
type SInput = source.Input
type STable = source.Table
type SortBy = source.Order
type GroupBy = source.GroupBy
type CountFN = source.CountFn
type PivotBy = source.PivotBy

// 聚合配置
type object = map[string]any

type Table struct {
	Scene string `json:"-"` // 构建来源(场景)

	Model *source.Model `json:"-"` // 模型信息
	Table *source.Table `json:"-"` // 表格信息
	// query scope
	Scope map[string]any `json:"-"` // 查询限制
	Query map[string]any `json:"-"` // 查询信息

	Sticky []string       `json:"sticky"` // 固定字段
	Groups []source.Group `json:"groups"` // 分组信息
	Fields []source.Field `json:"fields"` // 字段类型
	Clicks []source.Click `json:"clicks"` // 表格事件
	Refers map[string]any `json:"refers"` // 字典信息
	Others map[string]any `json:"others"` // 其他信息
	Source *source.Value  `json:"-"`      // 源数据

	// 传入信息
	Account *request.Account `json:"-"`       // 登陆用户
	Request *request.Search  `json:"request"` // 请求信息
}

func (self *Table) GetRequest() *request.Search {
	return self.Request
}

func (self *Table) GetSearch() *request.Search {
	if self.Request == nil {
		return nil
	}
	return convertor.DeepClone(self.Request)
}

func (self *Table) GetGroups() []source.Group {
	return self.Groups
}

func (self *Table) GetFields() []source.Field {
	return self.Fields
}

func (self *Table) GetRefers() []source.Refer {
	result := []source.Refer{}

	for _, group := range self.Groups {
		if group.Extra.Relation != "" {
			result = append(result, source.Refer{
				UUKey: group.Extra.Relation,
				Using: group.Extra.Relation,
				KeyBy: consts.FIELD_UUKEY,
			})
		}
	}

	for _, field := range self.Fields {
		if field.Refer == nil {
			continue
		}
		// optional,workflow 需要被排除
		if field.FType != consts.FTYPE_RELATION {
			continue
		}
		if field.Extra.Disabled == consts.DISABLED_ALWAYS {
			continue
		}
		// refer := ParseRefer(field.Refer)
		result = append(result, *field.Refer)
	}
	return result
}

func (self *Table) GetField(uukey string) *source.Field {
	for idx := range len(self.Fields) {
		if self.Fields[idx].GetKey() == uukey {
			return &self.Fields[idx]
		}
	}
	if self.Source == nil {
		return nil
	}
	if get, ok := self.Source.Fields[uukey]; ok {
		return &get
	}
	return nil
}

// format refer
func (self *Table) GetRefer(uukey string) *source.Refer {
	field := self.GetField(uukey)
	if field == nil {
		return nil
	}
	if !support.Bool(field.Extra) {
		return field.Refer
	}
	return field.GetRefer()
}

func (self *Table) BuildQuery() []Query {
	result := []Query{}
	requst := map[string]any{}
	if support.Bool(self.Query) {
		requst = self.Query
	}
	// 表格配置的查询条件
	if len(self.Table.Query) > 0 {
		keys := maputil.KeysBy(requst, func(key string) string {
			return strings.Split(key, ":")[0]
		})
		// 优先前端查询参数
		for key, value := range self.Table.Query {
			uukey := strings.Split(key, ":")[0]
			if !slice.Contain(keys, uukey) {
				requst[key] = value
			}
		}
	}
	// @todo 此处需要优化
	if len(self.Request.Query) > 0 {
		keys := maputil.KeysBy(requst, func(key string) string {
			return strings.Split(key, ":")[0]
		})
		// 优先前端查询参数
		for key, value := range self.Request.Query {
			uukey := strings.Split(key, ":")[0]
			if !slice.Contain(keys, uukey) {
				requst[key] = value
			}
		}
	}
	// 替换查询条件
	requst = BuildPreset(requst, self)
	query := map[string]any{}
	for key, value := range requst {
		parts := strings.Split(key, ":")
		if len(parts) == 1 {
			parts = append(parts, "")
		}
		uukey, logic := parts[0], parts[1]
		if slice.Contain(SubQuery, logic) {
			query[uukey+":"+logic] = value
			continue
		}
		if field := self.GetField(uukey); field == nil {
			continue
		} else if logic == "" {
			query[field.UUKey] = value
		} else {
			query[uukey+":"+logic] = value
		}
	}

	// set scope query
	logID := ""
	if self.Request != nil {
		logID = self.Request.LogID
	}
	if queries, err := BuildQuery(self.Scope); err == nil {
		result = append(result, queries...)
	} else {
		support.LogError(logID, "BuildQuery Wrong", err, query)
	}
	// set merged query
	if queries, err := BuildQuery(query); err == nil {
		result = append(result, queries...)
	} else {
		support.LogError(logID, "BuildQuery Wrong", err, query)
	}
	return result
}

// Get Subject Field for Schema
func (self *Table) GetSubject() *source.Field {
	for _, field := range self.Fields {
		if field.FType == consts.FTYPE_SUBJECT {
			return &field
		}
	}
	return self.GetField(consts.FIELD_UUKEY)
}

// SetDisplay for Relation Field
func (self *Table) SetDisplay(refer *source.Refer, display string) {
	refer.TxtBy = display
	for idx, field := range self.Fields {
		if field.FType != consts.FTYPE_RELATION {
			continue
		}
		if field.Extra.Relation != refer.Using {
			continue
		}
		self.Fields[idx].Refer = refer
		self.Fields[idx].Extra.Display = display
	}
}

func (self *Table) GetPlain(uukey string, value any) string {
	if !support.Bool(value) {
		return ""
	}
	field := self.GetField(uukey)
	if field == nil {
		return ""
	}
	switch field.FType {
	case consts.FTYPE_RELATION:
	case consts.FTYPE_OPTIONAL:
	case consts.FTYPE_WORKFLOW:
	default:
		return fmt.Sprint(value)
	}

	refer := field.GetRefer()
	if refer == nil {
		return fmt.Sprint(value)
	}

	list, _ := self.Refers[refer.Using]
	if refer.Using == "" || list == nil {
		return fmt.Sprint(value)
	}

	// 优先使用 Option 类型进行匹配
	if items, ok := list.([]Option); ok {
		for _, get := range items {
			if get.UUKey != value {
				continue
			}
			return get.Label
		}
		return fmt.Sprint(value)
	}

	// 否则使用 object 类型进行匹配
	if items, ok := list.([]object); ok {
		for _, item := range items {
			if get, ok := item[refer.KeyBy]; !ok {
				continue
			} else if get != value {
				continue
			}
			if get, ok := item[refer.TxtBy]; ok {
				return fmt.Sprint(get)
			}
		}
		return fmt.Sprint(value)
	}

	return fmt.Sprint(value)
}

// others.pivot=index|value|state
func (self *Table) BuildPivotBy() *PivotBy {
	result := &PivotBy{State: "none"}
	if item, ok := self.Others["pivot"]; ok {
		parts := strings.Split(item.(string), "|")
		switch len(parts) {
		case 3:
			result.Pivot = parts[0]
			result.Value = parts[1]
			result.State = parts[2]
		case 2:
			result.Pivot = parts[0]
			result.Value = parts[1]
		default:
			result.Pivot = parts[0]
		}
	}
	return result
}

// PivotByActive 透视是否开启。前端保存 on/off，关闭后 Pivot/Value 可能仍有值，必须以 State 为准。
func PivotByActive(p *PivotBy) bool {
	if p == nil {
		return false
	}
	if p.State == "" {
		return false
	}
	if p.State == "off" {
		return false
	}
	if p.State == "none" {
		return false
	}
	return true
}

// others.group=[]string{index|sort|format}
// 按车型分组:["basic.vehicle|ASC|车型"]
func (self *Table) BuildGroupBy() []GroupBy {
	group, _ := self.Others["group"].([]any)
	if len(group) == 0 {
		return nil
	}

	result := []GroupBy{}
	for _, item := range group {
		pivot := GroupBy{}
		parts := strings.Split(item.(string), "|")
		switch len(parts) {
		case 3:
			pivot.Index = parts[0]
			pivot.SortBy = parts[1]
			pivot.Format = parts[2]
		case 2:
			pivot.Index = parts[0]
			pivot.SortBy = parts[1]
		default:
			pivot.Index = parts[0]
		}
		result = append(result, pivot)
	}
	return result
}

// others.count=[]string{index|func|label}
// 金额求和:["basic.amount|SUM|合计金额"]
func (self *Table) BuildCountFn() []CountFN {
	aggrs, _ := self.Others["count"].([]any)
	if len(aggrs) == 0 {
		return nil
	}
	result := []CountFN{}
	for _, item := range aggrs {
		aggr := CountFN{}
		parts := strings.Split(item.(string), "|")
		switch len(parts) {
		case 3:
			aggr.Index = parts[0]
			aggr.Func = parts[1]
			aggr.Label = parts[2]
		case 2:
			aggr.Index = parts[0]
			aggr.Func = parts[1]
		default:
			aggr.Index = parts[0]
		}
		result = append(result, aggr)
	}
	return result
}

func (self *Table) BuildDigest() *Digest {
	result := &Digest{
		Table:  self,
		Fields: self.Fields,
		Groups: self.Groups,
		Sticky: self.Sticky,
	}
	result.PivotBy = self.BuildPivotBy()
	result.GroupBy = self.BuildGroupBy()
	result.CountFn = self.BuildCountFn()
	return result
}
