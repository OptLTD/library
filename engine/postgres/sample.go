package postgres

import (
	"fmt"
	"strings"

	"github.com/OptLTD/library/search/consts"
	"github.com/OptLTD/library/search/schema"
)

// Sample returns per-field distinct values under the table's query (data sampling
// for Excel-style filters). limit caps values per field (default 100, max 500).
func (self *Engine) Sample(table *schema.Table, fields []string, limit int) (map[string][]string, error) {
	if table == nil || len(fields) == 0 {
		return map[string][]string{}, nil
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	merged := table.BuildQuery()
	queries := self.buildQuery(consts.LOGIC_SUBAND, &merged)
	out := make(map[string][]string, len(fields))
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		expr := fieldSQLExpr(f)
		rows := []map[string]any{}
		db := self.client.Table(table.Model.Search).Clauses(queries...)
		sel := fmt.Sprintf("(%s)::text AS v", expr)
		where := fmt.Sprintf("(%s) IS NOT NULL AND btrim((%s)::text) <> ''", expr, expr)
		groupExpr := fmt.Sprintf("(%s)::text", expr)
		if err := db.Select(sel).Where(where).Group(groupExpr).Order(groupExpr).Limit(limit).Find(&rows).Error; err != nil {
			return nil, fmt.Errorf("sample %s: %w", f, err)
		}
		vals := make([]string, 0, len(rows))
		seen := map[string]struct{}{}
		for _, r := range rows {
			v := strings.TrimSpace(fmt.Sprint(r["v"]))
			if v == "" || v == "<nil>" {
				continue
			}
			if _, ok := seen[v]; ok {
				continue
			}
			seen[v] = struct{}{}
			vals = append(vals, v)
		}
		out[f] = vals
	}
	return out, nil
}
