package schema

import "search/request"

type ISchema interface {
	GetGroups() []Group
	GetFields() []Field
	GetRefers() []Refer
	GetSearch() *request.Search
	GetField(key string) *Field
}
