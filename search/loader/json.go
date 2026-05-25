package loader

import (
	"context"
	"encoding/json"
	"search/consts"
	"search/source"
	"search/support"
	"strings"

	parser "github.com/buger/jsonparser"
	"github.com/duke-git/lancet/v2/fileutil"
)

const ENTRY_JSON = "entry.json"

type JSONLoader struct {
	base string
}

func (self *JSONLoader) Init() error {
	return nil
}
func (self *JSONLoader) Load(ctx context.Context, name string) (*source.Value, error) {
	logID := GetLogID(ctx)
	if err := self.Init(); err != nil {
		support.LogError(logID, "Load error", err)
		return nil, err
	}

	if base := GetBase(ctx); base != "" {
		self.base = strings.ReplaceAll(base+"/", "//", "/")
	}
	entry, err := self.getEntry(ENTRY_JSON)
	if err != nil {
		return nil, support.EntryNotExsit
	}

	configParse, err := self.loadFile(entry, name, "config")
	if err != nil {
		return nil, support.ConfigNotExsit
	}

	value, err := self.parseSource(name, configParse)
	if err != nil {
		return value, err
	}

	// 处理tables
	tables, err := self.loadFile(entry, name, "tables")
	if err == nil {
		value.Tables = self.parseTables(tables)
	}

	// 处理inputs
	inputs, err := self.loadFile(entry, name, "inputs")
	if err == nil {
		value.Inputs = self.parseInputs(inputs)
	}

	// 处理xrules
	xrules, err := self.loadFile(entry, name, "xrules")
	if err == nil {
		value.XRules = self.parseXRules(xrules)
	}

	return value, nil
}

func (self *JSONLoader) parseSource(name string, configJson string) (*source.Value, error) {
	config := []byte(configJson)
	// 读取文件内容
	driver, err := parser.GetString(config, "model", "driver")
	value, err := parser.GetString(config, "model", "source")
	if value == "" || err != nil {
		return nil, support.ErrorModelSource
	}
	// 读取文件内容
	search, err := parser.GetString(config, "model", "search")
	if search == "" || err != nil {
		search = value
	}

	// 处理group
	idx, groups := uint16(0), map[string]source.Group{}
	parser.ObjectEach(config, func(key []byte, value []byte, dataType parser.ValueType, offset int) error {
		seqno, _ := parser.GetInt(value, "seqno")
		title, _ := parser.GetString(value, "title")
		gtype, _ := parser.GetString(value, "gtype")

		extraValue := source.GExtra{}
		extra, vtype, _, err := parser.Get(value, "extra")
		if err == nil && vtype.String() == "object" {
			// unmarshal extra if exist
			json.Unmarshal(extra, &extraValue)
		}
		idx, groups[string(key)] = idx+1, source.Group{
			Model: name, UUKey: string(key),
			Title: title, Extra: extraValue,
			GType: strings.ToUpper(gtype),
			SeqNo: support.Or(uint16(seqno), idx),
		}
		return nil
	}, "groups")

	// 处理action
	clicks := map[string]source.Click{}
	parser.ObjectEach(config, func(key []byte, value []byte, dataType parser.ValueType, offset int) error {
		click := source.Click{UUKey: string(key)}
		if json.Unmarshal(value, &click) == nil {
			clicks[string(key)] = click
		}
		return nil
	}, "clicks")

	// reset default group, seqno to 0
	groups["basic"] = ResetDefaultGroup(groups)
	idx, fields := uint16(0), map[string]source.Field{}
	for key, item := range groups {
		gname, gtype := item.Title, item.GType
		parser.ObjectEach(config, func(key []byte, value []byte, dataType parser.ValueType, offset int) error {
			theField := self.parseField(value, item)
			theUUKey := item.UUKey + "." + string(key)
			theField.Field, theField.UUKey = string(key), theUUKey
			theField.GName, theField.GType = gname, gtype
			theField.SeqNo = theField.SeqNo + idx
			switch strings.ToUpper(item.GType) {
			case consts.GTYPE_FLATTEN:
				theField.Index = string(key)
			case consts.GTYPE_GROUPED:
				theField.Index = item.UUKey
			}
			fields[theUUKey], idx = theField, idx+1
			return nil
		}, "fields", key)

		parser.ObjectEach(config, func(key []byte, value []byte, dataType parser.ValueType, offset int) error {
			prefix, uukey := item.UUKey+".", string(key)
			if !strings.HasPrefix(uukey, prefix) {
				return nil // 不匹配
			}
			theField := self.parseField(value, item)
			field, _ := strings.CutPrefix(uukey, prefix)
			theField.Field, theField.UUKey = field, uukey
			theField.GName, theField.GType = gname, gtype
			theField.SeqNo = theField.SeqNo + idx
			switch strings.ToUpper(item.GType) {
			case consts.GTYPE_FLATTEN:
				theField.Index = string(key)
			case consts.GTYPE_GROUPED:
				theField.Index = item.UUKey
			}
			fields[uukey], idx = theField, idx+1
			return nil
		}, "fields")
	}

	model := source.Model{
		UUKey:  name,
		Source: value, Search: search,
		Driver: strings.ToUpper(driver),
	}

	// 读取文件内容
	model.Title, err = parser.GetString(config, "model", "title")
	model.Brief, err = parser.GetString(config, "model", "brief")
	extra, vtype, _, err := parser.Get(config, "model", "extra")
	if err == nil && vtype.String() == "object" {
		json.Unmarshal(extra, &model.Extra)
	}

	return &source.Value{
		Model: model, Clicks: clicks,
		Fields: fields, Groups: groups,
	}, nil
}

