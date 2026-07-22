package parser_test

import (
	"testing"

	"github.com/OptLTD/library/search/consts"
	"github.com/OptLTD/library/search/parser"
	"github.com/OptLTD/library/search/request"
	"github.com/OptLTD/library/search/source"
)

func TestBuildFieldsUsesTableFieldOrder(t *testing.T) {
	val := &source.Value{
		Model: source.Model{UUKey: "tms.waybill"},
		Groups: map[string]source.Group{
			"basic":   {UUKey: "basic", GType: consts.GTYPE_FLATTEN},
			"income":  {UUKey: "income", GType: consts.GTYPE_GROUPED},
			"payment": {UUKey: "payment", GType: consts.GTYPE_GROUPED},
		},
		Fields: map[string]source.Field{
			"basic.uukey":     {UUKey: "basic.uukey", Field: "uukey", Group: "basic", Label: "运单号", SeqNo: 1},
			"basic.utime":     {UUKey: "basic.utime", Field: "utime", Group: "basic", Label: "开单日期", SeqNo: 2},
			"income.freight":  {UUKey: "income.freight", Field: "freight", Group: "income", Label: "收入运费", SeqNo: 1},
			"payment.freight": {UUKey: "payment.freight", Field: "freight", Group: "payment", Label: "支出运费", SeqNo: 1},
		},
		Tables: map[string]source.Table{
			"default": {
				UUKey: "default",
				Fields: []string{
					"basic.uukey",
					"basic.utime",
					"income.freight",
					"payment.freight",
				},
			},
		},
	}
	table, err := parser.NewTableParser(nil).Build(val, &request.Search{
		Model: "tms.waybill",
		Using: "default",
		Scene: consts.SCENE_SEARCH,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"basic.uukey", "basic.utime", "income.freight", "payment.freight"}
	if len(table.Fields) != len(want) {
		t.Fatalf("len=%d fields=%v", len(table.Fields), fieldKeys(table.Fields))
	}
	for i, key := range want {
		if table.Fields[i].UUKey != key {
			t.Fatalf("idx %d: got %s want %s (all=%v)", i, table.Fields[i].UUKey, key, fieldKeys(table.Fields))
		}
	}
}

func TestBuildFieldsOrdersByGroupThenSeq(t *testing.T) {
	val := &source.Value{
		Model: source.Model{UUKey: "tms.waybill"},
		Groups: map[string]source.Group{
			"basic":   {UUKey: "basic", GType: consts.GTYPE_FLATTEN, SeqNo: 0},
			"goods":   {UUKey: "goods", GType: consts.GTYPE_GROUPED, SeqNo: 1},
			"income":  {UUKey: "income", GType: consts.GTYPE_GROUPED, SeqNo: 2},
			"payment": {UUKey: "payment", GType: consts.GTYPE_GROUPED, SeqNo: 3},
		},
		Fields: map[string]source.Field{
			"income.freight":  {UUKey: "income.freight", Field: "freight", Group: "income", Label: "收入运费", SeqNo: 1},
			"basic.uukey":     {UUKey: "basic.uukey", Field: "uukey", Group: "basic", Label: "运单号", SeqNo: 1},
			"payment.freight": {UUKey: "payment.freight", Field: "freight", Group: "payment", Label: "支出运费", SeqNo: 1},
			"goods.weight":    {UUKey: "goods.weight", Field: "weight", Group: "goods", Label: "重量", SeqNo: 2},
			"goods.name":      {UUKey: "goods.name", Field: "name", Group: "goods", Label: "货名", SeqNo: 1},
			"basic.utime":     {UUKey: "basic.utime", Field: "utime", Group: "basic", Label: "开单日期", SeqNo: 2},
		},
		Tables: map[string]source.Table{
			"default": {UUKey: "default", Fields: []string{}},
		},
	}
	table, err := parser.NewTableParser(nil).Build(val, &request.Search{
		Model: "tms.waybill",
		Using: "default",
		Scene: consts.SCENE_SEARCH,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"basic.uukey", "basic.utime",
		"goods.name", "goods.weight",
		"income.freight",
		"payment.freight",
	}
	if len(table.Fields) != len(want) {
		t.Fatalf("len=%d fields=%v", len(table.Fields), fieldKeys(table.Fields))
	}
	for i, key := range want {
		if table.Fields[i].UUKey != key {
			t.Fatalf("idx %d: got %s want %s (all=%v)", i, table.Fields[i].UUKey, key, fieldKeys(table.Fields))
		}
	}
}


func fieldKeys(fields []source.Field) []string {
	out := make([]string, len(fields))
	for i, f := range fields {
		out[i] = f.UUKey
	}
	return out
}
