package source

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/OptLTD/library/search/consts"
	"github.com/OptLTD/library/search/support"

	"github.com/duke-git/lancet/v2/convertor"
	"github.com/duke-git/lancet/v2/slice"
)

type Field struct {
	UUKey string `json:"uukey" gorm:"column:uukey"`   // 数据类型
	GType string `json:"gtype" bson:"-" gorm:"-:all"` // group type
	GName string `json:"gname" bson:"-" gorm:"-:all"` // group name

	FType string `json:"ftype" bson:"ftype" gorm:"column:ftype"` // field type
	Group string `json:"group,omitempty" bson:"group,omitempty"` // 逻辑类型
	Field string `json:"field,omitempty" bson:"field,omitempty"` // 字段key
	Label string `json:"label,omitempty" bson:"label,omitempty"` // 字段名称
	Index string `json:"index,omitempty" bson:"index,omitempty"` // 搜索字段
	Shown bool   `json:"shown,omitempty" bson:"shown,omitempty"` // 是否显示
	Width uint16 `json:"width,omitempty" bson:"width,omitempty"` // 是否显示

	Refer  *Refer `json:"refer,omitempty" bson:"refer,omitempty"`   // 引用数据
	SeqNo  uint16 `json:"seqno,omitempty" bson:"seqno,omitempty"`   // 展示顺序
	Remark string `json:"remark,omitempty" bson:"remark,omitempty"` // 字段备注

	Extra FExtra `json:"extra,omitempty" bson:"extra,omitempty" gorm:"serializer:json"` // 关联属性
}

type FExtra struct {
	// internal & embedded & necessary
	Embedded bool `json:"embedded,omitempty" bson:"embedded,omitempty"`
	// edit props
	Required bool `json:"required,omitempty" bson:"required,omitempty"`
	Multiple bool `json:"multiple,omitempty" bson:"multiple,omitempty"`
	Implicit bool `json:"implicit,omitempty" bson:"implicit,omitempty"`
	// edit mode
	Editable string `json:"editable,omitempty" bson:"editable,omitempty"`
	Disabled string `json:"disabled,omitempty" bson:"disabled,omitempty"`
	// serialno
	DateTime string `json:"datetime,omitempty" bson:"datetime,omitempty"`
	Constant string `json:"constant,omitempty" bson:"constant,omitempty"`
	Counting int64  `json:"counting,omitempty" bson:"counting,omitempty"`
	// Formula
	Formula string `json:"formula,omitempty" bson:"formula,omitempty"`
	Display string `json:"display,omitempty" bson:"display,omitempty"`
	Depends []any  `json:"depends,omitempty" bson:"depends,omitempty"`
	// numberic
	MinValue  int64  `json:"minValue,omitempty" bson:"minValue,omitempty"`
	MaxValue  int64  `json:"maxValue,omitempty" bson:"maxValue,omitempty"`
	Precision int64  `json:"precision,omitempty" bson:"precision,omitempty"`
	RoundMode string `json:"roundMode,omitempty" bson:"roundMode,omitempty"`
	// ValueFmt 数字展示格式
	ValueFmt string `json:"valueFmt,omitempty" bson:"valueFmt,omitempty"`
	// AggrType 汇总行聚合方式
	AggrType string `json:"aggrType,omitempty" bson:"aggrType,omitempty"`
	// expense
	Expense string `json:"expense,omitempty" bson:"expense,omitempty"`
	// relation
	Relation string `json:"relation,omitempty" bson:"relation,omitempty"`
	// options
	Options []Option `json:"options,omitempty" bson:"options,omitempty"`
	DictKey string   `json:"dictKey,omitempty" bson:"dictKey,omitempty"`
	DataKey string   `json:"dataKey,omitempty" bson:"dataKey,omitempty"`
	TextKey string   `json:"textKey,omitempty" bson:"textKey,omitempty"`
	// Display string `json:"display,omitempty" bson:"display,omitempty"`
	// datetime|strings|uploads|numeric
	DataType string `json:"dataType,omitempty" bson:"dataType,omitempty"`
	// time term, NONE|26|27|28|29, 财务周期范围
	TimeTerm string `json:"timeTerm,omitempty" bson:"timeTerm,omitempty"`
	// upload
	AcceptType string `json:"acceptType,omitempty" bson:"acceptType,omitempty"`
}
type Option struct {
	UUKey  string `json:"uukey" bson:"uukey"`
	Label  string `json:"label" bson:"label"`
	Color  string `json:"color,omitempty" bson:"color,omitempty"`
	Parent string `json:"parent,omitempty" bson:"parent,omitempty"`
}