func (self *JSONLoader) parseField(value []byte, item source.Group) source.Field {
	label, _ := parser.GetString(value, "label")
	ftype, _ := parser.GetString(value, "ftype")
	index, _ := parser.GetString(value, "index")
	using, err := parser.GetString(value, "using")
	width, err := parser.GetInt(value, "width")
	seqno, _ := parser.GetInt(value, "seqno")
	if err != nil {
		width = 80
	}
	var refer *source.Refer
	if using != "" {
		refer = new(source.Refer)
		refer = refer.Parse(using)
	}
	extraValue := source.FExtra{}
	index = support.If(index, index, "")
	extra, vtype, _, err := parser.Get(value, "extra")
	if err == nil && vtype.String() == "object" {
		if json.Unmarshal(extra, &extraValue) != nil {
			// todo
		}
	}
	options, vtype, _, err := parser.Get(value, "options")
	if err == nil && vtype.String() == "array" {
		json.Unmarshal(options, &extraValue.Options)
	}

	return source.Field{
		Group: item.UUKey, Index: index, Label: label,
		Width: uint16(width), Refer: refer, FType: ftype,
		Extra: extraValue, SeqNo: support.Or(
			uint16(seqno), item.SeqNo*1000,
		),
	}
}

func (self *JSONLoader) parseTables(tablesParse string) map[string]source.Table {
	// 处理tables
	tables := map[string]source.Table{
		"default": {Title: "default", UUKey: "default"},
	}
	parser.ObjectEach([]byte(tablesParse), func(key []byte, value []byte, dataType parser.ValueType, offset int) error {
		title, _ := parser.GetString(value, "title")
		fields, _ := parser.GetString(value, "fields")
		table := source.Table{Title: title, UUKey: string(key)}

		table.Query = map[string]any{}
		err := parser.ObjectEach([]byte(value), func(k []byte, val []byte, dataType parser.ValueType, offset int) error {
			table.Query[string(k)] = support.GetVal(val, dataType)
			return nil
		}, "query")
		// 显示字段
		_, err = parser.ArrayEach([]byte(value), func(val []byte, dataType parser.ValueType, offset int, err error) {
			table.Fields = append(table.Fields, string(val))
		}, "fields")
		if err != nil && fields == "null" {
			// 在闭包中无法直接获取 logID，使用空字符串
			support.LogWarn("", "field error", err)
		}

		// action 处理
		parser.ArrayEach([]byte(value), func(val []byte, dataType parser.ValueType, offset int, err error) {
			table.Clicks = append(table.Clicks, string(val))
		}, "clicks")

		// 隐藏字段处理
		parser.ArrayEach([]byte(value), func(val []byte, dataType parser.ValueType, offset int, err error) {
			table.Hidden = append(table.Hidden, string(val))
		}, "hidden")
		// 聚合字段
		// _, err = parser.ArrayEach([]byte(value), func(val []byte, dataType parser.ValueType, offset int, err error) {
		// 	table.AggrBy = append(table.AggrBy, string(val))
		// }, "aggrby")
		// filters 处理
		// parser.ArrayEach([]byte(value), func(val []byte, dataType parser.ValueType, offset int, err error) {
		// 	table.Filters = append(table.Filters, string(val))
		// }, "filters")

		tables[string(key)] = table
		return nil
	})
	return tables
}
func (self *JSONLoader) parseInputs(inputsParse string) map[string]source.Input {
	// 处理 inputs
	inputs := map[string]source.Input{}
	parser.ObjectEach([]byte(inputsParse), func(key []byte, value []byte, dataType parser.ValueType, offset int) error {
		title, _ := parser.GetString(value, "title")
		input := source.Input{Title: title, UUKey: string(key)}
		// 显示字段
		fields := []string{}
		_, err := parser.ArrayEach([]byte(value), func(val []byte, dataType parser.ValueType, offset int, err error) {
			fields = append(fields, string(val))
		}, "fields")
		if err != nil {
			return support.ErrorInputField
		}
		input.Fields = fields

		// button 处理
		clicks := []string{}
		parser.ArrayEach([]byte(value), func(val []byte, dataType parser.ValueType, offset int, err error) {
			clicks = append(clicks, string(val))
		}, "clicks")
		input.Clicks = clicks

		// action 处理
		xrules := []string{}
		parser.ArrayEach([]byte(value), func(val []byte, dataType parser.ValueType, offset int, err error) {
			xrules = append(xrules, string(val))
		}, "xrules")
		input.XRules = xrules

		inputs[string(key)] = input
		return nil
	})
	return inputs
}

