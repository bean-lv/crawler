package scheduler

import (
	"BeanGithub/crawler/module"
	"BeanGithub/crawler/toolkit/buffer"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
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
func NewScheduler() Scheduler {
	return &myScheduler{}
}

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
	defer func() {
		if p := recover(); p != nil {
			errMsg := fmt.Sprintf("Fatal scheduler error: %s", p)
			fmt.Println(errMsg)
			err = genError(errMsg)
		}
	}()
	fmt.Println("Start scheduler...")
	// 检查状态。
	fmt.Println("Check status for start...")
	var oldStatus Status
	oldStatus, err = sched.checkAndSetStatus(SCHED_STATUS_STARTING)
	defer func() {
		sched.statusLock.Lock()
		if err != nil {
			sched.status = oldStatus
		} else {
			sched.status = SCHED_STATUS_STARTED
		}
		sched.statusLock.Unlock()
	}()
	if err != nil {
		return
	}
	// 检查参数。
	fmt.Println("Check first HTTP request...")
	if firstHTTPReq == nil {
		err = genParameterError("nil first HTTP request")
		return
	}
	fmt.Println("The first HTTP request is valid.")
	// 获得首次请求的主域名，并将其添加到可接受的主域名的字典。
	fmt.Println("Get the primary domain...")
	fmt.Printf("-- Host: %s", firstHTTPReq.Host)
	var primaryDomain string
	primaryDomain, err = getPrimaryDomain(firstHTTPReq.Host)
	if err != nil {
		return
	}
	fmt.Printf("-- Primary domain: %s", primaryDomain)
	sched.acceptedDomainMap.Store(primaryDomain, struct{}{})
	// 开始调度数据和组件。
	if err = sched.checkBufferPoolForStart(); err != nil {
		return
	}
	sched.download()
	sched.analyze()
	sched.pick()
	fmt.Println("Scheduler has been started.")
	// 放入第一个请求。
	firstReq := module.NewRequest(firstHTTPReq, 0)
	sched.sendReq(firstReq)

	return
}

func (sched *myScheduler) Stop() (err error) {
	fmt.Println("Stop scheduler...")
	// 检查状态。
	fmt.Println("Check status for stop...")
	var oldStatus Status
	oldStatus, err = sched.checkAndSetStatus(SCHED_STATUS_STOPPING)
	defer func() {
		sched.statusLock.Lock()
		if err != nil {
			sched.status = oldStatus
		} else {
			sched.status = SCHED_STATUS_STOPPED
		}
		sched.statusLock.Unlock()
	}()
	if err != nil {
		return
	}
	sched.cancelFunc()
	sched.reqBufferPool.Close()
	sched.respBufferPool.Close()
	sched.itemBufferPool.Close()
	sched.errorBufferPool.Close()
	fmt.Println("Scheduler has been stopped.")
	return
}

func (sched *myScheduler) Status() Status {
	var status Status
	sched.statusLock.RLock()
	status = sched.status
	sched.statusLock.RUnlock()
	return status
}

func (sched *myScheduler) ErrorChan() <-chan error {
	errBuffer := sched.errorBufferPool
	errCh := make(chan error, errBuffer.BufferCap())
	go func(errBuffer buffer.Pool, errCh chan error) {
		for {
			if sched.canceled() {
				close(errCh)
				break
			}
			datum, err := errBuffer.Get()
			if err != nil {
				fmt.Println("The error buffer pool was closed. Break error reception.")
				close(errCh)
				break
			}
			err, ok := datum.(error)
			if !ok {
				errMsg := fmt.Sprintf("incorrect error type: %T", datum)
				sendError(errors.New(errMsg), "", sched.errorBufferPool)
				continue
			}
			if sched.canceled() {
				close(errCh)
				break
			}
			errCh <- err
		}
	}(errBuffer, errCh)
	return errCh
}

func (sched *myScheduler) Idle() bool {
	moduleMap := sched.registrar.GetAll()
	for _, module := range moduleMap {
		if module.HandlingNumber() > 0 {
			return false
		}
	}
	if sched.reqBufferPool.Total() > 0 ||
		sched.respBufferPool.Total() > 0 ||
		sched.itemBufferPool.Total() > 0 {
		return false
	}
	return true
}

