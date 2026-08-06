package sqlite

import (
	"testing"

	"github.com/OptLTD/library/search/consts"
	"github.com/OptLTD/library/search/schema"
	"github.com/OptLTD/library/search/source"
)

func TestFieldSQLExprLogicalPathUsesJSON(t *testing.T) {
	if got := fieldSQLExpr("income.total"); got != "json_extract(income, '$.total')" {
		t.Fatalf("grouped: %s", got)
	}
	if got := fieldSQLExpr("basic.utime"); got != "utime" {
		t.Fatalf("basic: %s", got)
	}
	if got := fieldSQLExpr("gen_income_total"); got != "gen_income_total" {
		t.Fatalf("physical: %s", got)
	}
}

func TestHeterogeneousIndexSkipsJSON(t *testing.T) {
	table := &schema.Table{
		Fields: []source.Field{
			{UUKey: "basic.utime", Group: "basic", Field: "utime", GType: consts.GTYPE_FLATTEN, Index: "basic.utime"},
			{UUKey: "income.total", Group: "income", Field: "total", GType: consts.GTYPE_GROUPED, Index: "gen_income_total"},
			{UUKey: "income.name", Group: "income", Field: "name", GType: consts.GTYPE_GROUPED, Index: "income.name"},
		},
	}
	ApplySearchIndex(table)

	if table.Fields[0].Index != "utime" {
		t.Fatalf("flatten index=%q want utime", table.Fields[0].Index)
	}
	if table.Fields[1].Index != "gen_income_total" {
		t.Fatalf("hetero index must be preserved, got %q", table.Fields[1].Index)
	}
	if table.Fields[1].Field != "total" {
		t.Fatalf("store field must stay total")
	}
	if table.Fields[2].Index != "income.name" {
		t.Fatalf("homogenous index=%q", table.Fields[2].Index)
	}

	if got := fieldExpr(table, "income.total"); got != "gen_income_total" {
		t.Fatalf("fieldExpr hetero=%q", got)
	}
	if got := fieldExpr(table, "income.name"); got != "json_extract(income, '$.name')" {
		t.Fatalf("fieldExpr json=%q", got)
	}
	if got := fieldExpr(table, "basic.utime"); got != "utime" {
		t.Fatalf("fieldExpr flatten=%q", got)
	}

	qs := []schema.Query{{Field: "income.total", Logic: "EQ", Value: 1}}
	bindQueryIndexes(table, &qs)
	if qs[0].Field != "income.total" {
		t.Fatalf("logical field mutated: %q", qs[0].Field)
	}
	if qs[0].Index != "gen_income_total" {
		t.Fatalf("query index=%q", qs[0].Index)
	}
	if got := querySQLExpr(qs[0]); got != "gen_income_total" {
		t.Fatalf("querySQLExpr=%q", got)
	}
}