func (self *JSONLoader) parseXRules(xrulesParse string) map[string]source.XRule {
	// 处理 inputs
	xrules := map[string]source.XRule{}
	parser.ObjectEach([]byte(xrulesParse), func(key []byte, value []byte, dataType parser.ValueType, offset int) error {
		uukey, xrule := string(key), new(source.XRule)
		if err := json.Unmarshal(value, xrule); err != nil {
			return err
		}
		xrules[uukey] = *xrule
		return nil
	})
	return xrules
}

func (self *JSONLoader) parseModel(refstr string) source.Refer {
	refer := source.Refer{}
	parts := strings.Split(refstr, "@")
	if len(parts) == 1 {
		refer.Using = refstr
		refer.KeyBy = "uukey"
		refer.TxtBy = "label"
		return refer
	}
	fields := parts[0]
	refer.Using = parts[1]
	parts = strings.Split(fields, ",")
	if len(parts) == 1 {
		refer.KeyBy = "uukey"
		refer.TxtBy = parts[0]
	} else {
		refer.TxtBy = parts[1]
		refer.KeyBy = parts[0]
	}
	return refer
}

func (self *JSONLoader) getEntry(entry string) (string, error) {
	file := self.base + entry
	if !fileutil.IsExist(file) {
		return "", support.EntryNotExsit
	}
	return fileutil.ReadFileToString(file)
}

func (self *JSONLoader) loadFile(value string, name string, part string) (string, error) {
	// 读取文件内容
	file, err := parser.GetString([]byte(value), name, part)
	if err != nil {
		return "", err
	}
	file = self.base + file
	if !fileutil.IsExist(file) {
		return "", support.ConfigNotExsit
	}
	json, err := fileutil.ReadFileToString(file)
	if err != nil {
		return "", err
	}
	return json, nil
}
