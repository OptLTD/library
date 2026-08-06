package postgres

import (
	"fmt"
	"strings"

	"gorm.io/gorm/clause"
)

// fieldSQLExpr maps schema field keys to SQL column/expression.
// basic.x -> x; income.freight -> (income::jsonb)->>'freight'
func fieldSQLExpr(field string) string {
	if strings.HasPrefix(field, "basic.") {
		return quoteIdent(strings.TrimPrefix(field, "basic."))
	}
	if i := strings.Index(field, "."); i > 0 {
		group := field[:i]
		sub := field[i+1:]
		if group != "basic" {
			return fmt.Sprintf(`(%s::jsonb)->>'%s'`, quoteIdent(group), strings.ReplaceAll(sub, "'", "''"))
		}
	}
	return quoteIdent(field)
}

func fieldAlias(index string) string {
	return strings.ReplaceAll(index, ".", "_")
}

func isExprColumn(col string) bool {
	return strings.Contains(col, "::jsonb") || strings.Contains(col, "(")
}

// jsonb ->> returns text; numeric compare/bind must cast or pgx fails with
// "unable to encode <float> into text format for text (OID 25)".
func isJSONTextExpr(col string) bool {
	return strings.Contains(col, "->>")
}

func asNumericExpr(col string) string {
	if isJSONTextExpr(col) {
		return fmt.Sprintf("((NULLIF(%s, ''))::double precision)", col)
	}
	return quoteCol(col)
}

func isNumericValue(v any) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return true
	default:
		return false
	}
}

func aggrSQL(fn, expr string) string {
	switch strings.ToUpper(fn) {
	case "CNT":
		if expr == "" || expr == "*" {
			return "COUNT(*)"
		}
		return fmt.Sprintf("COUNT(%s)", expr)
	case "SUM":
		return fmt.Sprintf("SUM((%s)::double precision)", expr)
	case "AVG":
		return fmt.Sprintf("AVG((%s)::double precision)", expr)
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

func quoteIdent(col string) string {
	if isExprColumn(col) {
		return col
	}
	// Already quoted?
	if strings.HasPrefix(col, `"`) {
		return col
	}
	return fmt.Sprintf(`"%s"`, col)
}

func quoteCol(col string) string {
	return quoteIdent(col)
}

func eqClause(col string, value any) clause.Expression {
	if isJSONTextExpr(col) && isNumericValue(value) {
		return clause.Expr{SQL: fmt.Sprintf("%s = ?", asNumericExpr(col)), Vars: []any{value}}
	}
	if isExprColumn(col) {
		return clause.Expr{SQL: fmt.Sprintf("%s = ?", col), Vars: []any{value}}
	}
	return clause.Eq{Column: strings.Trim(col, `"`), Value: value}
}

func neClause(col string, value any) clause.Expression {
	if isJSONTextExpr(col) && isNumericValue(value) {
		return clause.Expr{SQL: fmt.Sprintf("%s <> ?", asNumericExpr(col)), Vars: []any{value}}
	}
	return clause.Expr{SQL: fmt.Sprintf("%s <> ?", quoteCol(col)), Vars: []any{value}}
}

func nilClause(col string) clause.Expression {
	q := quoteCol(col)
	return clause.Expr{SQL: fmt.Sprintf("(%s IS NULL OR %s::text = '')", q, q), Vars: []any{}}
}

func nnlClause(col string) clause.Expression {
	q := quoteCol(col)
	return clause.Expr{SQL: fmt.Sprintf("(%s IS NOT NULL AND %s::text <> '')", q, q), Vars: []any{}}
}

func likeClause(col string, value any) clause.Expression {
	if isExprColumn(col) {
		return clause.Expr{SQL: fmt.Sprintf("%s LIKE ?", col), Vars: []any{value}}
	}
	return clause.Like{Column: strings.Trim(col, `"`), Value: value}
}

func inClause(col string, values []any) clause.Expression {
	if isJSONTextExpr(col) && len(values) > 0 && isNumericValue(values[0]) {
		placeholders := strings.Repeat("?,", len(values))
		placeholders = strings.TrimSuffix(placeholders, ",")
		return clause.Expr{SQL: fmt.Sprintf("%s IN (%s)", asNumericExpr(col), placeholders), Vars: values}
	}
	if isExprColumn(col) {
		placeholders := strings.Repeat("?,", len(values))
		placeholders = strings.TrimSuffix(placeholders, ",")
		return clause.Expr{SQL: fmt.Sprintf("%s IN (%s)", col, placeholders), Vars: values}
	}
	return clause.IN{Column: strings.Trim(col, `"`), Values: values}
}

func cmpClause(col, op string, value any) clause.Expression {
	left := quoteCol(col)
	if isJSONTextExpr(col) {
		left = asNumericExpr(col)
	}
	return clause.Expr{SQL: fmt.Sprintf("%s %s ?", left, op), Vars: []any{value}}
}
