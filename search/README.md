# search

`github.com/OptLTD/library/search`

面向业务表单的**检索与数据模型引擎**：将模型、字段、分组、透视等配置解析为统一查询结构，并对接多种存储后端执行 CRUD 与聚合检索。

> **v0.2.0 升级**：数据库驱动已拆为独立子模块，见 [MIGRATION.md](MIGRATION.md)。

## 用途

- 定义数据模型（Model / Table / Field / Group）与查询输入（Input）
- 解析配置、组装 SQL / 查询 DSL（`parser`、`schema`、`source`）
- 通过统一 `engine` 接口访问 MySQL、MongoDB、Elasticsearch、内存存储
- 加载模型定义（JSON / MySQL / Mongo / embed）
- 返回结构化结果（`respond`）

## 模块结构

| 模块 | 路径 | 说明 |
|------|------|------|
| core | `github.com/OptLTD/library/search` | 接口、模型、Memory 引擎、JSON/Embed Loader |
| mysql | `.../search/driver/mysql` | MySQL 引擎 + Loader |
| mongo | `.../search/driver/mongo` | Mongo 引擎 + Loader |
| elastic | `.../search/driver/elastic` | Elasticsearch 引擎 |
| redis | `.../search/driver/redis` | Redis 缓存 + 序列号 |

## 包结构（core）

| 包 | 职责 |
|----|------|
| `consts` | 常量与类型枚举 |
| `source` | 模型、字段、分组、透视等源数据结构 |
| `schema` | 运行时查询 schema（Input / Table / Query） |
| `parser` | 配置解析 |
| `engine` | 存储引擎接口与 Memory 实现 |
| `loader` | 模型加载（JSON / Embed） |
| `request` / `respond` | 请求上下文与响应结构 |
| `support` | 工具函数 |

## 快速开始（仅 Memory，零 DB 依赖）

```go
import (
    "github.com/OptLTD/library/search/engine"
    "github.com/OptLTD/library/search/schema"
    "github.com/OptLTD/library/search/source"
)

mem := engine.NewMemoryEngine()
input := &schema.Input{
    Model: &source.Model{UUKey: "demo", Source: "users"},
    // ...
}
// mem.Store / Select / Search / Digest ...
```

## 按需引入驱动

```bash
go get github.com/OptLTD/library/search@v0.2.0
go get github.com/OptLTD/library/search/driver/mysql@v0.2.0  # 按需
```

```go
import searchmysql "github.com/OptLTD/library/search/driver/mysql"

eng := searchmysql.NewEngine(gormDB)
ldr := searchmysql.NewLoader(gormDB, tables)
```

完整示例见 [example/demo/](../example/demo/)。升级对照见 [MIGRATION.md](MIGRATION.md)。

## 测试

```bash
cd search && go test ./...
cd search/driver/mysql && go test ./...   # 各 driver 独立测试
```
