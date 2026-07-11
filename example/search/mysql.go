package search

import (
	"fmt"

	"github.com/OptLTD/library/engine/mysql"
	"github.com/OptLTD/library/search/loader"
)

// SearchMySQL 演示 v0.2.x 按需依赖：显式构造 MySQL 引擎与 Loader。
func SearchMySQL() error {
	fmt.Println("\n--- search (mysql engine, upgraded API) ---")

	tables := &loader.SchemaTableName{
		Model: "t_model",
		Table: "t_table",
		Input: "t_input",
	}

	// var gormDB *gorm.DB
	// eng := mysql.NewEngine(gormDB)
	// ldr := mysql.NewLoader(gormDB, tables)

	_ = mysql.NewEngine
	_ = tables

	fmt.Println("import: github.com/OptLTD/library/engine/mysql")
	fmt.Println("engine: mysql.NewEngine(gormDB)")
	fmt.Println("loader: mysql.NewLoader(gormDB, tables)")
	fmt.Println("go.mod: go get github.com/OptLTD/library/engine/mysql@v0.2.0")

	return nil
}
