//go:build ignore

// 以下为 v0.1.x 用法对照，不参与 go run ./example 编译。
// 升级后请改用 search_mysql.go。

package demo

import (
	"context"

	"github.com/OptLTD/library/search/consts"
	"github.com/OptLTD/library/search/engine"
	"github.com/OptLTD/library/search/loader"
	"github.com/OptLTD/library/search/support"
	"gorm.io/gorm"
)

func searchMySQLLegacy(gormDB *gorm.DB) error {
	tables := &loader.SchemaTableName{
		Model: "t_model",
		Table: "t_table",
		Input: "t_input",
	}

	// v0.1.x：通过 registry 注入连接，工厂按名称选择实现
	support.Register(consts.DATABASE_MYSQL, gormDB)
	support.Register(consts.DB_TABLE_NAMES, tables)

	eng := engine.NewEngine(consts.SEARCH_DBMYSQL, gormDB)
	ldr := loader.NewLoader(consts.LOADER_MYSQL)

	ctx := context.Background()
	value, err := ldr.Load(ctx, "demo_model")
	if err != nil {
		return err
	}
	_ = eng
	_ = value
	return nil
}