func (sched *myScheduler) Summary() SchedSummary {
	return sched.summary
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

// checkBufferPoolForStart 检查缓冲池是否已为调度器的启动准备就绪。
// 如果某个缓冲池不可用，就直接返回错误值报告此情况。
// 如果某个缓冲池已关闭，就按照原先的参数重新初始化它。
func (sched *myScheduler) checkBufferPoolForStart() error {
	// 检查请求缓冲池。
	if sched.reqBufferPool == nil {
		return genError("nil request buffer pool")
	}
	if sched.reqBufferPool != nil && sched.reqBufferPool.Closed() {
		sched.reqBufferPool, _ = buffer.NewPool(
			sched.reqBufferPool.BufferCap(), sched.reqBufferPool.MaxBufferNumber())
	}
	// 检查响应缓冲池。
	if sched.respBufferPool == nil {
		return genError("nil response buffer pool")
	}
	if sched.respBufferPool != nil && sched.respBufferPool.Closed() {
		sched.respBufferPool, _ = buffer.NewPool(
			sched.respBufferPool.BufferCap(), sched.respBufferPool.MaxBufferNumber())
	}
	// 检查条目缓冲池。
	if sched.itemBufferPool == nil {
		return genError("nil item buffer pool")
	}
	if sched.itemBufferPool != nil && sched.itemBufferPool.Closed() {
		sched.itemBufferPool, _ = buffer.NewPool(
			sched.itemBufferPool.BufferCap(), sched.itemBufferPool.MaxBufferNumber())
	}
	// 检查错误缓冲池。
	if sched.errorBufferPool == nil {
		return genError("nil error buffer pool")
	}
	if sched.errorBufferPool != nil && sched.errorBufferPool.Closed() {
		sched.errorBufferPool, _ = buffer.NewPool(
			sched.errorBufferPool.BufferCap(), sched.errorBufferPool.MaxBufferNumber())
	}
	return nil
}

// download 从请求缓冲池取出请求并下载，
// 然后把得到的响应放入响应缓冲池。
func (sched *myScheduler) download() {
	go func() {
		for {
			if sched.canceled() {
				break
			}
			datum, err := sched.reqBufferPool.Get()
			if err != nil {
				fmt.Println("The request buffer pool was closed. Break request reception.")
				break
			}
			req, ok := datum.(*module.Request)
			if !ok {
				errMsg := fmt.Sprintf("incorrect request type: %T", datum)
				sendError(errors.New(errMsg), "", sched.errorBufferPool)
			}
			sched.downloadOne(req)
		}
	}()
}

// downloadOne 根据指定的请求执行下载，并把响应放入响应缓冲池。
func (sched *myScheduler) downloadOne(req *module.Request) {
	if req == nil {
		return
	}
	if sched.canceled() {
		return
	}
	m, err := sched.registrar.Get(module.TYPE_DOWNLOADER)
	if err != nil || m == nil {
		errMsg := fmt.Sprintf("couldn't get a downloader: %s", err)
		sendError(errors.New(errMsg), "", sched.errorBufferPool)
		sched.sendReq(req)
		return
	}
	downloader, ok := m.(module.Downloader)
	if !ok {
		errMsg := fmt.Sprintf("incorrect downloader type: %T (MID: %s)",
			m, m.ID())
		sendError(errors.New(errMsg), m.ID(), sched.errorBufferPool)
		sched.sendReq(req)
		return
	}
	resp, err := downloader.Download(req)
	if resp != nil {
		sendResp(resp, sched.respBufferPool)
	}
	if err != nil {
		sendError(err, m.ID(), sched.errorBufferPool)
	}
}

// analyze 从响应缓冲池取出响应并解析，
// 然后把得到的条目或请求放入相应的缓冲池。
func (sched *myScheduler) analyze() {
	go func() {
		for {
			if sched.canceled() {
				break
			}
			datum, err := sched.respBufferPool.Get()
			if err != nil {
				fmt.Println("The response buffer pool was closed. Break response reception.")
				break
			}
			resp, ok := datum.(*module.Response)
			if !ok {
				errMsg := fmt.Sprintf("incorrect response type: %T", datum)
				sendError(errors.New(errMsg), "", sched.errorBufferPool)
			}
			sched.analyzeOne(resp)
		}
	}()
}

// analyzeOne 根据指定的响应执行解析并把结果放入相应的缓冲池。
func (sched *myScheduler) analyzeOne(resp *module.Response) {
	if resp == nil {
		return
	}
	if sched.canceled() {
		return
	}
	m, err := sched.registrar.Get(module.TYPE_ANALYZER)
	if err != nil || m == nil {
		errMsg := fmt.Sprintf("couldn't get an analyzer: %s", err)
		sendError(errors.New(errMsg), "", sched.errorBufferPool)
		sendResp(resp, sched.respBufferPool)
		return
	}
	analyzer, ok := m.(module.Analyzer)
	if !ok {
		errMsg := fmt.Sprintf("incorrect analyzer type: %T (MID: %s)",
			m, m.ID())
		sendError(errors.New(errMsg), m.ID(), sched.errorBufferPool)
		sendResp(resp, sched.respBufferPool)
		return
	}
	dataList, errs := analyzer.Analyze(resp)
	if dataList != nil {
		for _, data := range dataList {
			if data == nil {
				continue
			}
			switch d := data.(type) {
			case *module.Request:
				sched.sendReq(d)
			case module.Item:
				sendItem(d, sched.itemBufferPool)
			default:
				errMsg := fmt.Sprintf("Unsupported data type %T! (data: %#v)", d, d)
				sendError(errors.New(errMsg), m.ID(), sched.errorBufferPool)
			}
		}
	}
	if errs != nil {
		for _, err := range errs {
			sendError(err, m.ID(), sched.errorBufferPool)
		}
	}
}

// pick 从条目缓冲池取出条目并处理。
func (sched *myScheduler) pick() {
	go func() {
		for {
			if sched.canceled() {
				break
			}
			datum, err := sched.itemBufferPool.Get()
			if err != nil {
				fmt.Println("The item buffer pool was closed. Break item reception.")
				break
			}
			item, ok := datum.(module.Item)
			if !ok {
				errMsg := fmt.Sprintf("incorrect item type: %T", datum)
				sendError(errors.New(errMsg), "", sched.errorBufferPool)
			}
			sched.pickOne(item)
		}
	}()
}

// piclOne 处理给定的条目。
func (sched *myScheduler) pickOne(item module.Item) {
	if sched.canceled() {
		return
	}
	m, err := sched.registrar.Get(module.TYPE_PIPELINE)
	if err != nil || m == nil {
		errMsg := fmt.Sprintf("couldn't get a pipeline: %s", err)
		sendError(errors.New(errMsg), "", sched.errorBufferPool)
		sendItem(item, sched.itemBufferPool)
		return
	}
	pipeline, ok := m.(module.Pipeline)
	if !ok {
		errMsg := fmt.Sprintf("incorrect pipeline type: %T (MID: %s)",
			m, m.ID())
		sendError(errors.New(errMsg), m.ID(), sched.errorBufferPool)
		sendItem(item, sched.itemBufferPool)
		return
	}
	errs := pipeline.Send(item)
	if errs != nil {
		for _, err := range errs {
			sendError(err, m.ID(), sched.errorBufferPool)
		}
	}
}

// sendReq 向请求缓冲池发送请求。
// 不符合要求的请求会被过滤掉。
func (sched *myScheduler) sendReq(req *module.Request) bool {
	if req == nil {
		return false
	}
	if sched.canceled() {
		return false
	}
	httpReq := req.HTTPReq()
	if httpReq == nil {
		fmt.Println("Ignore the request! Its HTTP request is invalid!")
		return false
	}
	reqURL := httpReq.URL
	if reqURL == nil {
		fmt.Println("Ignore the request! Ites URL is invalid!")
		return false
	}
	scheme := strings.ToLower(reqURL.Scheme)
	if scheme != "http" && scheme != "https" {
		fmt.Printf("Ignore the request! Its URL scheme is %q, but should be %q or %q. (URL: %s)\n",
			scheme, "http", "https", reqURL)
		return false
	}
	if v, _ := sched.urlMap.Load(reqURL.String()); v != nil {
		fmt.Printf("Ignore the request! Its URL is repeated. (URL: %s)\n", reqURL)
		return false
	}
	pd, _ := getPrimaryDomain(httpReq.Host)
	if v, _ := sched.acceptedDomainMap.Load(pd); v == nil {
		if pd == "bing.net" {
			panic(httpReq.URL)
		}
		fmt.Printf("Ignore the request! Its host %q is not in accepted primary domain map. (URL: %s)\n",
			httpReq.Host, reqURL)
		return false
	}
	if req.Depth() > sched.maxDepth {
		fmt.Printf("Ignore the request! Its depth %d is greater than %d. (URL: %s)\n",
			req.Depth(), sched.maxDepth, reqURL)
		return false
	}
	go func(req *module.Request) {
		if err := sched.reqBufferPool.Put(req); err != nil {
			fmt.Println("The request buffer pool was closed. Ignore request sending.")
		}
	}(req)
	sched.urlMap.Store(reqURL.String(), struct{}{})
	return true
}

// sendResp 向响应缓冲池发送响应。
func sendResp(resp *module.Response, respBufferPool buffer.Pool) bool {
	if resp == nil || respBufferPool == nil || respBufferPool.Closed() {
		return false
	}
	go func(resp *module.Response) {
		if err := respBufferPool.Put(resp); err != nil {
			fmt.Println("The response buffer pool was closed. Ignore response sending.")
		}
	}(resp)
	return true
}

// sendItem 向条目缓冲池发送条目。
func sendItem(item module.Item, itemBufferPool buffer.Pool) bool {
	if item == nil || itemBufferPool == nil || itemBufferPool.Closed() {
		return false
	}
	go func(item module.Item) {
		if err := itemBufferPool.Put(item); err != nil {
			fmt.Println("The item buffer pool was closed. Ignore item sending.")
		}
	}(item)
	return true
}

// canceled 判断调度器的上下文是否已被取消。
func (sched *myScheduler) canceled() bool {
	select {
	case <-sched.ctx.Done():
		return true
	default:
		return false
	}
}
