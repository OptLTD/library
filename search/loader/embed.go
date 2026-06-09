package loader

import (
	"context"
	"embed"
	"github.com/OptLTD/library/search/consts"
	"github.com/OptLTD/library/search/source"
	"github.com/OptLTD/library/search/support"
	"strings"

	parser "github.com/buger/jsonparser"
)

func joinEmbedPath(base, rel string) string {
	base = strings.TrimSuffix(strings.Trim(base, "/"), "/")
	rel = strings.TrimPrefix(strings.Trim(rel, "/"), "/")
	return base + "/" + rel
}

type EmbedLoader struct {
	JSONLoader
	fs *embed.FS
}

func (self *EmbedLoader) Init() error {
	value, ok := support.GetValue(consts.OBJECT_EMBEDFS)
	if !ok || !support.Bool(value) {
		return support.ErrorConfigClient
	}

	self.fs = value.(*embed.FS)
	return nil
}

func (self *EmbedLoader) Load(ctx context.Context, name string) (*source.Value, error) {
	if err := self.Init(); err != nil {
		logID := GetLogID(ctx)
		support.LogError(logID, "Load error", err)
		return nil, err
	}
	if base := GetBase(ctx); base != "" {
		self.base = strings.Trim(base, "/")
	}
	file := joinEmbedPath(self.base, ENTRY_JSON)
	strValue, err := self.fs.ReadFile(file)
	if err != nil {
		return nil, support.EntryNotExsit
	}

	value := string(strValue)
	config, err := self.loadFile(value, name, "config")
	if err != nil {
		return nil, support.ConfigNotExsit
	}

	source, err := self.JSONLoader.loadSchemaWithExtends(ctx, self, name, config)
	if err != nil {
		return source, err
	}

	// 处理tables
	tables, err := self.loadFile(value, name, "tables")
	if err == nil {
		source.Tables = self.parseTables(tables)
	}

	// 处理inputs
	inputs, err := self.loadFile(value, name, "inputs")
	if err == nil {
		source.Inputs = self.parseInputs(inputs)
	}

	// 处理xrules
	xrules, err := self.loadFile(value, name, "xrules")
	if err == nil {
		source.XRules = self.parseXRules(xrules)
	}

	return source, nil
}

func (self *EmbedLoader) getEntry(entry string) (string, error) {
	file := joinEmbedPath(self.base, entry)
	strValue, err := self.fs.ReadFile(file)
	if err != nil {
		return "", support.EntryNotExsit
	}
	return string(strValue), nil
}

func (self *EmbedLoader) loadFile(value string, name string, part string) (string, error) {
	// 读取文件内容
	file, err := parser.GetString([]byte(value), name, part)
	if err != nil {
		return "", err
	}

	file = joinEmbedPath(self.base, file)
	if json, err := self.fs.ReadFile(file); err != nil {
		return "", err
	} else {
		return string(json), nil
	}
}
