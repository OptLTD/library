package engine

import (
	"github.com/OptLTD/library/search/consts"
	"github.com/OptLTD/library/search/respond"
	"github.com/OptLTD/library/search/schema"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
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

func NewEngine(name string, client any) IEngine {
	var engine IEngine
	switch strings.ToUpper(name) {
	case consts.SEARCH_DBMYSQL:
		session := &gorm.Session{QueryFields: true}
		engine = &MySQLEngine{
			client: client.(*gorm.DB).Session(session),
		}
	case consts.SEARCH_MONGODB:
		engine = &MongoEngine{
			client: client.(*mongo.Database),
		}
	case consts.SEARCH_ELASTIC:
		engine = &ElasticEngine{
			client: client.(*elasticsearch.Client),
		}
	case consts.SEARCH_MEMORY:
		engine = NewMemoryEngine()
	default:
		engine = &ElasticEngine{
			client: client.(*elasticsearch.Client),
		}
	}
	return engine
}
