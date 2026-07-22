package sqlite

import (
	"fmt"
	"strings"

	"gorm.io/gorm/clause"
)

// fieldSQLExpr maps schema field keys to SQL column/expression.
// basic.x -> x; income.freight -> json_extract(income, '$.freight').
func fieldSQLExpr(field string) string {
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
