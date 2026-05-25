package formula

// 文字类：字符串相关的独立公式函数（如 CONCAT、LEN、LEFT 等）在此定义接收者（如 TextFuncs）并注册到 build.go 的 exprOptions。
//
// 当前版本未单独暴露文本类 expr 函数。以下仍归属其它文件：
//   - SWITCH 的字符串分支匹配 → logic.go（LogicFuncs）
//   - IF/IFS 对「空串 / 非空串」的真假 → logic.go
//   - sum 将数字字符串按数值累加 → math.go（MathFuncs）
