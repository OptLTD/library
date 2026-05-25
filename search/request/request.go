package request

import (
	"strings"
)

type object = map[string]any
type Order struct {
	Field string `json:"field"` // 排序字段
	Order string `json:"order"` // 排序方式
}
type Search struct {
	// base info
	Model string `json:"model"` // 模型名称
	UUKey string `json:"uukey"` // 模型标示
	Scene string `json:"scene"` // 请求场景
	LogID string `json:"logid"` // logid
	// debug
	Debug string `json:"debug,omitempty"` // 调试数据

	// page info
	Page uint16 `json:"page,omitempty"` // 分页索引
	Size uint16 `json:"size,omitempty"` // 分页大小

	// omit if empty
	Using string `json:"using,omitempty"` // 查询字段
	Order *Order `json:"order,omitempty"` // 排序条件
	Query object `json:"query,omitempty"` // 查询条件
	// Value object `json:"value,omitempty"` // 请求数据
	// Batch []any  `json:"batch,omitempty"` // 批量处理
	Extra object `json:"extra,omitempty"` // 额外参数
	// 登录用户
	Login *Account `json:"-"`
}

func (self *Search) GetBase() string {
	base, _, _ := strings.Cut(self.Model, ".")
	return base
}

func (self *Search) IsInner() bool {
	switch self.GetBase() {
	case "app", "sys":
		return true
	case "team", "su":
		return true
	}
	return false
}

func NewRequest(model string, uukey string) *Search {
	request := &Search{
		Model: model,
		UUKey: uukey,
	}
	return request
}
