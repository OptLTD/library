package sqlite

// Search Index convention (no in-process registry):
//
//	Field / UUKey — storage path (JSON key or flatten column name)
//	Field.Index   — search path; when Index != UUKey it is a heterogeneous index
//	                (physical column such as gen_income_total, or flatten col).
//
// Apps (worth) create STORED columns and set Field.Index before query.
// Engine only consumes Index via fieldExpr — never invents gen_* names.
