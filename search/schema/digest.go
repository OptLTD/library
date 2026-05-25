package schema

import (
	"search/consts"
	"search/source"
	"strings"

	"github.com/duke-git/lancet/v2/slice"
)

var BucketValues = []string{
	consts.VALUE_TERMS,
	consts.VALUE_RANGE,
	consts.VALUE_HIST1,
	consts.VALUE_HIST2,
}

type Digest struct {
	Table *Table

	Fields []Field
	Groups []Group
	Sticky []string

	PivotBy *PivotBy
	GroupBy []source.GroupBy
	CountFn []source.CountFn
}

func (self *Digest) Parse(payload object) ([]source.CountFn, error) {
	result := []source.CountFn{}
	for key, val := range payload {
		aggr := source.CountFn{}
		if strings.Contains(key, ":") == false {
			key = consts.VALUE_CNT + ":" + key
		}
		parts := strings.Split(key, ":")
		aggr.Index, aggr.Func = parts[1], strings.ToUpper(parts[0])
		// SUBQUERY 处理
		if !slice.Contain(BucketValues, aggr.Func) {
			aggr.Label = val.(string)
			result = append(result, aggr)
			continue
		}

		if child, ok := val.(map[string]any); ok {
			items, err := self.Parse(child)
			if err != nil {
				return nil, err
			}
			aggr.Items = items
			aggr.Label = strings.ToLower(key)
			result = append(result, aggr)
		}
	}
	return result, nil
}
