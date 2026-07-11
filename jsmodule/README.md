# jsmodule

`github.com/OptLTD/library/jsmodule`

为 `jsrunner` 提供的 **JavaScript 内置模块**集合，在运行时注入 Web / Node 风格 API（fetch、Buffer、timers 等）。

## 用途

- 注册可被 JS `import` 或全局访问的内置模块（`modules.Register`）
- 提供模块加载器（CJS / ESM、`require`、`import()`）
- 按需 blank import 子包即可启用对应能力

## 内置模块

| 模块 | 说明 |
|------|------|
| `fetch` | HTTP 请求 / 响应、Headers、Cookie |
| `http/server` | 轻量 HTTP Server（`serve`） |
| `node:buffer` | Buffer、Blob、File |
| `node:stream/web` | ReadableStream 等 |
| `node:url` | URL、URLSearchParams |
| `node:encoding` | TextEncoder / TextDecoder |
| `base64` | Base64 编解码 |
| `timers` | setTimeout、setInterval |
| `signal` | AbortController / AbortSignal |
| `crypto` | 摘要与对称加密 |
| `cache` | 内存缓存 |
| `dom` | 简易 DOM 事件 |
| `html` | HTML 解析 |
| `ext` | 扩展工具 |

## 快速开始

```go
import (
    js "github.com/OptLTD/library/jsrunner"
    _ "github.com/OptLTD/library/jsmodule/fetch"
    _ "github.com/OptLTD/library/jsmodule/timers"
)

// VM 创建后，JS 中可直接使用 fetch、setTimeout 等
val, err := js.RunString(ctx, `
    const res = await fetch("https://example.com");
    return res.status;
`)
```

自定义模块可调用 `modules.Register` 注册，参见 `modules.go` 中的示例。

完整示例见 [example/ability/jsmodule.go](../example/ability/jsmodule.go)。

## 依赖说明

与 `jsrunner` 存在循环依赖，本仓库内通过 `go.work` 及 `require` + `replace` 联调；发版时需保持两边版本号一致，详见 [deploy.md](../deploy.md)。

## 致谢

本模块源自 [shiroyk/ski](https://github.com/shiroyk/ski)。

## 测试

```bash
cd jsmodule && go test ./...
```
