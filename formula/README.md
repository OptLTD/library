# formula

`github.com/OptLTD/library/formula`

基于 [expr-lang/expr](https://github.com/expr-lang/expr) 的**公式计算**模块，提供接近 Excel 语义的逻辑、数学、日期时间与文本函数，供业务规则与 `search` 等模块调用。

## 用途

- 将扁平业务字段（如 `basic.qty`）映射为嵌套环境（`FormulaEnv`）
- 编译并执行公式字符串（`Build` / `BuildWithOptions` / `ExprOptions`）
- 内置 `IF` / `IFS` / `SWITCH`、`SUM`、`MARGINAL`、`NOW`、`DAYS` 等函数
- 预处理存量小写 `if(` 为 `IF(`（`NormalizeExcelIFCalls`）

## 文件说明

| 文件 | 内容 |
|------|------|
| `build.go` | `Build`、`BuildWithOptions`、`ExprOptions` |
| `logic.go` | `IF`、`IFS`、`SWITCH` |
| `math.go` | 数值运算、`SUM`、`MARGINAL` |
| `datetime.go` | `NOW`、`DAYS`、`HOURS`、`MINUTES` |
| `text.go` | 文本函数扩展点 |
| `env.go` | `FormulaEnv` |
| `normalize.go` | 小写 `if(` 规范化 |

## 快速开始

```go
import "github.com/OptLTD/library/formula"

combine := map[string]any{
    "basic.qty":   3.0,
    "basic.price": 100.0,
}
env := formula.FormulaEnv(combine)

total, err := formula.Build(`basic.qty * basic.price`, env)
// total == 300

level, err := formula.Build(
    `IFS(basic.qty >= 3, "bulk", basic.qty >= 1, "retail", true, "none")`,
    env,
)
// level == "bulk"
```

需要自定义 expr 选项或非默认行为时，使用 `BuildWithOptions` 或 `ExprOptions()` 自行 `expr.Compile` / `expr.Run`。

## 公式示例

```text
IF(basic.qty > 1, 100, 0)
IFS(basic.qty >= 3, "bulk", basic.qty >= 1, "retail", true, "none")
SUM(1, 2, 3)
MARGINAL(350, [200, 400], [0.5, 0.6, 0.8])
DAYS(结束时间 - 开始时间)
```

## 注意事项

- 逻辑函数请写 **`IF(...)`**（大写）；expr 内置小写 `if` 是语法关键字，不是 Excel 风格函数
- 存量小写 `if(` 可用 `formula.NormalizeExcelIFCalls(raw)` 预处理后再 `Build`
- `Build` 对数值结果截断保留 3 位小数；字符串等非数值结果原样返回

完整示例见 [example/demo/formula.go](../example/demo/formula.go)。

## 测试

```bash
cd formula && go test ./...
```
