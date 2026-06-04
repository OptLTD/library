package schema

import (
	"fmt"
	"strings"
	"time"

	"github.com/OptLTD/library/search/consts"
	"github.com/OptLTD/library/search/support"

	"github.com/duke-git/lancet/v2/slice"
)

var SubQuery = []string{
	consts.LOGIC_SUBOR,
	consts.LOGIC_SUBAND,
	consts.LOGIC_SUBNOT,
	consts.LOGIC_SUBRAW,
	consts.LOGIC_NESTED,
}
var NULL_LOGIC = []string{
	consts.LOGIC_VAL_NULL,
	consts.LOGIC_NOT_NULL,
}

type Query struct {
	Logic string   `json:"logic"`           // 逻辑类型
	Field string   `json:"field"`           // 字段名称
	DType string   `json:"dtype"`           // 数据类型
	Value any      `json:"value"`           // 数据值
	Items *[]Query `json:"items,omitempty"` // 子查询
}

func BuildQuery(payload map[string]any) ([]Query, error) {
	result := []Query{}
	for key, item := range payload {
		query := Query{}
		if strings.Contains(key, ":") == false {
			key = key + ":" + consts.LOGIC_EQUALSTO
		}
		parts := strings.Split(key, ":")
		query.Value = item
		query.Field = parts[0]
		query.Logic = strings.ToUpper(parts[1])
		null_logic := slice.Contain(NULL_LOGIC, query.Logic)
		if !null_logic && (item == nil || item == "") {
			continue
		}
		// 数据类型处理
		query.DType = support.GetType(item)
		if "array" == query.DType && query.Logic == consts.LOGIC_EQUALSTO {
			query.Logic = consts.LOGIC_INCLUDES
		}
		// 跳过空查询
		switch query.Logic {
		case consts.LOGIC_INCLUDES:
			items, ok := item.([]any)
			if !ok || len(items) == 0 {
				continue
			}
		case consts.LOGIC_BETWEEN:
			items, ok := item.([]any)
			if !ok || len(items) != 2 {
				continue
			}
		}
		if query.DType == "unknown" {
			// BuildQuery 没有 Request 上下文，使用空字符串作为 logID
			support.LogErrorf("", "data type error:%s, %s", key, item)
			return nil, support.UnexpectedFormat
		}
		if query.Logic == consts.LOGIC_STR_LIKE {
			query.Value = fmt.Sprint(item)
		}
		// SUBQUERY 处理
		if slice.Contain(SubQuery, query.Logic) {
			if query.DType != "object" {
				// BuildQuery 没有 Request 上下文，使用空字符串作为 logID
				support.LogErrorf("", "data type error:%s, %s", key, item)
				return nil, support.UnexpectedFormat
			}
			subs, err := BuildQuery(item.(map[string]any))
			if err != nil {
				return nil, err
			}
			query.Items = &subs
			result = append(result, query)
			continue
		}

		result = append(result, query)
	}
	return result, nil
}

