package scheduler

import (
	"BeanGithub/crawler/module"
	"BeanGithub/crawler/toolkit/buffer"
	"context"
	"fmt"
	"net/http"
	"sync"
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

// NewScheduler 创建一个调度器实例。
// func NewScheduler() Scheduler {
// 	return &myScheduler{}
// }

// myScheduler 调度器的实现类型。
type myScheduler struct {
	// maxDepth 爬取的最大深度。首次请求的深度为0。
	maxDepth uint32
	// acceptedDomainMap 可以接受的URL的主域名的字典。
	acceptedDomainMap sync.Map
	// registrar 组件注册器。
	registrar module.Registrar
	// reqBufferPool 请求缓冲池。
	reqBufferPool buffer.Pool
	// respBufferPool 响应缓冲池。
	respBufferPool buffer.Pool
	// itemBufferPool 条目缓冲池。
	itemBufferPool buffer.Pool
	// errorBufferPool 错误缓冲池。
	errorBufferPool buffer.Pool
	// urlMap 已处理的URL的字典。
	urlMap sync.Map
	// ctx 上下文，用于感知调度器的停止。
	ctx context.Context
	// cancelFunc 取消函数，用于停止调度器。
	cancelFunc context.CancelFunc
	// status 状态。
	status Status
	// statusLock 专用于状态的读写锁。
	statusLock sync.RWMutex
	// summary 摘要信息。
	summary SchedSummary
}

func (sched *myScheduler) Init(
	requestArgs RequestArgs,
	dataArgs DataArgs,
	moduleArgs ModuleArgs) (err error) {

	// 检查状态。
	fmt.Println("Check status for initialization...")
	var oldStatus Status
	oldStatus, err = sched.checkAndSetStatus(SCHED_STATUS_INITIALIZING)
	if err != nil {
		return
	}
	defer func() {
		sched.statusLock.Lock()
		if err != nil {
			sched.status = oldStatus
		} else {
			sched.status = SCHED_STATUS_INITIALIZED
		}
		sched.statusLock.Unlock()
	}()

	// 检查参数。
	fmt.Println("Check request arguments...")
	if err = requestArgs.Check(); err != nil {
		return
	}
	fmt.Println("Request arguments are valid.")
	fmt.Println("Check data arguments...")
	if err = dataArgs.Check(); err != nil {
		return
	}
	fmt.Println("Data arguments are valid.")
	fmt.Println("Check module arguments...")
	if err = moduleArgs.Check(); err != nil {
		return
	}
	fmt.Println("Module arguments are valid.")

	// 初始化内部字段。
	fmt.Println("Initialize scheduler's fields...")
	if sched.registrar == nil {
		sched.registrar = module.NewRegistrar()
	} else {
		sched.registrar.Clear()
	}
	sched.maxDepth = requestArgs.MaxDepth
	fmt.Printf("-- Max depth: %d", sched.maxDepth)
	sched.acceptedDomainMap = sync.Map{}
	for _, domain := range requestArgs.AcceptedDomains {
		sched.acceptedDomainMap.Store(domain, struct{}{})
	}
	fmt.Printf("-- Accepted primary domains: %v",
		requestArgs.AcceptedDomains)
	sched.urlMap = sync.Map{}
	sched.initBufferPool(dataArgs)
	sched.resetContext()
	sched.summary = newSchedSummary(requestArgs, dataArgs, moduleArgs, sched)

	// 注册组件。
	fmt.Println("Register modules...")
	if err = sched.registerModules(moduleArgs); err != nil {
		return
	}
	fmt.Println("Scheduler has been initialized.")
	return
}

func (sched *myScheduler) Start(firstHTTPReq *http.Request) (err error) {
	return
}

func (sched *myScheduler) Stop(err error) {
	return
}

func (sched *myScheduler) Status() Status {
	return 0
}

func (sched *myScheduler) ErrorChan() <-chan error {
	return nil
}

func (sched *myScheduler) Idle() bool {
	return false
}

func (sched *myScheduler) Summary() SchedSummary {
	return nil
}

// checkAndSetStatus 用于状态的检查，并在条件满足时设置状态。
func (sched *myScheduler) checkAndSetStatus(
	wantedStatus Status) (oldStatus Status, err error) {
	sched.statusLock.Lock()
	defer sched.statusLock.Unlock()
	oldStatus = sched.status
	err = checkStatus(oldStatus, wantedStatus, nil)
	if err == nil {
		sched.status = wantedStatus
	}
	return
}

// initBufferPool 按照给定的参数初始化缓冲池。
// 如果某个缓冲池可用且未关闭，就先关闭该缓冲池。
func (sched *myScheduler) initBufferPool(dataArgs DataArgs) {
	// 初始化请求缓冲池。
	if sched.reqBufferPool != nil && !sched.reqBufferPool.Closed() {
		sched.reqBufferPool.Close()
	}
	sched.reqBufferPool, _ = buffer.NewPool(
		dataArgs.ReqBufferCap, dataArgs.ReqMaxBufferNumber)
	fmt.Printf("-- Request buffer pool: bufferCap: %d, maxBufferNumber: %d",
		sched.reqBufferPool.BufferCap(), sched.reqBufferPool.MaxBufferNumber())
	// 初始化响应缓冲池。
	if sched.respBufferPool != nil && !sched.respBufferPool.Closed() {
		sched.respBufferPool.Close()
	}
	sched.respBufferPool, _ = buffer.NewPool(
		dataArgs.RespBufferCap, dataArgs.RespMaxBufferNumber)
	fmt.Printf("-- Response buffer pool: bufferCap: %d, maxBufferNumber: %d",
		sched.respBufferPool.BufferCap(), sched.respBufferPool.MaxBufferNumber())
	// 初始化条目缓冲池。
	if sched.itemBufferPool != nil && !sched.itemBufferPool.Closed() {
		sched.itemBufferPool.Close()
	}
	sched.itemBufferPool, _ = buffer.NewPool(
		dataArgs.ItemBufferCap, dataArgs.ItemMaxBufferNumber)
	fmt.Printf("-- Item buffer pool: bufferCap: %d, maxBufferNumber: %d",
		sched.itemBufferPool.BufferCap(), sched.itemBufferPool.MaxBufferNumber())
	// 初始化错误缓冲池。
	if sched.errorBufferPool != nil && !sched.errorBufferPool.Closed() {
		sched.errorBufferPool.Close()
	}
	sched.errorBufferPool, _ = buffer.NewPool(
		dataArgs.ErrorBufferCap, dataArgs.ErrorMaxBufferNumber)
	fmt.Printf("-- Error buffer pool: bufferCap: %d, maxBufferNumber: %d",
		sched.errorBufferPool.BufferCap(), sched.errorBufferPool.MaxBufferNumber())
}

// resetContext 重置调度器的上下文。
func (sched *myScheduler) resetContext() {
	sched.ctx, sched.cancelFunc = context.WithCancel(context.Background())
}

// registerModules 注册所有给定的组件。
func (sched *myScheduler) registerModules(moduleArgs ModuleArgs) error {
	for _, d := range moduleArgs.Downloaders {
		if d == nil {
			continue
		}
		ok, err := sched.registrar.Register(d)
		if err != nil {
			return genErrorByError(err)
		}
		if !ok {
			errMsg := fmt.Sprintf("Couldn't register downloader instance with MID %q!", d.ID())
			return genError(errMsg)
		}
		fmt.Printf("All downloaders have been registered. (number: %d)",
			len(moduleArgs.Downloaders))
		for _, a := range moduleArgs.Analyzers {
			if a == nil {
				continue
			}
			ok, err := sched.registrar.Register(a)
			if err != nil {
				return genErrorByError(err)
			}
			if !ok {
				errMsg := fmt.Sprintf("Couldn't register analyzer instance with MID %q!", a.ID())
				return genError(errMsg)
			}
		}
		fmt.Printf("All analyzers have been registered. (number: %d)",
			len(moduleArgs.Analyzers))
		for _, p := range moduleArgs.Pipelines {
			if p == nil {
				continue
			}
			ok, err := sched.registrar.Register(p)
			if err != nil {
				return genErrorByError(err)
			}
			if !ok {
				errMsg := fmt.Sprintf("Couldn't register pipeline instance with MID %q!", p.ID())
				return genError(errMsg)
			}
		}
		fmt.Printf("All pipelines have been registered. (number: %d)",
			len(moduleArgs.Pipelines))
		return nil
	}
	return nil
}
