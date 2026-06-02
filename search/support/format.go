package support

import (
	"encoding/base64"
	"encoding/json"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	parser "github.com/buger/jsonparser"
)

func ParseDate(val any) *time.Time {
	datetime, ok := val.(string)
	if !ok || datetime == "" {
		return nil
	}
	layout := time.RFC3339
	if !strings.Contains(datetime, "T") {
		layout = time.DateTime
	}
	ret, err := time.Parse(layout, datetime)
	if err == nil && !ret.IsZero() {
		return &ret
	}

	layout = time.DateOnly // YYYY-MM-DD
	ret, err = time.Parse(layout, datetime)
	if err == nil && !ret.IsZero() {
		return &ret
	}
	return nil
}

func ParseNumber(val any) float64 {
	num, _ := strconv.ParseFloat(val.(string), 64)
	return num
}

func IsInvalidUUKey(key string) bool {
	matcher := regexp.MustCompile(`^[0-9a-zA-Z\-\_]+$`)
	return matcher.MatchString(key) == false
}

func Contains(array []string, target string) bool {
	sort.Strings(array)
	index := sort.SearchStrings(array, target)
	//index的取值：[0,len(str_array)]
	//需要注意此处的判断，先判断 &&左侧的条件，如果不满足则结束此处判断，不会再进行右侧的判断
	if index < len(array) && array[index] == target {
		return true
	}
	return false
}

func GetVal(value []byte, dataType parser.ValueType) any {
	switch dataType {
	case parser.Boolean:
		if b, err := parser.ParseBoolean(value); err == nil {
			return b
		}
		return string(value) == "true"
	case parser.Object:
		var obj map[string]any
		if json.Unmarshal(value, &obj) == nil {
			return obj
		}
		return nil
	case parser.Array:
		var arr []any
		if json.Unmarshal(value, &arr) == nil {
			return arr
		}
		return nil
	case parser.String:
		return string(value)
	case parser.Number:
		intval, _ := parser.ParseInt(value)
		dblval, _ := parser.ParseFloat(value)
		if dblval != float64(intval) {
			return dblval
		} else {
			return intval
		}
	}
	return nil
}

// NormalizeQueryValue 修正 query 条件值：JSON 数组/对象、[]byte、以及历史遗留的 base64 编码 JSON。
func NormalizeQueryValue(val any) any {
	if val == nil {
		return val
	}
	switch v := val.(type) {
	case []byte:
		var parsed any
		if json.Unmarshal(v, &parsed) == nil {
			return parsed
		}
		return v
	case string:
		if decoded, err := base64.StdEncoding.DecodeString(v); err == nil {
			var parsed any
			if json.Unmarshal(decoded, &parsed) == nil {
				switch parsed.(type) {
				case []any, map[string]any:
					return parsed
				}
			}
		}
		return v
	default:
		return val
	}
}

// NormalizeQueryObject 规范化 table.query / 默认筛选 map。
func NormalizeQueryObject(query map[string]any) map[string]any {
	if len(query) == 0 {
		return query
	}
	flat := map[string]any{}
	data, err := json.Marshal(query)
	if err != nil {
		return query
	}
	if err := json.Unmarshal(data, &flat); err != nil {
		return query
	}
	for key, val := range flat {
		flat[key] = NormalizeQueryValue(val)
	}
	return flat
}

func GetType(val any) string {
	getType := reflect.TypeOf(val)
	if getType == nil {
		return "unknown"
	}
	switch getType.Kind() {
	case reflect.Bool:
		return "bool"
	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		fallthrough
	case reflect.Uint:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		return "number"
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		fallthrough
	case reflect.Complex64:
		fallthrough
	case reflect.Complex128:
		return "float"
	case reflect.String:
		return "string"
	case reflect.Slice:
		return "array"
	case reflect.Map:
		return "object"
	default:
		return "unknown"
	}
}
