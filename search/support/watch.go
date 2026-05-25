package support

import (
	"time"

	"search/consts"
)

// DefaultSlowLogThresholdMs 未 Register(consts.SLOW_THRESHOLD) 时的回退值（毫秒）。
const DefaultSlowLogThresholdMs int64 = 50

// SlowLogThreshold 返回慢日志阈值：读 GetValue(consts.SLOW_THRESHOLD)，期望为整型（如 int64）。
func SlowLogThreshold() int64 {
	raw, ok := GetValue(consts.SLOW_THRESHOLD)
	if !ok || raw == nil {
		return DefaultSlowLogThresholdMs
	}
	switch t := raw.(type) {
	case int64:
		return clampSlowThreshold(t)
	case int:
		return clampSlowThreshold(int64(t))
	case uint:
		return clampSlowThreshold(int64(t))
	case uint32:
		return clampSlowThreshold(int64(t))
	case uint64:
		return clampSlowThreshold(int64(t))
	default:
		return DefaultSlowLogThresholdMs
	}
}

func clampSlowThreshold(ms int64) int64 {
	if ms < 0 {
		return DefaultSlowLogThresholdMs
	}
	return ms
}

// LogKVByCostMs 按 SlowLogThreshold 选择 INFO/DEBUG，并写入 cost_ms（置于 attrs 前部）。
func LogKVByCostMs(logid, msg string, costMs int64, attrs ...any) {
	prefix := append([]any{"cost_ms", costMs}, attrs...)
	if costMs > SlowLogThreshold() {
		LogInfoKV(logid, msg, prefix...)
	} else {
		LogDebugKV(logid, msg, prefix...)
	}
}

// LogWatchCostMs 用于 defer：根据注册 defer 时传入的 start，记录到当前返回为止的耗时。
// cost_ms > SlowLogThreshold 时写 INFO，否则 DEBUG。
// logid 可为空；msg 为日志语义标题；extra 为成对键值，语义与 LogInfoKV 的 attrs 一致。
// 会自动在首部附加 cost_ms（毫秒整数）。
//
// 用法：defer LogWatchCostMs(time.Now(), logid, "my op", "key1", v1, "key2", v2)
func LogWatchCostMs(start time.Time, logid, msg string, extra ...any) {
	ms := time.Since(start).Milliseconds()
	attrs := append([]any{"cost_ms", ms}, extra...)
	if ms > SlowLogThreshold() {
		LogInfoKV(logid, msg, attrs...)
	} else {
		LogDebugKV(logid, msg, attrs...)
	}
}
