package search

import (
	"fmt"

	"github.com/OptLTD/library/engine/memory"
	"github.com/OptLTD/library/search/consts"
	"github.com/OptLTD/library/search/request"
	"github.com/OptLTD/library/search/respond"
	"github.com/OptLTD/library/search/schema"
	"github.com/OptLTD/library/search/source"
)

func SearchMemory() error {
	fmt.Println("\n--- search (memory, engine/memory) ---")

	input := demoSearchInput()
	table := demoSearchTable(input)

	mem := memory.NewEngine()
	rows := []struct {
		uukey string
		name  string
		age   int
	}{
		{"u001", "Alice", 25},
		{"u002", "Bob", 30},
	}

	for _, row := range rows {
		record := &respond.Record{
			UUKey:  row.uukey,
			Model:  "demo_model",
			OpType: "INSERT",
			Default: map[string]any{
				"uukey": row.uukey,
			},
			Prepare: map[string]any{
				"name":  row.name,
				"age":   row.age,
				"uukey": row.uukey,
			},
		}
		if err := mem.Store(input, record); err != nil {
			return err
		}
		fmt.Printf("stored %s (%s, age=%d)\n", row.uukey, row.name, row.age)
	}

	result, err := mem.Search(table)
	if err != nil {
		return err
	}
	fmt.Printf("search count=%d\n", result.Count)
	for i, value := range result.Values {
		fmt.Printf("  [%d] basic.name=%v basic.age=%v\n", i, value["basic.name"], value["basic.age"])
	}

	return nil
}

func demoSearchInput() *schema.Input {
	return &schema.Input{
		Model: &source.Model{
			UUKey:  "demo_model",
			Source: "demo_users",
			Search: "demo_users",
		},
		Groups: []source.Group{{
			UUKey: "basic",
			GType: consts.GTYPE_FLATTEN,
		}},
		Fields: []source.Field{
			{
				UUKey: "basic.name",
				Field: "name",
				Group: "basic",
				GType: consts.GTYPE_FLATTEN,
				FType: consts.FTYPE_STRINGS,
				Index: "name",
				Shown: true,
			},
			{
				UUKey: "basic.age",
				Field: "age",
				Group: "basic",
				GType: consts.GTYPE_FLATTEN,
				FType: consts.FTYPE_NUMERIC,
				Index: "age",
				Shown: true,
			},
			{
				UUKey: "basic.uukey",
				Field: "uukey",
				Group: "basic",
				GType: consts.GTYPE_FLATTEN,
				FType: consts.FTYPE_STRINGS,
				Index: "uukey",
				Shown: true,
			},
		},
		Unique: []string{"uukey"},
	}
}

func demoSearchTable(input *schema.Input) *schema.Table {
	return &schema.Table{
		Model: input.Model,
		Table: &source.Table{UUKey: "default"},
		Request: &request.Search{
			Model: "demo_model",
			Page:  1,
			Size:  10,
		},
		Groups: input.Groups,
		Fields: input.Fields,
	}
}
