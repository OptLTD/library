package source

type XRule struct {
	Title string `json:"title"`
	Model string `json:"model"`
	UUKey string `json:"uukey"`

	SeqNo int16    `json:"seqno"`
	Scope []string `json:"scope"`

	XWhen []XRuleWhen `json:"xwhen"`
	XThen []XRuleThen `json:"xthen"`
	XElse []XRuleThen `json:"xelse"`
}

type XRuleWhen struct {
	Logic  string `json:"logic"`
	Value  string `json:"value"`
	Target string `json:"target"`
}

type XRuleThen struct {
	Action string `json:"action"`
	Target string `json:"target"`
	TgType string `json:"tgtype"`
	Values any    `json:"values"`
}
