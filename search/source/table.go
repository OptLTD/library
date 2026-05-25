package source

type object = map[string]any

type Table struct {
	UUKey  string   `json:"uukey"`  // 数据类型
	Title  string   `json:"title"`  // 字段名称
	Query  object   `json:"query"`  // 默认筛选
	Fields []string `json:"fields"` // 可用字段
	Hidden []string `json:"hidden"` // 可用字段
	Clicks []string `json:"clicks"` // 表格动作

	Order []Order   `json:"-" bson:"-" gorm:"-"` // 默认排序
	Pivot []GroupBy `json:"-" bson:"-" gorm:"-"` // 聚合字段
	Aggrs []CountFn `json:"-" bson:"-" gorm:"-"` // 聚合配置
}

type Order struct {
	Field string `json:"field"` // 字段名称
	Order string `json:"order"` // 排序方式
}
