package scheduler

import (
	"net/http"
)

// Scheduler 调度器的接口类型。
type Scheduler interface {
	// Init 初始化调度器。
	// 参数requestArgs 请求相关的参数。
	// 参数dataArgs 数据相关的参数。
	// 参数moduleArgs 组件相关的参数。
	Init(requestArgs RequestArgs,
		dataArgs DataArgs,
		moduleArgs ModuleArgs) error
	// Start 启动调度器并执行爬取流程。
	// 参数firstHTTPReq 首次请求。调度器以此为起点开始执行爬取流程。
	Start(firstHTTPReq *http.Request) error
	// Stop 停止调度器的运行。
	// 所有处理模块执行的流程都会被终止。
	Stop() error
	// Status 获取调度器的状态。
	Status() Status
	// ErrorChan 获得错误通道。
	// 调度器以及各个处理模块运行过程中出现的所有错误都会被发送到该通道。
	// 若结果值为nil，则说明错误通道不可用或调度器已被停止。
	ErrorChan() <-chan error
	// Idle 判断所有处理模块是否都处于空闲状态。
	Idle() bool
	// Summary 获取摘要实例。
	Summary() SchedSummary
}
