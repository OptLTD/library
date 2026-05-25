# library

Go 多 module monorepo，供 superuser-api 及其他项目引用。

| 目录 | Module Path | 说明 |
|------|-------------|------|
| `search/` | `github.com/OptLTD/library/search` | 检索与数据模型引擎 |
| `formula/` | `github.com/OptLTD/library/formula` | 公式计算 |
| `jsrunner/` | `github.com/OptLTD/library/jsrunner` | JavaScript 运行时 |
| `jsmodule/` | `github.com/OptLTD/library/jsmodule` | JS 内置模块 |

## 本地开发

```bash
go work sync
go test ./search/... ./formula/... ./jsrunner/... ./jsmodule/...
```

## 发布前检查

Release workflow 会校验：

1. 各 `go.mod` 的 `module` 路径为 `github.com/OptLTD/library/<name>`
2. 不存在 `replace` 指令（否则外部 `go get` 会失败）
3. 测试通过

**发布 v0.1.1 前**请确认 `jsrunner/go.mod` 与 `jsmodule/go.mod` 里互相引用的版本号与本次 tag 一致。

Tag 格式为 `<module>/<version>`，例如 `search/v0.1.0`。

## GitHub Actions

仅在推送**版本 tag** 时触发，push 到 `main` 不会运行。

| 触发方式 | 行为 |
|----------|------|
| push `search/v0.1.0` 等 module tag | 跑测试 |
| push `release/v0.1.0` | 跑测试 + 为全部 module 打 tag 并 push |
| Actions 页面手动 Run workflow | 跑测试 + 发布（可指定 module） |

### 方式一：Actions 页面手动发布

1. 打开 **Actions → Release → Run workflow**
2. 填写 `version`（如 `v0.1.0`）
3. 可选：修改 `modules`（默认四个全发）
4. Run workflow

成功后自动创建并推送：

- `search/v0.1.0`
- `formula/v0.1.0`
- `jsrunner/v0.1.0`
- `jsmodule/v0.1.0`

### 方式二：推送 release tag 触发

```bash
git tag release/v0.1.0
git push origin release/v0.1.0
```

推送 `release/v0.1.0` 后，workflow 会自动为全部 module 打对应版本 tag。

## 引用方式

```bash
go get github.com/OptLTD/library/search@v0.1.0
go get github.com/OptLTD/library/formula@v0.1.0
go get github.com/OptLTD/library/jsrunner@v0.1.0
go get github.com/OptLTD/library/jsmodule@v0.1.0
```

```go
import (
    "github.com/OptLTD/library/search/engine"
    "github.com/OptLTD/library/formula"
    js "github.com/OptLTD/library/jsrunner"
    _ "github.com/OptLTD/library/jsmodule/cache"
)
```

私有仓库需设置：

```bash
go env -w GOPRIVATE=github.com/OptLTD/*
```

## 致谢

本仓库中的 `jsrunner` 与 `jsmodule` 源自 [shiroyk/ski](https://github.com/shiroyk/ski)（基于 Goja 的 JavaScript 运行时，含 Buffer、fetch、timers 等内置模块）。感谢原作者 [shiroyk](https://github.com/shiroyk) 的开源贡献。
