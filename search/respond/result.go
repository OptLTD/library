package respond

import (
	"search/request"
	"search/support"
	"slices"
	"strings"
)

type object = map[string]any

type Result struct {
	Page  uint16 `json:"page"`
	Size  uint16 `json:"size"`
	Count uint64 `json:"count"`

	Refers object `json:"refers"` // 引用数据
	Totals object `json:"totals"` // 汇总数据

	Values []object `json:"values"` // 返回数据
}

func (self *Result) Sort(order *request.Order) {
	if len(self.Values) == 0 || order == nil {
		return
	}
	if order.Field == "" || order.Order == "" {
		return
	}
	isDesc, field := false, order.Field
	switch strings.ToUpper(order.Order) {
	case "DESC", "DESCENDING":
		isDesc = true
	}
	slices.SortFunc(self.Values, func(a, b object) int {
		valA, okA := a[field]
		valB, okB := b[field]
		if !okA && !okB {
			return 0
		}
		if !okA {
			return 1
		}
		if !okB {
			return -1
		}
		return support.Compare(valA, valB, isDesc)
	})
}
