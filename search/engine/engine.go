package engine

import (
	"github.com/OptLTD/library/search/respond"
	"github.com/OptLTD/library/search/schema"
)

type IEngine interface {
	Using(ICallable) IEngine
	First(*schema.Input, *respond.Record) error
	Store(*schema.Input, *respond.Record) error
	Update(*schema.Input, map[string]any) error
	Upsert(*schema.Input, []*respond.Record) error
	Select(*schema.Input, []*respond.Record) error
	Search(*schema.Table) (*respond.Result, error)
	Digest(*schema.Digest) (*respond.Result, error)
}

type ICallable interface {
	BeforeUpsert(*schema.Input, *respond.Record) error
	HandleUpsert(*schema.Input, *respond.Record) error
	SearchResult(*schema.Table, *respond.Result) error
	DigestResult(*schema.Digest, *respond.Result) error
}
