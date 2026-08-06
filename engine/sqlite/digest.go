package sqlite

import (
	"fmt"
	"strings"

	"github.com/OptLTD/library/search/consts"
	"github.com/OptLTD/library/search/request"
	"github.com/OptLTD/library/search/respond"
	"github.com/OptLTD/library/search/schema"
)

func (self *Engine) Digest(skma *schema.Digest) (*respond.Result, error) {
	if skma == nil || skma.Table == nil {
		return &respond.Result{}, nil
	}
	skma.Table = self.resetTable(skma.Table)
	table := skma.Table
	req := table.Request
	groups := skma.GroupBy
	aggrs := skma.CountFn
	if len(groups) == 0 && len(aggrs) == 0 {
		return self.Search(table)
	}

	merged := table.BuildQuery()
	bindQueryIndexes(table, &merged)
	queries := self.buildQuery(consts.LOGIC_SUBAND, &merged)

	selectParts := []string{}
	groupCols := []string{}
	aliasToIndex := map[string]string{}

	for _, g := range groups {
		expr := fieldExpr(table, g.Index)
		alias := fieldAlias(g.Index)
		selectParts = append(selectParts, fmt.Sprintf("%s AS %s", expr, alias))
		groupCols = append(groupCols, expr)
		aliasToIndex[alias] = g.Index
	}
	for _, a := range aggrs {
		expr := "*"
		if a.Index != "" {
			expr = fieldExpr(table, a.Index)
		}
		alias := a.Label
		if alias == "" {
			alias = fmt.Sprintf("%s_%s", strings.ToLower(a.Func), fieldAlias(a.Index))
		}
		safeAlias := fieldAlias(alias)
		selectParts = append(selectParts, fmt.Sprintf("%s AS %s", aggrSQL(a.Func, expr), safeAlias))
		aliasToIndex[safeAlias] = alias
	}
	if len(selectParts) == 0 {
		selectParts = append(selectParts, "COUNT(*) AS cnt")
		aliasToIndex["cnt"] = "cnt"
	}

	db := self.client.Table(table.Model.Search).Clauses(queries...)
	if len(groupCols) > 0 {
		db = db.Group(strings.Join(groupCols, ", "))
	}
	values := []map[string]any{}
	if err := db.Select(strings.Join(selectParts, ", ")).Find(&values).Error; err != nil {
		return nil, err
	}

	decoded := make([]map[string]any, 0, len(values))
	totals := map[string]any{}
	for _, row := range values {
		item := map[string]any{}
		for alias, val := range row {
			key := aliasToIndex[alias]
			if key == "" {
				key = alias
			}
			item[key] = val
		}
		decoded = append(decoded, item)
		for key, val := range item {
			if _, ok := totals[key]; !ok {
				totals[key] = val
				continue
			}
			totals[key] = addNumeric(totals[key], val)
		}
	}

	result := &respond.Result{
		Page:   req.Page,
		Size:   req.Size,
		Count:  uint64(len(decoded)),
		Values: decoded,
		Totals: totals,
	}
	if req.Order != nil && req.Order.Field != "" {
		result.Sort(req.Order)
	} else if len(groups) > 0 {
		result.Sort(&request.Order{
			Field: groups[0].Index,
			Order: groups[0].SortBy,
		})
	}
	for _, h := range self.handles {
		if err := h.DigestResult(skma, result); err != nil {
			return result, err
		}
	}
	return result, nil
}

func addNumeric(a, b any) any {
	switch av := a.(type) {
	case int64:
		if bv, ok := b.(int64); ok {
			return av + bv
		}
	case float64:
		if bv, ok := b.(float64); ok {
			return av + bv
		}
	}
	return b
}
