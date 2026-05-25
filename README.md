# library

Go 多 module monorepo，为 [superuser-api](https://github.com/OptLTD) 等业务提供可复用的基础能力。

## 模块

### [search](search/README.md) — 检索与数据模型引擎

将模型、字段、分组配置解析为统一查询，对接 MySQL / MongoDB / Elasticsearch / 内存存储。

```go
import (
    "github.com/OptLTD/library/search/engine"
    "github.com/OptLTD/library/search/respond"
    "github.com/OptLTD/library/search/schema"
)

mem := engine.NewMemoryEngine()
mem.Store(input, &respond.Record{ /* ... */ })
result, _ := mem.Search(table)
```

### [formula](formula/README.md) — 公式计算

基于 expr，提供接近 Excel 的 `IF` / `IFS` / `SUM` 等函数，将扁平字段映射为公式环境。

```go
import "github.com/OptLTD/library/formula"

env := formula.FormulaEnv(map[string]any{
    "basic.qty": 3.0, "basic.price": 100.0,
})
total, _ := formula.Build(`basic.qty * basic.price`, env)
level, _ := formula.Build(`IFS(basic.qty >= 3, "bulk", true, "retail")`, env)
```

### [jsrunner](jsrunner/README.md) — JavaScript 运行时

在 Go 中嵌入 ESM/CJS 模块执行，支持事件循环与 Promise。

```go
import (
    "context"
    js "github.com/OptLTD/library/jsrunner"
)

ctx := context.Background()
val, _ := js.RunString(ctx, `1 + 2 * 3`)

mod, _ := js.CompileModule("greet", `export default (name) => "hello, " + name`)
greeting, _ := js.RunModule(ctx, mod, "world")
```

### [jsmodule](jsmodule/README.md) — JS 内置模块

为 jsrunner 注入 fetch、Buffer、timers、TextEncoder 等 Web / Node 风格 API。

```go
import (
    js "github.com/OptLTD/library/jsrunner"
    _ "github.com/OptLTD/library/jsmodule/buffer"
    _ "github.com/OptLTD/library/jsmodule/encoding"
)

encoded, _ := js.RunString(ctx, `btoa("hello")`)
length, _ := js.RunString(ctx, `new TextEncoder().encode("你好").length`)
```

运行完整示例：`go run ./example`（详见 [example/README.md](example/README.md)）。

## 文档

- [deploy.md](deploy.md) — 发布流程、`go get` 引用方式
- 各模块详细说明见对应目录下的 `README.md`

## License

MIT — 见 [LICENSE](LICENSE)。

## 致谢

`jsrunner` 与 `jsmodule` 源自 [shiroyk/ski](https://github.com/shiroyk/ski)（基于 Goja 的 JavaScript 运行时）。感谢 [shiroyk](https://github.com/shiroyk) 的开源贡献。
