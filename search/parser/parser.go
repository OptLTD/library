package parser

import (
	"search/request"
	"search/schema"
	"search/source"
)

type IParser interface {
	Using(ICallable)
	Build(*source.Value, *request.Search, *request.Account) (*any, error)
}

type ICallable interface {
	InputSchema(*schema.Input)
	SearchSchema(*schema.Table)
}

func NewTableParser(handles []ICallable) *TableParser {
	instance := &TableParser{}
	for _, handle := range handles {
		instance.Using(handle)
	}
	return instance
}

func NewInputParser(handles []ICallable) *InputParser {
	instance := &InputParser{}
	for _, handle := range handles {
		instance.Using(handle)
	}
	return instance
}
