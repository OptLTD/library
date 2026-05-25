package formula

// 数学类：标量/数组与浮点转换、舍入、边际计费、sum 等。

import (
	"fmt"
	"log"
	"math"
	"reflect"
	"strconv"

	"github.com/duke-git/lancet/v2/slice"
)

// MathFuncs 为无状态的数学类 expr 函数接收者。
type MathFuncs struct{}

// defaultMath 供 build 注册与 Build 结果截断使用。
var defaultMath = MathFuncs{}

// --- 类型与数组（供 marginal、逻辑比较等复用） ---

func toInt64(v any) int64 {
	switch x := v.(type) {
	case int:
		return int64(x)
	case int32:
		return int64(x)
	case int64:
		return x
	default:
		return 0
	}
}

func argToFloat64(v any) (float64, error) {
	switch x := v.(type) {
	case float64:
		return x, nil
	case float32:
		return float64(x), nil
	case int:
		return float64(x), nil
	case int32:
		return float64(x), nil
	case int64:
		return float64(x), nil
	case uint:
		return float64(x), nil
	case uint32:
		return float64(x), nil
	case uint64:
		return float64(x), nil
	case string:
		return strconv.ParseFloat(x, 64)
	default:
		return 0, fmt.Errorf("无法将 %T 转为数字", v)
	}
}

