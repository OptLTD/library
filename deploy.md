# 发布与部署

本文档说明如何将 monorepo 中的 Go module 发布到 GitHub / pkg.go.dev，供外部项目引用。

## 发布前检查

Release workflow 会校验：

1. 各 `go.mod` 的 `module` 路径为 `github.com/OptLTD/library/<name>`
2. 不存在 `replace` 指令（`jsrunner` / `jsmodule` 开发用 replace 指向本地目录，Release workflow 打 tag 前会自动去掉）
3. `jsrunner` 与 `jsmodule` 互相 `require` 的版本号与本次 tag 一致
4. 测试通过

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

### 本地开发引用

在本 monorepo 内使用 `go.work` 即可，无需额外配置。

外部项目引用本地 checkout 时，在消费方 `go.mod` 中添加 `replace`：

```go
replace (
    github.com/OptLTD/library/jsrunner => ../option-library/jsrunner
    github.com/OptLTD/library/jsmodule => ../option-library/jsmodule
)
```

`jsrunner` 与 `jsmodule` 在本仓库内通过 `require` + `replace` 互相指向 sibling 目录；发版时 workflow 会自动处理。

### 私有仓库

若仓库设为 private，需配置：

```bash
go env -w GOPRIVATE=github.com/OptLTD/*
```
