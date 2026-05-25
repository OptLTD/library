# package formula

基于 `expr-lang/expr` 的公式模块；**按类型分文件**，便于对照 Excel 语义扩展。逻辑函数推荐 **`IF` / `IFS` / `SWITCH`（大写）**；`SUM`、`NOW`、`DAYS` 等同时提供小写别名（如 `sum` / `SUM`），与存量公式兼容。

| 文件 | 类型 | 内容 |
|------|------|------|
| `build.go` | 入口 | `Build`、`ExprOptions` 与各类型 `default*` 的 expr 注册 |
| `logic.go` | 逻辑 | `LogicFuncs`：`IF`、`IFS`、`SWITCH`、条件真假与匹配 |
| `math.go` | 数学 | `MathFuncs`：数值/数组转换、舍入、公式 `SUM` / `marginal` / `MARGINAL` |
| `datetime.go` | 日期时间 | `DateTimeFuncs`：`NOW` / `DAYS` / `HOURS` / `MINUTES`（Duration→标量） |
| `text.go` | 文字 | 占位说明；独立文本公式（如 CONCAT）在此扩展 |
| `env.go` | 环境 | `FormulaEnv`：扁平 `a.b` → 嵌套 map |
| `normalize.go` | 预处理 | `NormalizeExcelIFCalls`：小写 `if(` → `IF(`（词边界，不误伤 `sumif` 等） |

`search/support` 仍可通过 `support.Build` / `support.FormulaEnv` 薄封装调用本包。

---

## Go 侧用法

将业务 **Combine**（扁平键，带点号）交给 `FormulaEnv`，再 `Build` 公式字符串：

```go
import formulalib "formula"

combine := map[string]any{
    "basic.opt1": 3.0,
    "basic.opt2": 1.0,
    "amount":     100.0,
}
env := formulalib.FormulaEnv(combine)

out, err := formulalib.Build(`IFS(basic.opt1 > 1, amount * 2, basic.opt1 > 0, amount, 0)`, env)
// out 为 float64，且会按 Build 规则截断保留 2 位小数
```

需要**自行控制 Run**（例如保留字符串结果）时，使用 `ExprOptions()` 与 `expr.Compile` / `expr.Run` 即可。

---

## 公式用例（字符串）

以下假定已用 `FormulaEnv` 把 `basic.opt1` 等转成嵌套字段，公式里写 `basic.opt1`。

**IFS（多条件，从左到右第一个为真即返回对应值）**

```text
IFS(basic.opt1 > 1, 100, basic.opt1 > 0, 50, true, 0)
```

小写别名：`ifs(...)` 与 `IFS(...)` 等价。

**IF（二选一）**

```text
IF(basic.opt1 > 1, IF(basic.opt2 > 0, 99, 0), -1)
```

函数请用 **`IF(...)`**。

**SWITCH（匹配表达式；最后一项可为默认值）**

```text
SWITCH(basic.opt1, 1, "档A", 2, "档B", "其它")
```

**SUM**

```text
SUM(1, 2, 3)
SUM([1, 2, 3])
```

对行数组按字段求和（第二参为字段名）：

```text
SUM(rows, "price")
```

**边际计费 MARGINAL**

```text
MARGINAL(350, [200, 400], [0.5, 0.6, 0.8])
```

单档：`MARGINAL(qty, [], [单价])` 等价于 `qty * 单价`。

**日期与时间差（结果为 `time.Duration` 时，用 DAYS/HOURS/MINUTES 转成数）**

```text
(结束时间 - NOW()) / (结束时间 - 开始时间) * 100
DAYS(结束时间 - 开始时间)
```

（具体字段名以你 env 里的 `time.Time` / `Duration` 为准。）

---

## `Build` 行为说明

- 成功时：若结果是 `float64` 或整数类型，会**截断保留 2 位小数**（与原先 `search/support` 一致）。
- 若结果是其它类型（如字符串），当前实现会落成 **0**；需要字符串等非数字结果时请用 `ExprOptions()` 自行 `Compile`/`Run`。

更多边界与嵌套场景见 `build_test.go`、`env_test.go`。

---

## 小写 `if` 能否关掉（像 Excel 一样写 `if(...)`）？

**不能。** `github.com/expr-lang/expr` **没有**提供「关闭关键字 `if`」的 `Option`。词法阶段会把单词 `if` 标成运算符，语法层在表达式开头会解析成 `if 条件 { … } else { … }`，与「函数调用 `if(a,b,c)`」不是同一条路，因此也无法像 `DisableBuiltin("sum")` 那样用配置关掉。

可行做法：

1. **推荐**：公式里统一写 **`IF(...)`**（ lexer 里大写 `IF` 是普通标识符，可走自定义函数）。
2. **入库前预处理**：若存量数据是小写 `if(`，可在落库或执行前把 **`if(` 替换成 `IF(`**（需约定公式里不使用 expr 自带的 `if x {} else {}` 语法，否则可能误伤）。
3. **上游**：向 expr 提需求或 fork，在 lexer 里不再把 `if` 当作关键字（维护成本较高）。

原生条件语法仍可用，例如：`if basic.opt1 > 1 { 100 } else { 0 }`，与 Excel 的 `IF` 三参形式不同。

### 存量公式是小写 `if(` 时如何安全替换？

不要用裸字符串替换 `if(`，否则理论上可能误伤（极少见）。请用**词边界**：把「独立单词 `if` + 可选空白 + `(`」改成 `IF(`。

本包提供：

```go
code := formula.NormalizeExcelIFCalls(raw) // 再 Build(code, env)
```

- **`sumif(`、`COUNTIF(`、`verify(`** 等：**不会**被改（`if` 前面仍是字母，不满足 `\bif`）。
- **如何确认**：看 `normalize_test.go` 里的用例；你也可在业务侧加 golden 测试（典型公式跑一遍 `Normalize` + `Compile`）。

注意：若公式里出现 **`else if (`**（`else` 与 `if` 分开写），也会被替换成 `else IF (`，一般 Excel 公式里极少这样写；若存在，请避免依赖本替换或单独处理。
