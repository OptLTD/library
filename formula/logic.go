package formula

// 逻辑类：Excel 风格 IF / IFS / SWITCH 及条件真假、匹配比较。

import (
	"fmt"
	"math"
	"reflect"
)

// LogicFuncs 为无状态的逻辑类 expr 函数接收者。
type LogicFuncs struct{}

var defaultLogic = LogicFuncs{}

// excelTruthy 模拟 Excel 中 IF/IFS 的条件：FALSE、0、空文本、nil 为假；非空文本为真（含 "0"、"FALSE"）。
func excelTruthy(v any) bool {
	if v == nil {
		return false
	}
	switch x := v.(type) {
	case bool:
		return x
	case float64:
		return x != 0 && !math.IsNaN(x)
	case float32:
		f := float64(x)
		return f != 0 && !math.IsNaN(f)
	case int:
		return x != 0
	case int32:
		return x != 0
	case int64:
		return x != 0
	case uint:
		return x != 0
	case uint32:
		return x != 0
	case uint64:
		return x != 0
	case string:
		return x != ""
	default:
		return true
	}
}

func switchEqual(a, b any) bool {
	if fa, errA := argToFloat64(a); errA == nil {
		if fb, errB := argToFloat64(b); errB == nil {
			return fa == fb
		}
	}
	if sa, ok := a.(string); ok {
		if sb, ok := b.(string); ok {
			return sa == sb
		}
	}
	if ba, ok := a.(bool); ok {
		if bb, ok := b.(bool); ok {
			return ba == bb
		}
	}
	return reflect.DeepEqual(a, b)
}

// excelIF 对应 Excel IF(logical_test, value_if_true, value_if_false)。
func (f LogicFuncs) excelIF(args ...any) (any, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("IF 需要 3 个参数：条件、真值、假值")
	}
	if excelTruthy(args[0]) {
		return args[1], nil
	}
	return args[2], nil
}

// excelIFS 对应 Excel IFS(cond1, val1, cond2, val2, …)；无一成立时返回错误（类 #N/A）。
func (f LogicFuncs) excelIFS(args ...any) (any, error) {
	if len(args) < 2 || len(args)%2 != 0 {
		return nil, fmt.Errorf("IFS 需要成对参数：条件1, 值1, 条件2, 值2, …")
	}
	for i := 0; i < len(args); i += 2 {
		if excelTruthy(args[i]) {
			return args[i+1], nil
		}
	}
	return nil, fmt.Errorf("IFS: 没有满足的条件")
}

// excelSwitch 对应 Excel SWITCH(expression, val1, res1, … [, default])；
// expression 之后若参数个数为奇数，则最后一项为默认值；否则无一匹配时返回错误（类 #N/A）。
func (f LogicFuncs) excelSwitch(args ...any) (any, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("SWITCH 至少需要 3 个参数：表达式、值1、结果1")
	}
	exprVal := args[0]
	rest := args[1:]
	n := len(rest)
	if n%2 == 1 {
		pairCount := (n - 1) / 2
		for i := 0; i < pairCount; i++ {
			if switchEqual(exprVal, rest[i*2]) {
				return rest[i*2+1], nil
			}
		}
		return rest[n-1], nil
	}
	for i := 0; i < n; i += 2 {
		if switchEqual(exprVal, rest[i]) {
			return rest[i+1], nil
		}
	}
	return nil, fmt.Errorf("SWITCH: 没有匹配项且无默认值")
}
