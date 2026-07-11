# search

`github.com/OptLTD/library/search`

面向业务表单的**检索与数据模型引擎**：模型、字段、分组、透视配置解析为统一查询结构。存储实现位于同级 `engine/*` 模块。

> 升级见 [MIGRATION.md](MIGRATION.md)。

## 模块结构

| 模块 | 路径 | 说明 |
|------|------|------|
| core | `github.com/OptLTD/library/search` | 模型、schema、parser、loader、`storage` 接口 |
| memory | `github.com/OptLTD/library/engine/memory` | 内存引擎 |
| mysql | `github.com/OptLTD/library/engine/mysql` | MySQL 引擎 + Loader |
| postgres | `github.com/OptLTD/library/engine/postgres` | PostgreSQL 引擎 + Loader |
| sqlite | `github.com/OptLTD/library/engine/sqlite` | SQLite 引擎 + Loader |
| mongo | `github.com/OptLTD/library/engine/mongo` | Mongo 引擎 + Loader |
| elastic | `github.com/OptLTD/library/engine/elastic` | Elasticsearch 引擎 |
| redis | `github.com/OptLTD/library/engine/redis` | Redis 缓存 + 序列号 |

## 包结构（core）

| 包 | 职责 |
|----|------|
| `storage` | `IEngine` / `ICallable` 接口 |
| `consts` | 常量与类型枚举 |
| `source` | 模型、字段、分组等源数据结构 |
| `schema` | 运行时查询 schema |
| `parser` | 配置解析 |
| `loader` | 模型加载（JSON / Embed） |
| `request` / `respond` | 请求与响应结构 |
| `support` | 工具函数 |

## 快速开始

```go
import (
    "github.com/OptLTD/library/engine/memory"
    "github.com/OptLTD/library/search/schema"
    "github.com/OptLTD/library/search/source"
)

mem := memory.NewEngine()
```

## 按需引入

```bash
go get github.com/OptLTD/library/search@v0.2.0
go get github.com/OptLTD/library/engine/mysql@v0.2.0
```

```go
import "github.com/OptLTD/library/engine/mysql"

eng := mysql.NewEngine(gormDB)
ldr := mysql.NewLoader(gormDB, tables)
```

完整示例见 [example/search/](../example/search/)。升级对照见 [MIGRATION.md](MIGRATION.md)。

## 测试

```bash
cd search && go test ./...
cd engine/memory && go test ./...
cd engine/mysql && go build ./...
```
