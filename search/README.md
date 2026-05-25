# search

`github.com/OptLTD/library/search`

面向业务表单的**检索与数据模型引擎**：将模型、字段、分组、透视等配置解析为统一查询结构，并对接多种存储后端执行 CRUD 与聚合检索。

## 用途

- 定义数据模型（Model / Table / Field / Group）与查询输入（Input）
- 解析配置、组装 SQL / 查询 DSL（`parser`、`schema`、`source`）
- 通过统一 `engine` 接口访问 MySQL、MongoDB、Elasticsearch、内存存储
- 加载模型定义（JSON / MySQL / Mongo / embed）
- 返回结构化结果（`respond`）

## 包结构

| 包 | 职责 |
|----|------|
| `consts` | 常量与类型枚举 |
| `source` | 模型、字段、分组、透视等源数据结构 |
| `schema` | 运行时查询 schema（Input / Table / Query） |
| `parser` | 配置解析 |
| `engine` | 存储引擎抽象与实现 |
| `loader` | 模型加载 |
| `request` / `respond` | 请求上下文与响应结构 |
| `support` | 工具、公式桥接、序列号等 |

## 快速开始

```go
import (
    "github.com/OptLTD/library/search/consts"
    "github.com/OptLTD/library/search/engine"
    "github.com/OptLTD/library/search/schema"
    "github.com/OptLTD/library/search/source"
)

eng := engine.NewEngine(consts.SEARCH_MEMORY, nil)
input := &schema.Input{
    Model: &source.Model{UUKey: "demo", Source: "users"},
    // ...
}
// eng.Store / Select / Search / Digest ...
```

完整示例见 [example/demo/search.go](../example/demo/search.go)。

## 测试

```bash
cd search && go test ./...
```
