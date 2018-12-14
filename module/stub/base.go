package stub

import (
	"BeanGithub/crawler/module"
)

// ModuleInternal 组件的内部基础接口类型。
type ModuleInternal interface {
	module.Module
	// IncrCalledCount 调用计数增加1。
	IncrCalledCount()
	// IncrAcceptedCount 接受计数增加1。
	IncrAcceptedCount()
	// IncrCompletedCount 成功完成计数增加1。
	IncrCompletedCount()
	// IncrHandlingNumber 实时处理数增加1。
	IncrHandlingNumber()
	// DecrHandlingNumber 实时处理数减1。
	DecrHandlingNumber()
	// Clear 清空所有计数。
	Clear()
}
