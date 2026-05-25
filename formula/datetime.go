package formula

// 日期时间类：当前时刻、Duration 转天/小时/分钟等标量。

import "time"

// DateTimeFuncs 为无状态的日期时间类 expr 函数接收者。
type DateTimeFuncs struct{}

var defaultDateTime = DateTimeFuncs{}

// now 返回当前时间，用于与日期字段做运算，如 (结束日期-now())/(结束日期-开始日期)*100
func (f DateTimeFuncs) now(args ...any) (any, error) {
	return time.Now(), nil
}

func (f DateTimeFuncs) durationDays(args ...any) (any, error) {
	if len(args) < 1 {
		return float64(0), nil
	}
	if d, ok := args[0].(time.Duration); ok {
		return d.Hours() / 24, nil
	}
	return float64(0), nil
}

func (f DateTimeFuncs) durationHours(args ...any) (any, error) {
	if len(args) < 1 {
		return float64(0), nil
	}
	if d, ok := args[0].(time.Duration); ok {
		return d.Hours(), nil
	}
	return float64(0), nil
}

func (f DateTimeFuncs) durationMinutes(args ...any) (any, error) {
	if len(args) < 1 {
		return float64(0), nil
	}
	if d, ok := args[0].(time.Duration); ok {
		return d.Minutes(), nil
	}
	return float64(0), nil
}
