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
    ├── search.go            # search 入口（Memory + MySQL driver）
    ├── search_memory.go     # 升级后：core-only Memory 引擎
    ├── search_mysql.go      # 升级后：driver/mysql 显式 API
    ├── search_mysql_legacy.go  # 升级前对照（//go:build ignore）
    ├── jsrunner.go
    └── jsmodule.go
```

## Search 升级对照

| 文件 | 说明 |
|------|------|
| `search_memory.go` | 升级后：仅用 `search` core，零 DB SDK |
| `search_mysql_legacy.go` | 升级前：`NewEngine` + `NewLoader` + registry（不参与编译） |
| `search_mysql.go` | 升级后：`search/driver/mysql` 显式构造 |

详见 [search/MIGRATION.md](../search/MIGRATION.md)。
