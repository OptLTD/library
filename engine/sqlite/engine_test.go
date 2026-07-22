package sqlite_test

import (
	"testing"

	enginesqlite "github.com/OptLTD/library/engine/sqlite"
	"github.com/OptLTD/library/search/consts"
	"github.com/OptLTD/library/search/request"
	"github.com/OptLTD/library/search/schema"
	"github.com/OptLTD/library/search/source"
	glebarezsqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := "file:" + t.Name() + "?mode=memory&cache=shared"
	db, err := gorm.Open(glebarezsqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	ddl := `CREATE TABLE demo (
		uukey TEXT PRIMARY KEY,
		model TEXT,
		state INTEGER NOT NULL DEFAULT 1,
		name TEXT,
		amount REAL,
		income TEXT
	)`
	if err := db.Exec(ddl).Error; err != nil {
		t.Fatal(err)
	}
	rows := []map[string]any{
		{"uukey": "A1", "model": "demo", "state": 1, "name": "alpha", "amount": 10, "income": `{"freight":100}`},
		{"uukey": "A2", "model": "demo", "state": 1, "name": "beta", "amount": 20, "income": `{"freight":200}`},
		{"uukey": "A3", "model": "demo", "state": 1, "name": "", "amount": 5, "income": `{"freight":50}`},
	}
	for _, row := range rows {
		if err := db.Table("demo").Create(row).Error; err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func demoTable(query map[string]any, order *request.Order) *schema.Table {
	return &schema.Table{
		Model: &source.Model{Source: "demo", Search: "demo"},
		Table: &source.Table{},
		Request: &request.Search{
			Model: "demo.model",
			Page:  1,
			Size:  50,
			Query: query,
			Order: order,
		},
		Scope: map[string]any{"state": consts.STATE_NORMAL},
		Fields: []source.Field{
			{UUKey: "basic.uukey", Field: "uukey", Group: "basic", GType: consts.GTYPE_FLATTEN},
			{UUKey: "basic.name", Field: "name", Group: "basic", GType: consts.GTYPE_FLATTEN, FType: consts.FTYPE_SUBJECT},
			{UUKey: "basic.amount", Field: "amount", Group: "basic", GType: consts.GTYPE_FLATTEN, FType: consts.FTYPE_NUMERIC},
			{UUKey: "income.freight", Field: "freight", Group: "income", GType: consts.GTYPE_GROUPED, FType: consts.FTYPE_NUMERIC},
		},
		Groups: []source.Group{
			{UUKey: "basic", GType: consts.GTYPE_FLATTEN},
			{UUKey: "income", GType: consts.GTYPE_GROUPED},
		},
	}
}

func TestSearchOrderByAmountDesc(t *testing.T) {
	db := testDB(t)
	eng := enginesqlite.NewEngine(db)
	skma := demoTable(nil, &request.Order{Field: "basic.amount", Order: "desc"})
	res, err := eng.Search(skma)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Values) != 3 {
		t.Fatalf("count %d", len(res.Values))
	}
	first := res.Values[0]
	if first["basic.uukey"] != "A2" {
		t.Fatalf("first row uukey %#v", first)
	}
	freight := first["income.freight"]
	if freight != float64(200) && freight != int64(200) {
		t.Fatalf("first row %#v", first)
	}
}

func TestSearchFilterNEAndGroupedJSON(t *testing.T) {
	db := testDB(t)
	eng := enginesqlite.NewEngine(db)
	skma := demoTable(map[string]any{
		"basic.name:NE":      "alpha",
		"income.freight:GT": 80,
	}, nil)
	res, err := eng.Search(skma)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Values) != 1 {
		t.Fatalf("got len=%d %#v", len(res.Values), res.Values)
	}
	freight := res.Values[0]["income.freight"]
	if freight != float64(200) && freight != int64(200) {
		t.Fatalf("got %#v", res.Values)
	}
}

func TestSearchFilterNIL(t *testing.T) {
	db := testDB(t)
	eng := enginesqlite.NewEngine(db)
	skma := demoTable(map[string]any{"basic.name:NIL": true}, nil)
	res, err := eng.Search(skma)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Values) != 1 {
		t.Fatalf("got %#v", res.Values)
	}
	freight := res.Values[0]["income.freight"]
	if freight != float64(50) && freight != int64(50) {
		t.Fatalf("got %#v", res.Values)
	}
}

func TestDigestSumByName(t *testing.T) {
	db := testDB(t)
	eng := enginesqlite.NewEngine(db)
	table := demoTable(nil, nil)
	table.Others = map[string]any{
		"group": []any{"basic.name|ASC|名称"},
		"count": []any{"basic.amount|SUM|合计"},
	}
	digest := table.BuildDigest()
	res, err := eng.Digest(digest)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Values) < 2 {
		t.Fatalf("values %#v", res.Values)
	}
	found := false
	for _, row := range res.Values {
		if row["basic.name"] == "alpha" {
			found = true
			if row["合计"] != float64(10) && row["合计"] != int64(10) {
				t.Fatalf("sum %#v", row["合计"])
			}
		}
	}
	if !found {
		t.Fatal("missing alpha group")
	}
}
