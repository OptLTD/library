package parser

import (
	"maps"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/OptLTD/library/search/consts"
	"github.com/OptLTD/library/search/request"
	"github.com/OptLTD/library/search/schema"
	"github.com/OptLTD/library/search/source"
	"github.com/OptLTD/library/search/support"

	"github.com/duke-git/lancet/v2/maputil"
	"github.com/duke-git/lancet/v2/slice"
)

type TableParser struct {
	schema  *schema.Table
	source  *source.Value
	request *request.Search
	handles []ICallable
}

func (self *TableParser) Using(handle ICallable) {
	if self.handles == nil {
		self.handles = []ICallable{}
	}

	self.handles = append(self.handles, handle)
}

func (self *TableParser) Build(value *source.Value, request *request.Search) (*schema.Table, error) {
	self.source = value
	self.request = self.PrepareReq(request)

	var match bool
	var table *source.Table
	model := self.BuildModel()
	tables := self.BuildTables()
	for _, item := range tables {
		if item.UUKey == request.Using {
			table, match = &item, true
		}
	}
	if table == nil && !match {
		clicks := []string{"[SETUP][*]"}
		table = &source.Table{Clicks: clicks}
		keys := maputil.Keys(value.Tables)
		support.LogDebugf(self.request.LogID,
			"Table not found %s of [%s], keys: %s",
			request.Using, request.Model, keys,
		)
	}

	fields := self.BuildFields(table)
	clicks := self.BuildClicks(table)
	groups := maputil.Values(value.Groups)
	slice.SortByField(groups, "SeqNo")
	refers := map[string]any{}
	others := map[string]any{}
	if request.Extra != nil {
		// 合并 request.Extra 到 others
		// 主要为了digest场景下带入分组求和信息
		maps.Copy(others, request.Extra)
	}
	if table != nil && len(table.Extra) > 0 {
		maps.Copy(others, table.Extra)
	}
	self.schema = &schema.Table{
		Model: model, Table: table, Query: request.Query,
		Groups: groups, Fields: fields, Scene: value.Scene,
		Request: request, Others: others, Source: value,
		Account: request.Login, Refers: refers, Clicks: clicks,
	}

	for _, handle := range self.handles {
		handle.SearchSchema(self.schema)
	}
	return self.schema, nil
}

func (self *TableParser) PrepareReq(req *request.Search) *request.Search {
	req.Page = support.Or(req.Page, 1)
	req.Size = support.Or(req.Size, 100)
	req.Using = support.Or(req.Using, "default")
	req.Order = support.Or(req.Order, &request.Order{
		Field: "basic.utime", Order: "desc",
	})
	return req
}

func (self *TableParser) BuildModel() *source.Model {
	// model := self.source.Model
	// model.Model = self.request.Model
	// model.UUKey = self.request.UUKey
	return &self.source.Model
}

func (self *TableParser) BuildFields(table *source.Table) []source.Field {
	groups := self.source.Groups
	fields := self.source.Fields
	result := []source.Field{}

	for _, field := range fields {
		group, ok := groups[field.Group]
		if !ok || !support.Bool(group) {
			continue
		}
		if !strings.EqualFold(field.Group, group.UUKey) {
			continue
		}
		// fill info
		field.GName = group.Title
		field.GType = strings.ToUpper(group.GType)
		// field.UUKey = strings.ToLower(field.UUKey)
		if field.Extra.Disabled == consts.DISABLED_EMPTY {
			field.Extra.Disabled = consts.DISABLED_NEVER
		}

		// check flag
		flag := self.CheckShown(field, table)
		if flag == consts.VISIBLE_NONE {
			continue
		}

		field.Shown = flag == consts.VISIBLE_SHOW
		// idx := slice.LastIndexOf(table.Orders, field.UUKey)
		// field.SeqNo = support.If(idx == -1, field.SeqNo, int64(idx))
		result = append(result, field)
	}
	slice.SortByField(result, "SeqNo", "asc")
	return result
}

func (self *TableParser) BuildClicks(table *source.Table) []source.Click {
	clicks := self.source.Clicks
	result := []source.Click{}

	// [NAME][*] => 从 source.Clicks 中收集所有以 [NAME] 为前缀的 key，作为展开结果
	var expandedKeys []string
	for _, key := range table.Clicks {
		if !strings.HasSuffix(key, "[*]") {
			expandedKeys = append(expandedKeys, key)
			continue
		}

		var matched []string
		prefix := strings.TrimSuffix(key, "[*]")
		for mapKey := range clicks {
			if strings.HasPrefix(mapKey, prefix) {
				matched = append(matched, mapKey)
			}
		}
		expandedKeys = append(expandedKeys, matched...)
	}
	for _, key := range expandedKeys {
		if act, ok := clicks[key]; ok {
			act.CType = strings.ToUpper(act.CType)
			act.CType = support.Or(act.CType, "BUTTON")
			act.Scene = support.Or(act.Scene, []string{
				consts.SCENE_DETAIL, consts.SCENE_SEARCH,
			})
			if slices.Contains(act.Scene, self.request.Scene) {
				result = append(result, act)
			}
		}
	}

	// SeqNo + UUKey 双关键字排序，保证 clicks 展开结果顺序稳定
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].SeqNo != result[j].SeqNo {
			return result[i].SeqNo < result[j].SeqNo
		}
		return result[i].UUKey < result[j].UUKey
	})
	return result
}

func (self *TableParser) BuildTables() []source.Table {
	tables := self.source.Tables
	result := []source.Table{}
	for key, table := range tables {
		table.UUKey = key
		result = append(result, table)
	}
	return result
}

func (self *TableParser) CheckShown(field source.Field, table *source.Table) string {
	visible, hidden := false, false
	if len(table.Fields) == 0 && len(table.Hidden) == 0 {
		return consts.VISIBLE_SHOW
	} else if len(table.Fields) == 0 {
		visible = true
	}

	for _, item := range table.Fields {
		matched, err := regexp.Match(item, []byte(field.UUKey))
		if err != nil || matched == false {
			continue
		}
		visible = true
	}
	// 系统字段不过滤,  配置之外, 过滤掉
	// @2026-01-15 output时才应该过滤, 现在不过滤
	if field.Extra.Embedded && !visible {
		// return consts.VISIBLE_NONE
	}

	for _, item := range table.Hidden {
		matched, err := regexp.Match(item, []byte(field.UUKey))
		if err != nil || matched == false {
			continue
		}
		hidden = true
	}
	if !visible || hidden {
		return consts.VISIBLE_HIDE
	}
	return consts.VISIBLE_SHOW
}