func BuildPreset(requst map[string]any, skma *Table) map[string]any {
	// 替换查询条件
	query := map[string]any{}
	for key, value := range requst {
		query[key] = value
		parts := strings.Split(key, ":")
		if len(parts) < 2 {
			parts = append(parts, "")
		}
		uukey, logic := parts[0], parts[1]
		logic = strings.ToUpper(logic)

		// 子查询(如 ITEMS:OR)需要递归处理其内部的 RCT 等预设值，
		// 否则 expires.*:RCT 只会停留在字符串层面，无法被转换为 BTW(Between)。
		if slice.Contain(SubQuery, logic) {
			switch v := value.(type) {
			case map[string]any:
				query[key] = BuildPreset(v, skma)
			}
			continue
		}

		var termField *Field
		if field := skma.GetField(uukey); field == nil {
			continue
		} else if field.FType != consts.FTYPE_DATETIME {
			continue
		} else if field.HasTimeTerm() {
			termField = field
		}
		// 特殊处理:1, 固定日期(YYYY-MM-DD)
		if logic == consts.LOGIC_INCLUDES || logic == "" {
			if d := support.ParseDate(value); d != nil {
				delete(query, key)
				key = uukey + ":" + consts.LOGIC_BETWEEN
				n, fmt := d.AddDate(0, 0, 1), time.DateOnly
				query[key] = []any{d.Format(fmt), n.Format(fmt)}
			}
			continue
		}
		// 特殊处理:2, 范围查询
		if logic == consts.LOGIC_BETWEEN && IsDateRange(value) {
			start := support.ParseDate(value.([]any)[0])
			finish := support.ParseDate(value.([]any)[1])
			if start == nil || finish == nil {
				continue
			}
			// 结束时间加1天
			query[key] = []any{
				start.AddDate(0, 0, 0).Format(time.DateOnly),
				finish.AddDate(0, 0, 1).Format(time.DateOnly),
			}
			continue
		}
		// 特殊处理:3, 按期/周
		if slice.Contain([]string{"TERM", "WEEK"}, logic) {
			var key = consts.FIELD_UTERM
			if logic == "WEEK" {
				key = consts.FIELD_UWEEK
			}
			query[key] = value
		}
		// 特殊处理:3, 最近/未来日期
		if logic == consts.LOGIC_RECENT || logic == consts.LOGIC_FUTURE {
			delete(query, key)
			key = uukey + ":" + consts.LOGIC_BETWEEN

			t := time.Now()
			start := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
			finish := time.Date(t.Year(), t.Month(), t.Day(), 24, 0, 0, 0, t.Location())
			// duration := time.Now().Local().UnixMilli() % (time.Hour.Milliseconds() * 24)
			// start := time.Now().Add(-time.Duration(duration * time.Millisecond.Nanoseconds()))
			// finish := time.Now().Add(-time.Duration(duration * time.Millisecond.Nanoseconds()))

			// 特殊处理:4, 按期
			switch v, _ := value.(string); v {
			case "CURRENT_TERM": // 本期
				query[consts.FIELD_UTERM], _ = termField.GetTermWeek(&t)
				continue
			case "PREVIOUS_TERM": // 上期
				t = t.AddDate(0, -1, 0)
				query[consts.FIELD_UTERM], _ = termField.GetTermWeek(&t)
				continue
			}

			switch v, _ := value.(string); v {
			case "CURRENT_WEEK": // 本周
				wd := int(t.Weekday())
				if wd == 0 {
					wd = 7
				}
				start = start.AddDate(0, 0, -(wd - 1))
				finish = start.AddDate(0, 0, 7)
			case "PREVIOUS_WEEK": // 上周
				wd := int(t.Weekday())
				if wd == 0 {
					wd = 7
				}
				currWeekStart := start.AddDate(0, 0, -(wd - 1))
				prevWeekStart := currWeekStart.AddDate(0, 0, -7)
				start = prevWeekStart
				finish = prevWeekStart.AddDate(0, 0, 7)
			case "CURRENT_MONTH": // 本月
				start = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
				finish = start.AddDate(0, 1, 0)
			case "PREVIOUS_MONTH": // 上月
				start = time.Date(t.Year(), t.Month()-1, 1, 0, 0, 0, 0, t.Location())
				finish = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
			case "CURRENT_YEAR": // 今年
				start = time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
				finish = start.AddDate(1, 0, 0)
			case "PREVIOUS_YEAR": // 去年
				start = time.Date(t.Year()-1, 1, 1, 0, 0, 0, 0, t.Location())
				finish = time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
			case "RECENT_3_DAYS":
				start = start.AddDate(0, 0, -3)
				finish = finish.AddDate(0, 0, 1)
			case "RECENT_1_WEEK":
				start = start.AddDate(0, 0, -7)
				finish = finish.AddDate(0, 0, 1)
			case "RECENT_2_WEEK":
				start = start.AddDate(0, 0, -14)
				finish = finish.AddDate(0, 0, 1)
			case "RECENT_1_MONTH":
				start = start.AddDate(0, -1, 0)
				finish = finish.AddDate(0, 0, 1)
			case "RECENT_3_MONTH":
				start = start.AddDate(0, -3, 0)
				finish = finish.AddDate(0, 0, 1)
			case "RECENT_6_MONTH":
				start = start.AddDate(0, -6, 0)
				finish = finish.AddDate(0, 0, 1)
			case "FUTURE_1_WEEK":
				finish = finish.AddDate(0, 0, 7)
			case "FUTURE_2_WEEK":
				finish = finish.AddDate(0, 0, 14)
			case "FUTURE_1_MONTH":
				finish = finish.AddDate(0, 1, 0)
			case "FUTURE_3_MONTH":
				finish = finish.AddDate(0, 3, 0)
			case "FUTURE_6_MONTH":
				finish = finish.AddDate(0, 6, 0)
			}
			// query[key] = []any{start, finish}
			query[key] = []any{
				start.Format(time.RFC3339),
				finish.Format(time.RFC3339),
			}
			continue
		}
	}
	return query
}

func IsDateRange(data any) bool {
	date, ok := data.([]any)
	if !ok {
		return false
	}
	layout := time.DateOnly // YYYY-MM-DD
	if len(date) != 2 {
		return false
	}
	for _, item := range date {
		if item == nil || item == "" {
			return false
		}
		str, ok := item.(string)
		if !ok || str == "" {
			return false
		}
		ret, err := time.Parse(layout, str)
		if err != nil {
			return false
		}
		if ret.IsZero() {
			return false
		}
	}
	return true
}
