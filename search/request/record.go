package request

import "strings"

type Record struct {
	// base info
	Model string `json:"model"` // 模型名称
	UUKey string `json:"uukey"` // 模型标示
	Scene string `json:"scene"` // 请求场景
	LogID string `json:"logid"` // logid
	// debug
	Debug string `json:"debug,omitempty"` // 调试数据
	Using string `json:"using,omitempty"` // 使用配置

	Value object `json:"value,omitempty"` // 请求数据
	Batch []any  `json:"batch,omitempty"` // 批量处理
	Extra object `json:"extra,omitempty"` // 额外参数
	// 登录用户
	Login *Account `json:"-"`
}

func (self *Record) GetBase() string {
	base, _, _ := strings.Cut(self.Model, ".")
	return base
}

func (self *Record) IsInner() bool {
	switch self.GetBase() {
	case "app", "sys":
		return true
	case "team", "su":
		return true
	}
	return false
}

// AsSearch 将 Record 中与加载元数据、插件、bundle 相关的字段投影为 Search，供 GetSource 等仍基于 Search 的路径使用；不携带 Value/Batch。
func (self *Record) AsSearch() *Search {
	if self == nil {
		return nil
	}
	return &Search{
		Model: self.Model,
		UUKey: self.UUKey,
		Scene: self.Scene,
		LogID: self.LogID,
		Debug: self.Debug,
		Using: self.Using,
		Extra: self.Extra,
		Login: self.Login,
		Page:  1,
		Size:  10000,
	}
}
