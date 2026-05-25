# jsrunner

`github.com/OptLTD/library/jsrunner`

基于 [Goja (sobek)](https://github.com/grafana/sobek) 的 **JavaScript 运行时**，在 Go 进程中嵌入 ESM/CJS 模块执行、事件循环与 Promise 支持。

## 用途

- 创建与管理 JS 虚拟机（`VM`）
- 编译并运行 ESM/CJS 模块（`CompileModule`、`RunModule`）
- 执行脚本字符串（`RunString`）
- 事件循环、定时任务调度、异步 Promise
- 与 `jsmodule` 配合加载 Buffer、fetch、timers 等内置能力

## 主要组件

| 包 / 文件 | 职责 |
|-----------|------|
| `vm.go` / `js.go` | VM 接口与工厂 |
| `eventloop.go` | 事件循环 |
| `scheduler.go` | 任务调度 |
| `loader.go` | 模块加载 |
| `promise/` | Promise 实现 |
| `types/` | 类型辅助 |
| `modulestest/` | 测试用 VM 封装 |

## 快速开始

```go
import (
    "context"
    "time"

    js "github.com/OptLTD/library/jsrunner"
)

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// 执行表达式
val, err := js.RunString(ctx, `1 + 2 * 3`)

// 运行 ESM 模块
mod, err := js.CompileModule("greet", `export default (name) => "hello, " + name`)
greeting, err := js.RunModule(ctx, mod, "world")
```

配合内置模块：

```go
import (
    js "github.com/OptLTD/library/jsrunner"
    _ "github.com/OptLTD/library/jsmodule/timers" // 按需 blank import
)
```

完整示例见 [example/demo/jsrunner.go](../example/demo/jsrunner.go)。

## 依赖说明

与 `jsmodule` 存在循环依赖，本仓库内通过 `go.work` 及 `require` + `replace` 联调；发版时需保持两边版本号一致，详见 [deploy.md](../deploy.md)。

## 测试

```bash
cd jsrunner && go test ./...
```
