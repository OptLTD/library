package parser

import (
	"regexp"
	"search/consts"
	"search/request"
	"search/schema"
	"search/source"
	"search/support"
	"strings"

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
		input = &source.Input{}
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
	self.schema = &schema.Input{
		Model: model, Input: input, Title: input.Title,
		Groups: groups, Fields: fields, Clicks: clicks,
		XRules: xrules, Source: value, Request: request,
		Account: request.Login, Refers: map[string]any{},
	}

	for _, handle := range self.handles {
		handle.InputSchema(self.schema)
	}
	return self.schema, nil
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
	result := []source.Click{}
	for key, click := range self.source.Clicks {
		click.CType = support.Or(click.CType, "ACTION")
		click.CType = strings.ToUpper(click.CType)
		if slice.Contain(input.Clicks, key) {
			result = append(result, click)
		}
	}
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
	if len(input.Fields) == 0 {
		return consts.VISIBLE_SHOW
	}
	if slice.Contain(input.Fields, "*") {
		return consts.VISIBLE_SHOW
	}

	visible := false
	for _, item := range input.Fields {
		matched, err := regexp.Match(item, []byte(field.UUKey))
		if err != nil || matched == false {
			continue
		}
		visible = true
	}
	return support.If(visible, consts.VISIBLE_SHOW, consts.VISIBLE_HIDE)
}
