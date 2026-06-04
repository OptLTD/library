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

type InputParser struct {
	schema  *schema.Input
	source  *source.Value
	request *request.Record
	handles []ICallable
}

func (self *InputParser) Using(handle ICallable) {
	if self.handles == nil {
		self.handles = []ICallable{}
	}

	self.handles = append(self.handles, handle)
}

func (self *InputParser) Build(value *source.Value, request *request.Record) (any, error) {
	self.source = value
	self.request = self.PrepareReq(request)

	var match bool
	var input *source.Input
	model := self.BuildModel()
	inputs := self.BuildInputs()
	for _, item := range inputs {
		if item.UUKey == request.Using {
			input, match = &item, true
		}
	}
	if input == nil && !match {
		clicks := []string{"[SETUP][*]"}
		input = &source.Input{Clicks: clicks}
		keys := maputil.Keys(value.Inputs)
		support.LogWarnf(self.request.LogID,
			"Input not found %s of [%s], keys: %s",
			request.Using, request.Model, keys,
		)
	}

	fields := self.BuildFields(input)
	clicks := self.BuildClicks(input)
	xrules := self.BuildXRules(input)
	groups := maputil.Values(value.Groups)
	slice.SortByField(groups, "SeqNo")
	others := map[string]any{}
	if input != nil && len(input.Extra) > 0 {
		maps.Copy(others, input.Extra)
	}
	self.schema = &schema.Input{
		Model: model, Input: input, Title: input.Title,
		Groups: groups, Fields: fields, Clicks: clicks,
		XRules: xrules, Source: value, Request: request,
		Account: request.Login, Refers: map[string]any{},
		Others: others,
	}

	if input != nil {
		self.applyPreset(input)
		applyFieldViewPatch(
			self.schema.Fields,
			viewObject(input.Rename),
			viewObject(input.Replace),
		)
	}

	for _, handle := range self.handles {
		handle.InputSchema(self.schema)
	}
	return self.schema, nil
}

func (self *InputParser) applyPreset(input *source.Input) {
	if input == nil || len(input.Preset) == 0 {
		return
	}
	if self.request.UUKey != "" {
		return
	}
	if self.request.Value == nil {
		self.request.Value = map[string]any{}
	}
	for key, val := range input.Preset {
		if _, ok := self.request.Value[key]; !ok {
			self.request.Value[key] = val
		}
	}
}

func (self *InputParser) PrepareReq(req *request.Record) *request.Record {
	req.Using = support.Or(req.Using, "default")
	return req
}
func (self *InputParser) BuildModel() *source.Model {
	// model := self.source.Model
	// model.Model = self.request.Model
	// model.UUKey = self.request.UUKey
	return &self.source.Model
}

func (self *InputParser) BuildFields(input *source.Input) []source.Field {
	groups := self.source.Groups
	fields := self.source.Fields
	result := []source.Field{}

	for _, field := range fields {
		// group
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
		flag := self.CheckShown(field, input)
		if flag == consts.VISIBLE_NONE {
			continue
		}
		field.Shown = flag == consts.VISIBLE_SHOW
		result = append(result, field)
	}
	slice.SortByField(result, "SeqNo", "asc")
	return result
}

func (self *InputParser) BuildClicks(input *source.Input) []source.Click {
	clicks := self.source.Clicks
	result := []source.Click{}

	reqScene := self.request.Scene
	switch reqScene {
	case consts.ACTION_UPDATE:
		reqScene = consts.SCENE_INPUT
	case consts.ACTION_INSERT:
		reqScene = consts.SCENE_INPUT
	}
	var expandedKeys = ExpandClicks(
		input.Clicks, maputil.Keys(clicks),
	)
	for _, key := range expandedKeys {
		if act, ok := clicks[key]; ok {
			act.CType = strings.ToUpper(act.CType)
			act.CType = support.Or(act.CType, "BUTTON")
			act.Scene = support.Or(act.Scene, []string{
				consts.SCENE_DETAIL, consts.SCENE_INPUT,
			})
			if slices.Contains(act.Scene, reqScene) {
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

func (self *InputParser) BuildXRules(input *source.Input) []source.XRule {
	result := []source.XRule{}
	for _, key := range input.XRules {
		if rule, ok := self.source.XRules[key]; ok {
			result = append(result, rule)
		}
	}
	return result
}

func (self *InputParser) BuildInputs() []source.Input {
	inputs := self.source.Inputs
	result := []source.Input{}
	for key, input := range inputs {
		input.UUKey = key
		result = append(result, input)
	}
	return result
}

func (self *InputParser) CheckShown(field source.Field, input *source.Input) string {
	if field.Extra.Implicit {
		return consts.VISIBLE_HIDE
	}
	visible, hidden := false, false
	if len(input.Fields) == 0 && len(input.Hidden) == 0 {
		return consts.VISIBLE_SHOW
	} else if len(input.Fields) == 0 {
		visible = true
	}
	for _, item := range input.Fields {
		matched, err := regexp.Match(item, []byte(field.UUKey))
		if err != nil || !matched {
			continue
		}
		visible = true
	}
	if slice.Contain(input.Fields, ".*") {
		visible = true
	}
	for _, item := range input.Hidden {
		matched, err := regexp.Match(item, []byte(field.UUKey))
		if err != nil || !matched {
			continue
		}
		hidden = true
	}
	if !visible || hidden {
		return consts.VISIBLE_HIDE
	}
	return consts.VISIBLE_SHOW
}
