package source

type CountFn struct {
	Label string `json:"label"` // 取值名称
	Index string `json:"index"` // 字段名称
	Func  string `json:"func"`  // 取值方式

	Option object `json:"option"` // 聚合选项

	Items []CountFn `json:"items,omitempty"` // 子查询
}

type GroupBy struct {
	Index  string `json:"index"`  // 字段名称
	SortBy string `json:"sortby"` // 排序方式
	Format string `json:"format"` // 格式化方式
}

type PivotBy struct {
	Pivot string `json:"pivot"` // 透视字段
	Value string `json:"value"` // 值字段
	State string `json:"state"` // 状态: active, none
}
