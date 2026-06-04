package source

type Input struct {
	UUKey  string            `json:"uukey"`  // 数据类型
	Title  string            `json:"title"`  // 字段名称
	Extra  object            `json:"extra"`  // 扩展信息
	Layout string            `json:"layout"` // 字段名称
	Preset object            `json:"preset"` // 新建默认值

	Groups []string `json:"groups"` // 字段名称
	Fields []string `json:"fields"` // 可用字段
	Hidden []string `json:"hidden"` // 隐藏字段
	Rename object   `json:"rename"` // uukey -> label
	Replace object  `json:"replace"` // uukey -> FExtra 属性
	XRules []string `json:"xrules"` // 字段名称
	Clicks []string          `json:"clicks"` // 按钮名称
}
