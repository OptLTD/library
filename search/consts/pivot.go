package consts

import "strings"

// VALUE
const (
	// 桶聚合
	VALUE_TERMS = "TERMS"
	VALUE_RANGE = "RANGE"
	VALUE_HIST1 = "HIST1"
	// histogram, "interval" : 50 // 分桶的间隔为50，意思就是price字段值按50间隔分组
	VALUE_HIST2 = "HIST2" // date_histogram
	// "calendar_interval" : "month", // 分组间隔：month代表每月、支持minute（每分钟）、hour（每小时）、day（每天）、week（每周）、year（每年）
	// "format" : "yyyy-MM-dd" // 设置返回结果中桶key的时间格式

	VALUE_NESTED = "NESTED"

	// 基础聚合
	VALUE_CNT = "CNT" // value_count
	VALUE_UNQ = "UNQ" // cardinality
	VALUE_SUM = "SUM"
	VALUE_AVG = "AVG"
	VALUE_MAX = "MAX"
	VALUE_MIN = "MIN"
)

func IsMetric(name string) bool {
	switch strings.ToUpper(name) {
	case VALUE_CNT, VALUE_UNQ, VALUE_SUM, VALUE_AVG, VALUE_MAX, VALUE_MIN:
		return true
	}
	return false
}

func IsUnique(name string) bool {
	switch strings.ToUpper(name) {
	case VALUE_UNQ, VALUE_CNT:
		return true
	}
	return false
}

func IsBucket(name string) bool {
	switch strings.ToUpper(name) {
	case VALUE_TERMS, VALUE_RANGE, VALUE_HIST1, VALUE_HIST2:
		return true
	}
	return false
}

func IsNested(name string) bool {
	return strings.ToUpper(name) == VALUE_NESTED
}
