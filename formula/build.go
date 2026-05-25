package formula

// 编译入口：expr 注册、Build / ExprOptions。

import (
	"fmt"

	"github.com/expr-lang/expr"
)

// BuildWithOptions 编译并执行公式；支持附加 expr.Option（如注入自定义函数）。
// 数值结果会按既有规则做截断；非数值结果原样返回。
func BuildWithOptions(code string, data map[string]any, extraOptions ...expr.Option) (any, error) {
	options := append(exprOptions(), extraOptions...)
	program, err := expr.Compile(code, options...)
	if err != nil {
		return nil, fmt.Errorf("公式解析失败：%v", err.Error())
	}

	output, err := expr.Run(program, data)
	if err != nil {
		return nil, fmt.Errorf("计算结果错误：%s", err.Error())
	}
	switch v := output.(type) {
	case float64:
		return defaultMath.truncFormat(v, 3)
	case int, int32, int64:
		return defaultMath.truncFormat(float64(toInt64(v)), 3)
	default:
		return output, nil
	}
}

// Build 编译并执行公式；保留历史行为，并在非数值结果时返回原值。
func Build(code string, data map[string]any) (any, error) {
	return BuildWithOptions(code, data)
}

func exprOptions() []expr.Option {
	m, dt, lg := defaultMath, defaultDateTime, defaultLogic
	return []expr.Option{
		expr.DisableBuiltin("sum"),
		// logic
		expr.Function("IF", lg.excelIF),
		expr.Function("ifs", lg.excelIFS),
		expr.Function("IFS", lg.excelIFS),
		expr.Function("switch", lg.excelSwitch),
		expr.Function("SWITCH", lg.excelSwitch),
		// datetime
		expr.Function("now", dt.now),
		expr.Function("days", dt.durationDays),
		expr.Function("hours", dt.durationHours),
		expr.Function("minutes", dt.durationMinutes),
		expr.Function("NOW", dt.now),
		expr.Function("DAYS", dt.durationDays),
		expr.Function("HOURS", dt.durationHours),
		expr.Function("MINUTES", dt.durationMinutes),
		// math
		expr.Function("sum", m.sum),
		expr.Function("SUM", m.sum),
		expr.Function("round", m.round),
		expr.Function("ROUND", m.round),
		expr.Function("floor", m.floor),
		expr.Function("FLOOR", m.floor),
		expr.Function("ceiling", m.ceiling),
		expr.Function("CEILING", m.ceiling),
		expr.Function("marginal", m.marginal),
		expr.Function("MARGINAL", m.marginal),
		expr.AllowUndefinedVariables(),
	}
}

// ExprOptions 返回与 Build 相同的 expr 编译选项，供需要自定义 Run 流程的场景使用。
func ExprOptions() []expr.Option {
	return exprOptions()
}
