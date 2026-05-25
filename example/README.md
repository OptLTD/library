# Example

各 package 的最小用法示例。

```bash
go run ./example
```

## 目录

```
example/
├── main.go          # 入口
├── go.mod
└── demo/            # 各 package 示例
    ├── formula.go
    ├── search.go
    ├── jsrunner.go
    └── jsmodule.go
```

## 致谢

本仓库中的 `jsrunner` 与 `jsmodule` 源自 [shiroyk/ski](https://github.com/shiroyk/ski)（基于 Goja 的 JavaScript 运行时，含 Buffer、fetch、timers 等内置模块）。感谢原作者 [shiroyk](https://github.com/shiroyk) 的开源贡献。
