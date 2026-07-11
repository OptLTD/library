# Example

各 package 的最小用法示例。

```bash
go run ./example
```

## 目录

```
example/
├── main.go
├── go.mod
├── ability/           # formula / jsrunner / jsmodule
│   ├── formula.go
│   ├── jsrunner.go
│   └── jsmodule.go
└── search/            # search core + engine/*
    ├── search.go      # 入口 Run()
    ├── memory.go
    ├── mysql.go
    └── mysql_legacy.go  # //go:build ignore
```

## Search 升级对照

| 文件 | 说明 |
|------|------|
| `search/memory.go` | 升级后：`engine/memory` |
| `search/mysql_legacy.go` | 升级前：工厂 + registry（不参与编译） |
| `search/mysql.go` | 升级后：`engine/mysql` 显式构造 |

详见 [search/MIGRATION.md](../search/MIGRATION.md)。
