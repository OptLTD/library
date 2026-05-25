package source

/**
 * scene1: {表头按钮}[新增], {点击后}弹出[修改]{表单}(分组/双列), {确认后}调用新增接口，{成功后}隐藏弹窗
 * scene2: {行内按钮}[删除], {点击后}弹出[删除]{对话框}(你确认删除么？), {确认后}:调用删除接口, {成功后}隐藏弹窗
 * scene3: {单元格}[编号], {点击后}弹出[详情]{表单}(分组/双列), **{确认后}:什么都不做**, {成功后}隐藏弹窗
 * scene3: {单元行}[编号], {点击后}弹出[详情]{表单}(分组/双列), **{确认后}:什么都不做**, {成功后}隐藏弹窗
 * scene4: {表头按钮}[审批], {点击后}弹出[审批]{表单}(分组/双列), {确认后}:调用审批接口, {成功后}继续下一个审批
 * -----------------------------------------------------------------------------------------------
 * 抽象点击对象 click, 基础属性 => 显示内容{label}, 作用类型{ctype},作用事件{event}
 *    作用类型：表头按钮{button}|按钮分组{groups}｜行内按钮{action}｜单元格{cell-click}｜单元行{row-click}
 *    作用事件：单击{click},双击{dblclick},聚焦{focus},悬浮{hover}等
 * 响应事件后，触发一系列行为，这一些列行为都由前端代码来进行正常捕获再做更详细的处理，简单抽象响应行为{action=>click}
 *    响应行为: 确认框{confirm},输入框{prompt},信息框{notice},输入表单{form},什么都不做{nothing},自定义脚本{handle}
 * 在响应行为中有一部分行为并没有交互操作，但是有一部分是包含交互操作的，没有交互操作的直接到下一个步骤确认后{confirm},
 * 在很多交互体验良好的设计中确认后也会做相当多处理，比如锁定按钮，锁定界面等等，但此处做抽象处理
 *    确认行为：接口请求{request},什么都不做{nothing},自定义脚本{handle}
 * 确认行为处理结束后一半会对页面状态进行初始化，如关闭弹窗、刷新页面等，对这一步骤抽象作为成功后{success}
 *    成功确认：刷新页面{refresh},刷新应用{reload},关闭弹窗{hidden},继续处理{continue},自定义脚本{handle}
 *-------------------------------------------------------------------------------------------------
 * 根据以上分析，我们可以定从上述逻辑中组合出很大一部分想要的交互效果，但问题是这个编排交互逻辑的工作应该交给谁来处理，
 * 面向最终使用者看起来并不是一个很友好的事情，而如果面向开发人士或者研发周边的相关角色看起来是个比较不错的选择，只是
 * 我们要那么多选项的编排么，编排界面怎么做，编排的数据如何存储都是个问题！
 * 那么还是先回到问题的本质吧，用简单的处理来完成我们想要的工作！
 *-------------------------------------------------------------------------------------------------
 * 保留 ctype: button|action|cell-click|row-click
 * 忽略 event:
 * 保留 click: confirm|prompt|notice|form|none|handle
 * 拆出 submit: request|nothing|handle
 * 拆出 success: continue|hidden|reload|refresh
 */

type Click struct {
	UUKey  string `json:"uukey"`  // 数据类型
	Label  string `json:"label"`  // 显示文本
	Action string `json:"action"` // 按钮动作

	SeqNo uint64   `json:"seqno,omitempty"` // 排序序号
	Group string   `json:"group,omitempty"` // 按钮分组
	Scene []string `json:"scene,omitempty"` // 应用场景

	Helps string `json:"helps,omitempty"` // 提示信息
	Click string `json:"click,omitempty"` // 字段名称
	CType string `json:"ctype,omitempty"` // 事件类型
	Props object `json:"props,omitempty"` // 按钮属性
	Extra object `json:"extra,omitempty"` // 补充信息
}
