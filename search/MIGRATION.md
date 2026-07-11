# Search v0.1.x → v0.2.x 升级指南

v0.2.0 将存储实现拆为与 `search` **同级**的 `engine/*` 子模块；core 仅保留模型、解析与 `storage` 接口。

## 变更摘要

| 类别 | v0.1.x（旧） | v0.2.x（新） |
|------|-------------|-------------|
| 模块 | 单一 `search` 含全部 DB SDK | `search` core + `engine/*` 按需引入 |
| 接口包 | `search/engine` | `search/storage` |
| 接口类型 | `engine.IEngine` | `storage.IEngine` |
| 引擎工厂 | `engine.NewEngine(name, client)` | `engine/<name>.NewEngine(client)` |
| Loader | `loader.NewLoader(name)` + registry Init | `engine/<name>.NewLoader(client, tables)` |
| Memory | `engine.NewMemoryEngine()`（core 内） | `engine/memory.NewEngine()` |
| Redis 缓存 | `schema.NewCache("redis")` + registry | `engine/redis.NewCache(...)` |
| 序列号 | `support.SerialNo` + registry | `engine/redis.NewSerialNo(client)` |

### 模块路径对照

| 能力 | v0.1.x | v0.2.x |
|------|--------|--------|
| core | `github.com/OptLTD/library/search` | 不变 |
| memory | （core 内） | `github.com/OptLTD/library/engine/memory` |
| mysql | （core 内） | `github.com/OptLTD/library/engine/mysql` |
| postgres | 无 | `github.com/OptLTD/library/engine/postgres` |
| sqlite | 无 | `github.com/OptLTD/library/engine/sqlite` |
| mongo | （core 内） | `github.com/OptLTD/library/engine/mongo` |
| elastic | （core 内） | `github.com/OptLTD/library/engine/elastic` |
| redis | （core 内） | `github.com/OptLTD/library/engine/redis` |

## go.mod 迁移步骤

```bash
# 1. 升级 core
go get github.com/OptLTD/library/search@v0.2.0

# 2. 按需添加 engine 模块
go get github.com/OptLTD/library/engine/memory@v0.2.0   # 内存引擎
go get github.com/OptLTD/library/engine/mysql@v0.2.0    # 按需
go get github.com/OptLTD/library/engine/postgres@v0.2.0 # 按需
go get github.com/OptLTD/library/engine/sqlite@v0.2.0   # 按需
go get github.com/OptLTD/library/engine/mongo@v0.2.0    # 按需
go get github.com/OptLTD/library/engine/elastic@v0.2.0  # 按需
go get github.com/OptLTD/library/engine/redis@v0.2.0    # 按需

# 3. 清理
go mod tidy
```

## 代码迁移对照

### Memory

```go
// 旧
import "github.com/OptLTD/library/search/engine"
mem := engine.NewMemoryEngine()

// 新
import "github.com/OptLTD/library/engine/memory"
mem := memory.NewEngine()
```

### MySQL

```go
// 旧
support.Register(consts.DATABASE_MYSQL, gormDB)
support.Register(consts.DB_TABLE_NAMES, tables)
eng := engine.NewEngine(consts.SEARCH_DBMYSQL, gormDB)
ldr := loader.NewLoader(consts.LOADER_MYSQL)

// 新
import "github.com/OptLTD/library/engine/mysql"
eng := mysql.NewEngine(gormDB)
ldr := mysql.NewLoader(gormDB, tables)
```

### PostgreSQL / SQLite

```go
import "github.com/OptLTD/library/engine/postgres"
eng := postgres.NewEngine(gormDB)
ldr := postgres.NewLoader(gormDB, tables)

import "github.com/OptLTD/library/engine/sqlite"
eng := sqlite.NewEngine(gormDB)
ldr := sqlite.NewLoader(gormDB, tables)
```

应用层需自行引入 GORM dialect：`gorm.io/driver/postgres`、`gorm.io/driver/sqlite`。

### MongoDB

```go
// 旧
eng := engine.NewEngine(consts.SEARCH_MONGODB, mongoDB)
ldr := loader.NewLoader(consts.LOADER_MONGO)

// 新
import "github.com/OptLTD/library/engine/mongo"
eng := mongo.NewEngine(mongoDB)
ldr := mongo.NewLoader(mongoDB, tables)
```

### Elasticsearch

```go
// 旧
eng := engine.NewEngine(consts.SEARCH_ELASTIC, esClient)

// 新
import "github.com/OptLTD/library/engine/elastic"
eng := elastic.NewEngine(esClient)
```

### Redis 缓存与序列号

```go
// 旧
support.Register(consts.DATABASE_REDIS, redisClient)
cache := schema.NewCache("redis", ttl)
sn := &support.SerialNo{}
sn.Init(kind, options)

// 新
import searchredis "github.com/OptLTD/library/engine/redis"
cache := searchredis.NewCache(redisClient, "cache:", ttl)
searchredis.SetGlobalCache(redisClient, "cache:", ttl) // 可选
sn := searchredis.NewSerialNo(redisClient)
sn.Init(kind, options)
```

## 无需改动的部分

- `consts`、`source`、`schema`、`parser`、`request`、`respond` 的 import 路径不变
- `storage.IEngine` 业务方法（`Store` / `Search` / `Digest` 等）不变
- JSON / Embed Loader：`loader.NewLoader(consts.LOADER_JSON)` 仍在 core 中

## 破坏性变更

- 删除 `search/engine` 包 → 改为 `search/storage` + `engine/*`
- 删除 `engine.NewEngine` 工厂函数
- 删除 `loader.NewLoader` 对 `LOADER_MYSQL` / `LOADER_MONGO` 的分支
- 删除 core 内 registry 隐式取 DB 连接的 Loader Init 流程
- `support.SerialNo` 移至 `engine/redis`
- `schema.NewCache("redis")` 不再可用

## 示例对照

本仓库 [example/search/](../example/search/) 提供可运行示例与 v0.1.x 对照：

| 文件 | 说明 |
|------|------|
| `memory.go` | 升级后：`engine/memory.NewEngine()` |
| `mysql.go` | 升级后：`engine/mysql` 显式构造 |
| `mysql_legacy.go` | 升级前：工厂 + registry（`//go:build ignore`，不参与编译） |

运行全部示例：

```bash
go run ./example
```

## GORM dialect 说明

`search` core 与各 SQL 类 `engine/*`（mysql / postgres / sqlite）均不引入 dialect driver。应用层需自行：

```go
import "gorm.io/driver/mysql"
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
```

PostgreSQL / SQLite 同理，分别使用 `gorm.io/driver/postgres`、`gorm.io/driver/sqlite`。
