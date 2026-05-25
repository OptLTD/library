package loader

import (
	"context"
	"search/consts"
	"search/source"
	"strings"
)

type SchemaTableName struct {
	Model string
	Table string
	Input string
	XRule string

	Field string
	Group string
}

type ILoader interface {
	Load(ctx context.Context, model string) (*source.Value, error)
}

func NewLoader(name string) ILoader {
	switch strings.ToUpper(name) {
	case consts.LOADER_MYSQL:
		return &MySQLLoader{}
	case consts.LOADER_MONGO:
		return &MongoLoader{}
	case consts.LOADER_EMBED:
		return &EmbedLoader{}
	case consts.LOADER_JSON:
		return &JSONLoader{}
	default:
		return &JSONLoader{}
	}
}

func ResetDefaultGroup(groups map[string]source.Group) source.Group {
	if basic, ok := groups["basic"]; ok {
		basic.SeqNo = 0
		return basic
	}
	return source.Group{
		UUKey: "basic",
		GType: "FLATTEN",
		Title: "基础信息",
	}
}

func getDefaultGroup() source.Group {
	return source.Group{
		UUKey: "basic",
		GType: "FLATTEN",
		Title: "基础信息",
	}
}
