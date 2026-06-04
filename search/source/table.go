package source

type object = map[string]any

type Table struct {
	UUKey  string   `json:"uukey" bson:"uukey"`   // 数据类型
	Title  string   `json:"title" bson:"title"`   // 字段名称
	Query  object   `json:"query" bson:"query"`   // 默认筛选
	Extra  object   `json:"extra" bson:"extra"`   // 透视/合并等默认配置
	Fields []string `json:"fields" bson:"fields"` // 可用字段
	Hidden []string `json:"hidden" bson:"hidden"` // 隐藏字段
	Rename object   `json:"rename" bson:"rename"` // uukey -> label
	Replace object  `json:"replace" bson:"replace"` // uukey -> FExtra 属性
	Clicks []string `json:"clicks" bson:"clicks"` // 表格动作

	Order []Order   `json:"-" bson:"-" gorm:"-"` // 默认排序
	Pivot []GroupBy `json:"-" bson:"-" gorm:"-"` // 聚合字段
	Aggrs []CountFn `json:"-" bson:"-" gorm:"-"` // 聚合配置
}

type Order struct {
	Field string `json:"field"` // 字段名称
	Order string `json:"order"` // 排序方式
}