func (s Field) GetKey() string {
	return s.Group + "." + s.Field
}

func (s Field) ToMap() object {
	extra, _ := convertor.StructToMap(s.Extra)
	value := map[string]any{
		"label": s.Label, "index": s.Index,
		"field": s.Field, "group": s.Group,
		"ftype": s.FType, "extra": extra,
	}
	if s.SeqNo > 0 {
		value["seqno"] = s.SeqNo
	}
	if s.Width > 0 {
		value["width"] = s.Width
	}
	if s.Remark != "" {
		value["remark"] = s.Remark
	}
	return value
}

func (s Field) HasTimeTerm() bool {
	ignore := []string{"NONE", "none", "", "0"}
	return !slice.Contain(ignore, s.Extra.TimeTerm)
}

func (s Field) GetTimeTerm() int {
	ignore := []string{"NONE", "none", "", "0"}
	if slice.Contain(ignore, s.Extra.TimeTerm) {
		return 0
	}
	val, err := strconv.Atoi(s.Extra.TimeTerm)
	if err != nil {
		return 0
	}
	return val
}

func (s Field) GetTermWeek(date *time.Time) (int, int) {
	start := s.GetTimeTerm()
	t := date.In(time.Local)
	if start < 1 || start > 28 {
		start = 26
	}
	isoYear, isoWeek := t.ISOWeek()
	uweek := isoYear*100 + isoWeek

	y, m, d := t.Date()
	var termYear int
	var termMonth time.Month
	if d < start {
		termYear, termMonth = y, m
	} else {
		termYear = y
		termMonth = m + 1
		if termMonth > time.December {
			termYear++
			termMonth = time.January
		}
	}
	uterm := termYear*100 + int(termMonth)
	return uterm, uweek
}

func (s Field) GetDataType() string {
	dataType := s.Extra.DataType
	switch strings.ToUpper(s.FType) {
	case consts.FTYPE_STRINGS:
		return support.Or(dataType, consts.DTYPE_KEYWORDS)
	case consts.FTYPE_NUMERIC:
		if s.Extra.Precision > 0 {
			return consts.DTYPE_DECIMAL
		}
		return support.Or(dataType, consts.DTYPE_INTEGER)
	case consts.FTYPE_DATETIME:
		return support.Or(dataType, consts.DTYPE_DATETIME)
	case consts.FTYPE_UPLOADS:
		return support.Or(dataType, consts.DTYPE_IMG_FILE)
	case consts.FTYPE_SOCIALS:
		return support.Or(dataType, consts.DTYPE_X_EMAIL)
	default:
		return s.FType
	}
}

func (s Field) GetFieldType() string {
	dataType := s.Extra.DataType
	if dataType == "" {
		return s.FType
	}
	switch strings.ToUpper(dataType) {
	case consts.DTYPE_LONGTEXT, consts.DTYPE_RICHTEXT,
		consts.DTYPE_KEYWORDS:
		return consts.FTYPE_STRINGS
	case consts.DTYPE_INTEGER, consts.DTYPE_DECIMAL,
		consts.DTYPE_LONGINT, consts.DTYPE_SCALED:
		return consts.FTYPE_NUMERIC
	case consts.DTYPE_ONLYDATE, consts.DTYPE_DATETIME:
		return consts.FTYPE_DATETIME
	case consts.DTYPE_IMG_FILE, consts.DTYPE_DOC_FILE,
		consts.DTYPE_ALL_FILE:
		return consts.FTYPE_UPLOADS
	case consts.DTYPE_X_EMAIL, consts.DTYPE_X_PHONE:
		return consts.FTYPE_SOCIALS
	case consts.DTYPE_OPTIONAL,
		consts.DTYPE_WORKFLOW,
		consts.DTYPE_RELATION:
		return dataType
	case consts.DTYPE_EXPENSE:
		return consts.FTYPE_EXPENSE
	default:
		return s.FType
	}
}

func (s Field) GetRefer() *Refer {
	switch s.FType {
	case consts.FTYPE_RELATION:
		refer := support.Or(s.Refer, &Refer{
			// todo text by
			KeyBy: consts.FIELD_UUKEY,
			TxtBy: s.Extra.TextKey,
			Using: s.Extra.Relation,
		})
		// default key by
		if s.Extra.DataKey != "" {
			refer.KeyBy = s.Extra.DataKey
		}
		// if refer.TxtBy == "" {
		// 	self.GetReferSubject(skma, refer)
		// }
		return refer
	case consts.FTYPE_OPTIONAL:
		refer := &Refer{
			KeyBy: "uukey",
			TxtBy: "label",
			Using: s.UUKey,
		}
		dictKey := s.Extra.DictKey
		if strings.HasPrefix(dictKey, "global:") {
			if key := dictKey[7:]; key != "" {
				refer.Using = key
				s.Extra.Options = nil
			}
		}
		if txt := s.Extra.TextKey; txt != "" {
			refer.TxtBy = txt
		}
		return refer
	case consts.FTYPE_WORKFLOW:
		return &Refer{
			KeyBy: "uukey",
			TxtBy: "label",
			Using: s.UUKey,
		}
	}
	return nil
}

