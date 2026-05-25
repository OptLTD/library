package source

type Value struct {
	Model Model  `json:"model"` // model name
	Scene string `json:"scene"` // source scene

	Fields map[string]Field `json:"fields"` // 逻辑类型
	Groups map[string]Group `json:"groups"` // 逻辑类型
	Tables map[string]Table `json:"tables"` // 表格配置
	Inputs map[string]Input `json:"inputs"` // 表单配置
	Clicks map[string]Click `json:"clicks"` // 表格事件
	XRules map[string]XRule `json:"xrules"` // 规则配置
}
