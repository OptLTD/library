package support

import (
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
		return value
	case parser.Object:
		return value
	case parser.Array:
		return value
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