func (s Field) Parse(get any) any {
	switch s.FType {
	case consts.FTYPE_DATETIME:
		switch get := get.(type) {
		case time.Time:
			return &get
		case *time.Time:
			return get
		case string:
			return s.ParseTime(get)
		case int32, int64, float32, float64:
			val, err := strconv.ParseInt(fmt.Sprint(get), 10, 64)
			if err == nil || val > 0 {
				v := time.UnixMicro(val)
				return &v
			} else {
				return nil
			}
		default:
			if v, ok := parseMongoDateTime(get); ok {
				return &v
			}
		}
	case consts.FTYPE_NUMERIC, consts.DTYPE_EXPENSE:
		// null\0\undefinded, reset value
		if !support.Bool(get) {
			return nil
		}
		final, ok := float64(0), true
		switch get := get.(type) {
		case []any:
			var result = make([]float64, 0)
			for _, item := range get {
				if get := s.Parse(item); get == nil {
					continue
				} else if get, ok := get.(float64); ok {
					result = append(result, get)
				}
			}
			return result
		case string:
			var err error
			numberStr := strings.ReplaceAll(get, ",", "")
			if final, err = strconv.ParseFloat(numberStr, 64); err != nil {
				ok = false
			}

		case int:
			final = float64(get)
		case uint:
			final = float64(get)
		case uint32:
			final = float64(get)
		case uint64:
			final = float64(get)
		case int32:
			final = float64(get)
		case int64:
			final = float64(get)
		case float32:
			final = float64(get)
		case float64:
			final = get
		}
		if !ok || get == 0 {
			return nil
		}
		if s.Extra.Precision == 0 {
			return final
		}
		switch s.Extra.RoundMode {
		case "ceil":
			d := math.Pow10(int(s.Extra.Precision))
			return math.Ceil(final*d) / d
		case "floor":
			d := math.Pow10(int(s.Extra.Precision))
			return math.Floor(final*d) / d
		case "round":
			d := math.Pow10(int(s.Extra.Precision))
			return math.Round(final*d) / d
		default:
			d := math.Pow10(int(s.Extra.Precision))
			return math.Round(final*d) / d
		}
	default:
		return get
	}
	return nil
}

func (s Field) Equal(old, get any) bool {
	switch s.FType {
	case consts.FTYPE_DATETIME:
		o, _ := s.Parse(old).(*time.Time)
		n, _ := s.Parse(get).(*time.Time)
		if o == nil || n == nil {
			return o == n
		}
		return o.Unix() == n.Unix()
	case consts.FTYPE_NUMERIC, consts.DTYPE_EXPENSE:
		return s.Parse(old) == s.Parse(get)
	}
	return reflect.DeepEqual(old, get)
}

func (s Field) ParseTime(get string) *time.Time {
	switch s.Extra.DataType {
	case "ONLYDATE", "ONLYMONTH":
		if strings.Contains(get, "00:00:00") {
			return support.ParseDate(get)
		}
		if strings.Contains(get, "00:00.000") {
			return support.ParseDate(get)
		}
		layout := time.DateOnly
		ret, err := time.Parse(layout, get)
		if err == nil {
			return &ret
		}
		// 兼容处理, 再尝试一次, 把日期当作日期时间解析
		if v := support.ParseDate(get); v != nil {
			return v
		}
	case "DATETIME":
		return support.ParseDate(get)
	default:
		return support.ParseDate(get)
	}
	return nil
}

func parseMongoDateTime(get any) (time.Time, bool) {
	if get == nil {
		return time.Time{}, false
	}
	val := reflect.ValueOf(get)
	for val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return time.Time{}, false
		}
		val = val.Elem()
	}
	if val.Type().PkgPath() == "go.mongodb.org/mongo-driver/bson/primitive" &&
		val.Type().Name() == "DateTime" {
		if method := val.MethodByName("Time"); method.IsValid() && method.Type().NumIn() == 0 {
			results := method.Call(nil)
			if len(results) == 1 {
				if t, ok := results[0].Interface().(time.Time); ok {
					return t, true
				}
			}
		}
	}
	return time.Time{}, false
}
