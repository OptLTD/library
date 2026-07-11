package demo

import (
	"fmt"

	"github.com/OptLTD/library/search/loader"
	searchmysql "github.com/OptLTD/library/search/driver/mysql"
)

// SearchMySQL 演示 v0.2.x 按需依赖：显式构造 MySQL 引擎与 Loader。
// 实际业务中需先 gorm.Open 获得 *gorm.DB，此处仅展示 API 形态。
func SearchMySQL() error {
	fmt.Println("\n--- search (mysql driver, upgraded API) ---")

	tables := &loader.SchemaTableName{
		Model: "t_model",
		Table: "t_table",
		Input: "t_input",
	}

	// 升级后：不再使用 engine.NewEngine / loader.NewLoader + registry
	// var gormDB *gorm.DB // gorm.Open(mysql.Open(dsn), &gorm.Config{})
	// eng := searchmysql.NewEngine(gormDB)
	// ldr := searchmysql.NewLoader(gormDB, tables)
	// value, err := ldr.Load(ctx, "demo_model")
	// result, err := eng.Search(table)

	_ = searchmysql.NewEngine
	_ = tables

	fmt.Println("import: github.com/OptLTD/library/search/driver/mysql")
	fmt.Println("engine: searchmysql.NewEngine(gormDB)")
	fmt.Println("loader: searchmysql.NewLoader(gormDB, tables)")
	fmt.Println("go.mod: go get github.com/OptLTD/library/search/driver/mysql@v0.2.0")

	return nil
}