func sliceToFloat64(v any) ([]float64, error) {
	if v == nil {
		return nil, nil
	}
	switch s := v.(type) {
	case []float64:
		return append([]float64(nil), s...), nil
	case []any:
		out := make([]float64, 0, len(s))
		for _, item := range s {
			f, err := argToFloat64(item)
			if err != nil {
				return nil, err
			}
			out = append(out, f)
		}
		return out, nil
	case []int:
		out := make([]float64, len(s))
		for i, n := range s {
			out[i] = float64(n)
		}
		return out, nil
	case []int32:
		out := make([]float64, len(s))
		for i, n := range s {
			out[i] = float64(n)
		}
		return out, nil
	case []int64:
		out := make([]float64, len(s))
		for i, n := range s {
			out[i] = float64(n)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("第二、三参数须为数字数组，当前为 %T", v)
	}
}

// --- 舍入（Build 结果截断也复用 truncFormat） ---

// truncFormat 保留 decimal 位小数，舍弃尾数，无进位运算。
func (f MathFuncs) truncFormat(num float64, decimal int) (float64, error) {
	d := float64(1)
	if decimal > 0 {
		d = math.Pow10(decimal)
	}
	res := strconv.FormatFloat(math.Trunc(num*d)/d, 'f', -1, 64)
	return strconv.ParseFloat(res, 64)
}

// CeilDecimal 舍弃的尾数不为 0 时强制进位。
func (f MathFuncs) CeilDecimal(num float64, decimal int) (float64, error) {
	d := float64(1)
	if decimal > 0 {
		d = math.Pow10(decimal)
	}
	res := strconv.FormatFloat(math.Ceil(num*d)/d, 'f', -1, 64)
	return strconv.ParseFloat(res, 64)
}

// FloorDecimal 强制舍弃尾数（向负无穷）。
func (f MathFuncs) FloorDecimal(num float64, decimal int) (float64, error) {
	d := float64(1)
	if decimal > 0 {
		d = math.Pow10(decimal)
	}
	res := strconv.FormatFloat(math.Floor(num*d)/d, 'f', -1, 64)
	return strconv.ParseFloat(res, 64)
}

// round 对应 Excel ROUND(number, num_digits)：num_digits 为非负时保留小数位，为负时舍入到十位、百位等。
// 仅传 number 时等价于 ROUND(number, 0)。
func (f MathFuncs) round(args ...any) (any, error) {
	if len(args) < 1 {
		return float64(0), fmt.Errorf("ROUND 需要至少 1 个参数")
	}
	num, err := argToFloat64(args[0])
	if err != nil {
		return float64(0), fmt.Errorf("ROUND: %w", err)
	}
	if len(args) == 1 {
		return math.Round(num), nil
	}
	decFloat, err := argToFloat64(args[1])
	if err != nil {
		return float64(0), fmt.Errorf("ROUND: %w", err)
	}
	decimal := int(decFloat)
	if decimal >= 0 {
		d := float64(1)
		if decimal > 0 {
			d = math.Pow10(decimal)
		}
		res := strconv.FormatFloat(math.Round(num*d)/d, 'f', -1, 64)
		return strconv.ParseFloat(res, 64)
	}
	d := math.Pow10(-decimal)
	res := strconv.FormatFloat(math.Round(num/d)*d, 'f', -1, 64)
	return strconv.ParseFloat(res, 64)
}

// ceiling 对应 Excel CEILING.MATH(number[, significance])：按基数向上舍入；省略 significance 时为 1。
func (f MathFuncs) ceiling(args ...any) (any, error) {
	if len(args) < 1 {
		return float64(0), fmt.Errorf("CEILING 需要至少 1 个参数")
	}
	num, err := argToFloat64(args[0])
	if err != nil {
		return float64(0), fmt.Errorf("CEILING: %w", err)
	}
	sig := 1.0
	if len(args) >= 2 {
		sig, err = argToFloat64(args[1])
		if err != nil {
			return float64(0), fmt.Errorf("CEILING: %w", err)
		}
		if sig <= 0 {
			return float64(0), fmt.Errorf("CEILING: 基数须为正数")
		}
	}
	res := strconv.FormatFloat(math.Ceil(num/sig)*sig, 'f', -1, 64)
	return strconv.ParseFloat(res, 64)
}

// floor 对应 Excel FLOOR.MATH(number[, significance])：按基数向下舍入；省略 significance 时为 1。
func (f MathFuncs) floor(args ...any) (any, error) {
	if len(args) < 1 {
		return float64(0), fmt.Errorf("FLOOR 需要至少 1 个参数")
	}
	num, err := argToFloat64(args[0])
	if err != nil {
		return float64(0), fmt.Errorf("FLOOR: %w", err)
	}
	sig := 1.0
	if len(args) >= 2 {
		sig, err = argToFloat64(args[1])
		if err != nil {
			return float64(0), fmt.Errorf("FLOOR: %w", err)
		}
		if sig <= 0 {
			return float64(0), fmt.Errorf("FLOOR: 基数须为正数")
		}
	}
	res := strconv.FormatFloat(math.Floor(num/sig)*sig, 'f', -1, 64)
	return strconv.ParseFloat(res, 64)
}

// --- 边际计费 ---

// marginal 按数量边际累进：总量按上界切段，每段「段内数量 × 该段单价」求和（类超额累进）。
// 参数：qty 计费数量；upperBounds 各段上界（不含 0、严格递增，长度 n）；unitPrices 各段单价（长度须为 n+1）。
// 例：MARGINAL(350, [200, 400], [0.5, 0.6, 0.8]) => 200×0.5 + 150×0.6 = 190；单档：MARGINAL(qty, [], [p]) 等价于 qty × p。
func (f MathFuncs) marginal(args ...any) (any, error) {
	if len(args) < 3 {
		return float64(0), fmt.Errorf("marginal 需要 3 个参数：数量、上界数组、单价数组")
	}
	qty, err := argToFloat64(args[0])
	if err != nil {
		return float64(0), fmt.Errorf("marginal: %w", err)
	}
	if qty < 0 {
		qty = 0
	}
	upper, err := sliceToFloat64(args[1])
	if err != nil {
		return float64(0), fmt.Errorf("marginal: %w", err)
	}
	prices, err := sliceToFloat64(args[2])
	if err != nil {
		return float64(0), fmt.Errorf("marginal: %w", err)
	}

	if len(upper) == 0 {
		if len(prices) != 1 {
			return float64(0), fmt.Errorf("marginal: 无上界时单价须恰好 1 个")
		}
		return qty * prices[0], nil
	}
	if len(prices) != len(upper)+1 {
		return float64(0), fmt.Errorf("marginal: 单价须比上界多 1 个（上界 %d 个、单价 %d 个）", len(upper), len(prices))
	}
	for i, u := range upper {
		if u <= 0 {
			return float64(0), fmt.Errorf("marginal: 上界须为正数且不含 0（首段已从 0 开始），见索引 %d", i)
		}
		if i > 0 && u <= upper[i-1] {
			return float64(0), fmt.Errorf("marginal: 上界须严格递增")
		}
	}

	var total float64
	prev := 0.0
	for i := 0; i < len(upper); i++ {
		limit := upper[i]
		seg := math.Min(qty, limit) - prev
		if seg > 0 {
			total += seg * prices[i]
		}
		prev = limit
		if qty <= limit {
			return total, nil
		}
	}
	total += (qty - prev) * prices[len(upper)]
	return total, nil
}

// --- 求和（含将数字字符串按数值累加，属数学运算） ---

func (f MathFuncs) doSum(args ...any) (any, error) {
	sum := 0.0
	for _, item := range args {
		if val, ok := item.(float64); ok {
			sum += float64(val)
		} else if val, ok := item.(float32); ok {
			sum += float64(val)
		} else if val, ok := item.(int64); ok {
			sum += float64(val)
		} else if val, ok := item.(int32); ok {
			sum += float64(val)
		} else if val, ok := item.(int); ok {
			sum += float64(val)
		} else if val, ok := item.(string); ok {
			num, err := strconv.ParseFloat(val, 64)
			if err != nil {
				log.Printf("类型转换错误: %s", item)
				continue
			}
			sum += num
		}
	}
	return (float64)(sum), nil
}

func isSlice(v any) bool {
	if v == nil {
		return false
	}
	return reflect.TypeOf(v).Kind() == reflect.Slice
}

func (f MathFuncs) sum(args ...any) (any, error) {
	if len(args) == 0 {
		return f.doSum()
	}
	if len(args) == 1 && isSlice(args[0]) {
		if values, ok := args[0].([]float64); ok {
			result := []any{}
			for _, val := range values {
				result = append(result, val)
			}
			return f.doSum(result...)
		}
		if values, ok := args[0].([]int64); ok {
			log.Println("int64 values===", values, ok)
			result := []any{}
			for _, val := range values {
				result = append(result, val)
			}
			return f.doSum(result...)
		}
		return 0, nil
	}
	if len(args) == 2 {
		if data, ok := args[0].([]any); ok {
			field, ok := args[1].(string)
			if !ok {
				return 0, nil
			}
			values := slice.Map(data, func(idx int, item any) any {
				if t, ok := item.(map[string]any); ok {
					return t[field]
				}
				return 0
			})
			return f.doSum(values...)
		}
	}
	return f.doSum(args...)
}
