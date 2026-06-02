package schema

import (
	"github.com/OptLTD/library/search/consts"
	"github.com/OptLTD/library/search/request"
	"github.com/OptLTD/library/search/source"
	"github.com/OptLTD/library/search/support"

	"github.com/duke-git/lancet/v2/convertor"
)

type Input struct {
	Model  *Model  `json:"-"` // 模型信息
	Input  *SInput `json:"-"` // 模型信息
	Source *Source `json:"-"` // 登陆用户
	Scope  object  `json:"-"` // 查询限制

	// this fields make record unique
	Unique []string `json:"-"` // 唯一字段

	Title string `json:"title,omitempty"` // 模型信息

	Groups []Group `json:"groups,omitempty"` // 分组信息
	Fields []Field `json:"fields,omitempty"` // 字段类型
	XRules []XRule `json:"xrules,omitempty"` // 表单事件
	Clicks []Click `json:"clicks,omitempty"` // 表单按钮

	Values object `json:"values,omitempty"` // 分组信息
	Refers object `json:"refers,omitempty"` // 字典信息
	Others object `json:"-"`                // 视图扩展配置

	// 传入信息
	Account *request.Account `json:"-"`       // 登陆用户
	Request *request.Record  `json:"request"` // 请求信息
}

func (self *Input) GetRequest() *request.Record {
	return self.Request
}

func (self *Input) GetSearch() *request.Search {
	if self.Request == nil {
		return nil
	}
	return convertor.DeepClone(self.Request.AsSearch())
}

func (self *Input) GetGroups() []Group {
	return self.Groups
}

func (self *Input) GetFields() []Field {
	return self.Fields
}

func (self *Input) GetRefers() []Refer {
	result := []Refer{}

	for _, group := range self.Groups {
		if group.Extra.Relation != "" {
			result = append(result, Refer{
				UUKey: group.Extra.Relation,
				Using: group.Extra.Relation,
				KeyBy: consts.FIELD_UUKEY,
			})
		}
	}

	for _, field := range self.Fields {
		if field.Refer == nil {
			continue
		}
		// optional,workflow 需要被排除
		if field.FType != consts.FTYPE_RELATION {
			continue
		}
		// refer := ParseRefer(field.Refer)
		result = append(result, *field.Refer)
	}
	return result
}

func (self *Input) GetField(uukey string) *source.Field {
	for idx := range len(self.Fields) {
		if self.Fields[idx].GetKey() == uukey {
			return &self.Fields[idx]
		}
	}
	return nil
}

func (self *Input) GetGroup(uukey string) *source.Group {
	for idx := range len(self.Groups) {
		if self.Groups[idx].UUKey == uukey {
			return &self.Groups[idx]
		}
	}
	return nil
}

// NeedRefers 与 EngineService.AttachRefers 中实际会发起关联查询的条件一致：
// 存在已展示、已配置 relation、且 Values 里该字段 UUKey 对应有效外键的 RELATION 字段。
func (self *Input) NeedRefers() bool {
	if self == nil || self.Values == nil {
		return false
	}
	for _, item := range self.Fields {
		if item.FType != consts.FTYPE_RELATION {
			continue
		}
		relation := item.Extra.Relation
		if relation == "" && item.Refer != nil {
			relation = item.Refer.Using
		}
		if relation == "" || !item.Shown {
			continue
		}
		uukey, ok := self.Values[item.UUKey].(string)
		if !ok || !support.Bool(uukey) {
			continue
		}
		return true
	}
	return false
}

// GetRelation 按 relation 字段在 Refers 中查找与 value 匹配的一条关联对象（扁平 map），与 Table.GetPlain 同源逻辑。
// uukey 为字段 key（如 basic.team_id）；value 为 Combine 里该字段存的外键值（与 refer.KeyBy 对应）。
// 未找到时返回 nil。
func (self *Input) GetRelation(uukey string, value any) object {
	if !support.Bool(value) {
		return nil
	}
	field := self.GetField(uukey)
	if field == nil {
		return nil
	}
	switch field.FType {
	case consts.FTYPE_RELATION:
	case consts.FTYPE_OPTIONAL:
	case consts.FTYPE_WORKFLOW:
	default:
		return nil
	}
	refer := field.GetRefer()
	if refer == nil || refer.Using == "" {
		return nil
	}

	list, _ := self.Refers[refer.Using].([]object)
	if refer.Using == "" || list == nil {
		return nil
	}

	for _, item := range list {
		if get, ok := item[refer.KeyBy]; !ok {
			continue
		} else if get != value {
			continue
		}
		return item
	}
	return nil
}
