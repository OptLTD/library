package sqlite

import (
	"fmt"
	"strings"

	"github.com/OptLTD/library/search/consts"
	"github.com/OptLTD/library/search/schema"
	"github.com/duke-git/lancet/v2/convertor"
	"gorm.io/gorm/clause"
)

// fieldSQLExpr maps a logical path to SQL when no heterogeneous Index is set.
// Physical column names (no ".") pass through; basic.x → column; group.field → json_extract.
// App (e.g. worth) sets Field.Index to gen_* for STORED columns — see fieldExpr.
func fieldSQLExpr(field string) string {
	if field == "" {
		return field
	}
	if !strings.Contains(field, ".") {
		return field
	}
	if strings.HasPrefix(field, "basic.") {
		return strings.TrimPrefix(field, "basic.")
	}
	if i := strings.Index(field, "."); i > 0 {
		group := field[:i]
		sub := field[i+1:]
		if group != "basic" {
			return fmt.Sprintf("json_extract(%s, '$.%s')", group, sub)
		}
	}
	return field
}

// querySQLExpr returns the SQL column/expression for a filter clause.
// Uses Query.Index when already bound; otherwise falls back to Field.
func querySQLExpr(q schema.Query) string {
	if q.Index != "" {
		return q.Index
	}
	return fieldSQLExpr(q.Field)
}

// bindQueryIndexes fills Query.Index from schema Field.Index (store vs search).
// Call after resetTable / ApplySearchIndex so Index is the physical column or JSON path.
func bindQueryIndexes(skma *schema.Table, queries *[]schema.Query) {
	if queries == nil {
		return
	}
	for i := range *queries {
		q := &(*queries)[i]
		if q.Items != nil && len(*q.Items) > 0 {
			bindQueryIndexes(skma, q.Items)
		}
		if q.Field == "" {
			continue
		}
		logic := strings.ToUpper(q.Logic)
		if logic == consts.LOGIC_SUBOR || logic == consts.LOGIC_SUBAND ||
			logic == consts.LOGIC_SUBNOT || logic == consts.LOGIC_SUBRAW ||
			logic == consts.LOGIC_NESTED {
			continue
		}
		q.Index = fieldExpr(skma, q.Field)
	}
}

// fieldExpr resolves logical key (usually UUKey) via schema Field.Index when set.
// Index != UUKey → heterogeneous search path (physical col / gen_* / expression); use as-is.
// Index == UUKey → same as store path; map via fieldSQLExpr (json_extract / flatten).
func fieldExpr(skma *schema.Table, key string) string {
	if skma != nil && key != "" {
		if f := skma.GetField(key); f != nil && f.Index != "" {
			logical := f.UUKey
			if logical == "" {
				logical = f.GetKey()
			}
			if f.Index != logical {
				return f.Index
			}
			return fieldSQLExpr(logical)
		}
	}
	return fieldSQLExpr(key)
}

// resetTable clones schema and normalizes Index for search (store vs index).
// FLATTEN → physical column (Field). GROUPED keeps app-set hetero Index (e.g. gen_*),
// otherwise Index stays / becomes UUKey for json_extract.
func (self *Engine) resetTable(skma *schema.Table) *schema.Table {
	if skma == nil || len(skma.Fields) == 0 {
		return skma
	}
	clone := convertor.DeepClone(skma)
	ApplySearchIndex(clone)
	return clone
}

// ApplySearchIndex sets default search Index in place.
// Does not invent gen_* — apps set heterogeneous Index before query.
func ApplySearchIndex(skma *schema.Table) {
	if skma == nil {
		return
	}
	for i := range skma.Fields {
		applyFieldSearchIndex(&skma.Fields[i])
	}
}

func applyFieldSearchIndex(f *schema.Field) {
	if f == nil {
		return
	}
	logical := f.UUKey
	if logical == "" {
		logical = f.GetKey()
	}
	// Preserve app-set heterogeneous Index (gen_*, flatten col, expression).
	if f.Index != "" && f.Index != logical {
		return
	}
	flatten := strings.EqualFold(f.GType, consts.GTYPE_FLATTEN) ||
		strings.EqualFold(f.Group, "basic")
	if flatten {
		if f.Field != "" {
			f.Index = f.Field
		}
		return
	}
	if logical != "" {
		f.Index = logical
		return
	}
	if f.Group != "" && f.Field != "" {
		f.Index = f.Group + "." + f.Field
	}
}

func fieldAlias(index string) string {
	return strings.ReplaceAll(index, ".", "_")
}

func isExprColumn(col string) bool {
	return strings.Contains(col, "json_extract(") || strings.Contains(col, "(")
}

func aggrSQL(fn, expr string) string {
	switch strings.ToUpper(fn) {
	case "CNT":
		if expr == "" || expr == "*" {
			return "COUNT(*)"
		}
		return fmt.Sprintf("COUNT(%s)", expr)
	case "SUM":
		return fmt.Sprintf("SUM(CAST(%s AS REAL))", expr)
	case "AVG":
		return fmt.Sprintf("AVG(CAST(%s AS REAL))", expr)
	case "MAX":
		return fmt.Sprintf("MAX(%s)", expr)
	case "MIN":
		return fmt.Sprintf("MIN(%s)", expr)
	case "UNQ":
		return fmt.Sprintf("COUNT(DISTINCT %s)", expr)
	default:
		if expr == "" {
			return "COUNT(*)"
		}
		return fmt.Sprintf("COUNT(%s)", expr)
	}
}

func quoteCol(col string) string {
	if isExprColumn(col) {
		return col
	}
	return fmt.Sprintf(`"%s"`, col)
}

func eqClause(col string, value any) clause.Expression {
	if isExprColumn(col) {
		return clause.Expr{SQL: fmt.Sprintf("%s = ?", col), Vars: []any{value}}
	}
	return clause.Eq{Column: col, Value: value}
}

func neClause(col string, value any) clause.Expression {
	return clause.Expr{SQL: fmt.Sprintf("%s <> ?", quoteCol(col)), Vars: []any{value}}
}

func nilClause(col string) clause.Expression {
	q := quoteCol(col)
	return clause.Expr{SQL: fmt.Sprintf("(%s IS NULL OR %s = '')", q, q), Vars: []any{}}
}

func nnlClause(col string) clause.Expression {
	q := quoteCol(col)
	return clause.Expr{SQL: fmt.Sprintf("(%s IS NOT NULL AND %s <> '')", q, q), Vars: []any{}}
}

func likeClause(col string, value any) clause.Expression {
	if isExprColumn(col) {
		return clause.Expr{SQL: fmt.Sprintf("%s LIKE ?", col), Vars: []any{value}}
	}
	return clause.Like{Column: col, Value: value}
}

func inClause(col string, values []any) clause.Expression {
	if isExprColumn(col) {
		placeholders := strings.Repeat("?,", len(values))
		placeholders = strings.TrimSuffix(placeholders, ",")
		return clause.Expr{SQL: fmt.Sprintf("%s IN (%s)", col, placeholders), Vars: values}
	}
	return clause.IN{Column: col, Values: values}
}

func cmpClause(col, op string, value any) clause.Expression {
	return clause.Expr{SQL: fmt.Sprintf("%s %s ?", quoteCol(col), op), Vars: []any{value}}
}
