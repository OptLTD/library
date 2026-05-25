package formula

import "regexp"

// excelIfCallPattern 只匹配「作为独立标识符的 if」后的左括号（允许 if 与 ( 之间有空格）。
// 词边界 \b 保证 sumif、verify、elseif 等内部的 if 不会被当成一次调用。
var excelIfCallPattern = regexp.MustCompile(`(?i)\bif\s*\(`)

// NormalizeExcelIFCalls 把 Excel 风格的三参调用 if(…) / If(…) 统一成 IF(…)，避免 expr 把小写 if 解析成关键字语法。
// 不会改写：sumif(、countif(、elseif(、以及 "if(" 若未来要做字符串字面量需另行处理。
func NormalizeExcelIFCalls(expr string) string {
	return excelIfCallPattern.ReplaceAllString(expr, "IF(")
}
