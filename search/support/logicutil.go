package support

import (
	"fmt"
	"reflect"
	"strings"
)

// copy from  condition.TernaryOperator
func If[T, U any](isTrue T, ifValue U, elseValue U) U {
	if Bool(isTrue) {
		return ifValue
	} else {
		return elseValue
	}
}

func Or[T any](ifValue T, elseValue T) T {
	if Bool(ifValue) {
		return ifValue
	} else {
		return elseValue
	}
}

func Bool[T any](value T) bool {
	val := any(value)

	// 处理 nil
	if val == nil {
		return false
	}

	switch m := val.(type) {
	case interface{ Bool() bool }:
		return m.Bool()
	case interface{ IsZero() bool }:
		return !m.IsZero()
	}

	// 使用反射检查值的零值
	rv := reflect.ValueOf(val)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.UnsafePointer:
		if rv.IsNil() {
			return false
		}
		rv = rv.Elem()
	case reflect.Map, reflect.Slice, reflect.Chan, reflect.Func:
		if rv.IsNil() {
			return false
		}
	}

	switch rv.Kind() {
	case reflect.Map, reflect.Slice:
		return rv.Len() != 0
	default:
		return !rv.IsZero()
	}
}

// Compare 比较两个值，返回 -1(a < b), 0(a == b), 1(a > b)
// isDesc 表示是否降序排序
func Compare(a, b any, isDesc bool) int {
	// 处理缺失值（nil）
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return 1 // nil 排在最后
	}
	if b == nil {
		return -1
	}

	// 处理零值：数字 0、空字符串等视为空值，排在最后
	if !Bool(a) && !Bool(b) {
		return 0
	}
	if !Bool(a) {
		return 1 // 零值排在最后
	}
	if !Bool(b) {
		return -1
	}

	// 类型比较：数字排在字符串之前
	typeA := getValueType(a)
	typeB := getValueType(b)
	if typeA != typeB {
		cmp := typeA - typeB // 数字类型(1) < 字符串类型(2)
		if isDesc {
			cmp = -cmp
		}
		return cmp
	}

	// 同类型比较
	var cmp int
	switch typeA {
	case 1: // 数字类型
		numA := toFloat64(a)
		numB := toFloat64(b)
		if numA < numB {
			cmp = -1
		} else if numA > numB {
			cmp = 1
		} else {
			cmp = 0
		}
	case 2: // 字符串类型
		strA := fmt.Sprintf("%v", a)
		strB := fmt.Sprintf("%v", b)
		cmp = strings.Compare(strA, strB)
	default:
		cmp = 0
	}

	if isDesc {
		cmp = -cmp
	}
	return cmp
}

// getValueType 获取值的类型：1-数字，2-字符串，0-其他
func getValueType(val any) int {
	switch val.(type) {
	case int, int32, int64, float32, float64:
		return 1
	case string:
		return 2
	default:
		return 0
	}
}

// toFloat64 将数值类型转换为 float64
func toFloat64(val any) float64 {
	switch v := val.(type) {
	case int:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case float32:
		return float64(v)
	case float64:
		return v
	default:
		return 0
	}
}
