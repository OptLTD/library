package sqlite

import (
	"testing"

	"github.com/OptLTD/library/search/consts"
	"github.com/OptLTD/library/search/request"
	"github.com/OptLTD/library/search/schema"
	"github.com/OptLTD/library/search/source"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSampleSkipsUnknownFields(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE TABLE app_vms_maintain (
		uukey TEXT PRIMARY KEY, state INTEGER, utime TEXT, title TEXT, vehicle TEXT
	)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO app_vms_maintain(uukey,state,utime,title,vehicle)
		VALUES ('M1',1,'2026-08-01','patchTire','V1')`).Error; err != nil {
		t.Fatal(err)
	}

	eng := NewEngine(db).(*Engine)
	table := &schema.Table{
		Model:   &source.Model{Search: "app_vms_maintain", Source: "app_vms_maintain"},
		Table:   &source.Table{},
		Request: &request.Search{},
		Fields: []schema.Field{
			{UUKey: "basic.title", Group: "basic", Field: "title", GType: consts.GTYPE_FLATTEN, Shown: true},
			{UUKey: "basic.vehicle", Group: "basic", Field: "vehicle", GType: consts.GTYPE_FLATTEN, Shown: true},
		},
		Query: map[string]any{},
	}
	ApplySearchIndex(table)

	out, err := eng.Sample(table, []string{"basic.title", "basic.driver", "extra.driver"}, 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(out["basic.title"]) != 1 || out["basic.title"][0] != "patchTire" {
		t.Fatalf("title sample=%v", out["basic.title"])
	}
	if len(out["basic.driver"]) != 0 {
		t.Fatalf("unknown field should be empty, got %v", out["basic.driver"])
	}
	if len(out["extra.driver"]) != 0 {
		t.Fatalf("unknown field should be empty, got %v", out["extra.driver"])
	}
}
