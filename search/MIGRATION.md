# Search v0.1.x → v0.2.x 升级指南

v0.2.0 将数据库驱动拆为独立子模块，core 模块不再携带 ES / Mongo / GORM / Redis 等 SDK。按需 `go get` 对应 driver 即可。

## 变更摘要

| 类别 | v0.1.x（旧） | v0.2.x（新） |
|------|-------------|-------------|
| 模块 | 单一 `search` 含全部 DB SDK | `search` core + `search/driver/*` 按需引入 |
| 引擎 | `engine.NewEngine(name, client)` | `driver/<name>.NewEngine(client)` |
| Loader | `loader.NewLoader(name)` + registry Init | `driver/<name>.NewLoader(client, tables)` |
| Redis 缓存 | `schema.NewCache("redis")` + registry | `redis.NewCache(client, prefix, ttl)` |
| 序列号 | `support.SerialNo` + registry | `redis.NewSerialNo(client)` |
| Memory | `engine.NewMemoryEngine()` 或 `NewEngine(SEARCH_MEMORY, nil)` | 仅 `engine.NewMemoryEngine()`（core 内置） |

## go.mod 迁移步骤

```bash
# 1. 升级 core
go get github.com/OptLTD/library/search@v0.2.0

# 2. 按需添加驱动（仅用 MySQL 则只加这一条）
go get github.com/OptLTD/library/search/driver/mysql@v0.2.0
go get github.com/OptLTD/library/search/driver/mongo@v0.2.0    # 可选
go get github.com/OptLTD/library/search/driver/elastic@v0.2.0  # 可选
go get github.com/OptLTD/library/search/driver/redis@v0.2.0    # 可选

# 3. 清理未使用的 indirect 依赖
go mod tidy
```

## 代码迁移对照

### Engine — MySQL

```go
// 旧
support.Register(consts.DATABASE_MYSQL, gormDB)
eng := engine.NewEngine(consts.SEARCH_DBMYSQL, gormDB)

// 新
import searchmysql "github.com/OptLTD/library/search/driver/mysql"
eng := searchmysql.NewEngine(gormDB)
```

### Loader — MySQL

```go
// 旧
support.Register(consts.DATABASE_MYSQL, gormDB)
support.Register(consts.DB_TABLE_NAMES, tables)
ldr := loader.NewLoader(consts.LOADER_MYSQL)

// 新
import searchmysql "github.com/OptLTD/library/search/driver/mysql"
ldr := searchmysql.NewLoader(gormDB, tables)
```

### Engine — MongoDB

```go
// 旧
eng := engine.NewEngine(consts.SEARCH_MONGODB, mongoDB)

// 新
import searchmongo "github.com/OptLTD/library/search/driver/mongo"
eng := searchmongo.NewEngine(mongoDB)
```

### Loader — MongoDB

```go
// 旧
ldr := loader.NewLoader(consts.LOADER_MONGO)

// 新
import searchmongo "github.com/OptLTD/library/search/driver/mongo"
ldr := searchmongo.NewLoader(mongoDB, tables)
```

### Engine — Elasticsearch

```go
// 旧
eng := engine.NewEngine(consts.SEARCH_ELASTIC, esClient)

// 新
import searchelastic "github.com/OptLTD/library/search/driver/elastic"
eng := searchelastic.NewEngine(esClient)
```

### Redis 缓存与序列号

```go
// 旧
support.Register(consts.DATABASE_REDIS, redisClient)
cache := schema.NewCache("redis", ttl)
sn := &support.SerialNo{}
sn.Init(kind, options)

// 新
import searchredis "github.com/OptLTD/library/search/driver/redis"
cache := searchredis.NewCache(redisClient, "cache:", ttl)
searchredis.SetGlobalCache(redisClient, "cache:", ttl) // 可选：设为全局缓存
sn := searchredis.NewSerialNo(redisClient)
sn.Init(kind, options)
```

## 无需改动的部分

- `consts`、`source`、`schema`、`parser`、`request`、`respond` 的 import 路径不变
- `IEngine` / `ILoader` 接口不变，业务层 `Store` / `Search` / `Digest` 调用不变
- Memory 引擎：`engine.NewMemoryEngine()` 前后一致
- JSON / Embed Loader：`loader.NewLoader(consts.LOADER_JSON)` 仍在 core 中

## 破坏性变更

- 删除 `engine.NewEngine`
- 删除 `loader.NewLoader` 对 `LOADER_MYSQL` / `LOADER_MONGO` 的分支
- 删除 core 内通过 registry 隐式取 DB 连接的 Loader Init 流程
- `support.SerialNo` 移至 `search/driver/redis`
- `schema.NewCache("redis")` 不再可用，请使用 redis driver

## 示例对照

本仓库 [example/demo/](../example/demo/) 提供：

| 文件 | 说明 |
|------|------|
| `search_memory.go` | 升级后：仅用 core，零 DB SDK |
| `search_mysql_legacy.go` | 升级前：工厂 + registry（`//go:build ignore`） |
| `search_mysql.go` | 升级后：显式 driver 构造 |

## MySQL dialect 说明

core 与 `driver/mysql` 均不引入 `gorm.io/driver/mysql`。应用层需自行：

```go
import "gorm.io/driver/mysql"
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
```
