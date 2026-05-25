package source

type Group struct {
	UUKey string `json:"uukey" bson:"uukey" gorm:"column:uukey"` // 数据类型
	Title string `json:"title" bson:"title" gorm:"column:title"` // 字段名称
	GType string `json:"gtype" bson:"gtype" gorm:"column:gtype"` // 查询条件
	// 补充信息
	Model string `json:"model,omitempty" bson:"model,omitempty" gorm:"column:model"`    // 数据类型
	SeqNo uint16 `json:"seqno,omitempty" bson:"seqno,omitempty" gorm:"column:seqno"`    // 显示顺序
	Extra GExtra `json:"extra,omitempty" bson:"extra,omitempty" gorm:"serializer:json"` // 关联属性
}

type GExtra struct {
	Nested bool     `json:"nested,omitempty" bson:"nested,omitempty"`
	Remark string   `json:"remark,omitempty" bson:"remark,omitempty"`
	Fields []string `json:"fields,omitempty" bson:"fields,omitempty"`

	Multiple bool   `json:"multiple,omitempty" bson:"multiple,omitempty"`
	Relation string `json:"relation,omitempty" bson:"relation,omitempty"`
}
