package source

type Model struct {
	UUKey  string `json:"uukey" gorm:"column:uukey"`   // 页面标识
	Title  string `json:"title" gorm:"column:title"`   // 页面标题
	Brief  string `json:"brief" gorm:"column:brief"`   // 页面描述
	Driver string `json:"driver" gorm:"column:driver"` // 驱动类型
	Source string `json:"source" gorm:"column:source"` // 存储数据源
	Search string `json:"search" gorm:"column:search"` // 索引数据源

	Groups []Group `json:"-" gorm:"serializer:json"` // 分组
	Fields []Field `json:"-" gorm:"serializer:json"` // 字段
	Clicks []Click `json:"-" gorm:"serializer:json"` // 按钮

	Extra object `json:"-" gorm:"serializer:json"` // 关联属性
}
