package scheduler

import (
	"BeanGithub/crawler/module"
	"BeanGithub/crawler/toolkit/buffer"
	"encoding/json"
	"fmt"
	"sort"
)

// SchedSummary 调度器摘要的接口类型。
type SchedSummary interface {
	// Struct 获得摘要信息的结构化形式。
	Struct() SummaryStruct
	// String 获得摘要信息的字符串形式。
	String() string
}

// newSchedSummary 创建一个调度器摘要的实例。
func newSchedSummary(
	requestArgs RequestArgs,
	dataArgs DataArgs,
	moduleArgs ModuleArgs,
	sched *myScheduler) SchedSummary {
	if sched == nil {
		return nil
	}
	return &mySchedSummary{
		requestArgs: requestArgs,
		dataArgs:    dataArgs,
		moduleArgs:  moduleArgs,
		sched:       sched,
	}
}

// mySchedSummary 调度器摘要的实现类型。
type mySchedSummary struct {
	// requestArgs 请求相关的参数。
	requestArgs RequestArgs
	// dataArgs 数据相关的参数。
	dataArgs DataArgs
	// moduleArgs 组件相关的参数。
	moduleArgs ModuleArgs
	// maxDepth 爬取的最大深度。
	maxDepth uint32
	// sched 调度器实例。
	sched *myScheduler
}

// SummaryStruct 调度器摘要的结构。
type SummaryStruct struct {
	RequestArgs     RequestArgs             `json:"request_args"`
	DataArgs        DataArgs                `json:"data_args"`
	ModuleArgs      ModuleArgsSummary       `json:"module_args"`
	Status          string                  `json:"status"`
	Downloaders     []module.SummaryStruct  `json:"downloaders"`
	Analyzers       []module.SummaryStruct  `json:"analyzers"`
	Pipelines       []module.SummaryStruct  `json:"pipelines"`
	ReqBufferPool   BufferPoolSummaryStruct `json:"request_buffer_pool"`
	RespBufferPool  BufferPoolSummaryStruct `json:"response_buffer_pool"`
	ItemBufferPool  BufferPoolSummaryStruct `json:"item_buffer_pool"`
	ErrorBufferPool BufferPoolSummaryStruct `json:"error_buffer_pool"`
	NumberURL       uint64                  `json:"url_number"`
}

func (ss *mySchedSummary) Struct() SummaryStruct {
	registrar := ss.sched.registrar
	return SummaryStruct{
		RequestArgs:     ss.requestArgs,
		DataArgs:        ss.dataArgs,
		ModuleArgs:      ss.moduleArgs.Summary(),
		Status:          GetStatusDescription(ss.sched.Status()),
		Downloaders:     getModuleSummaries(registrar, module.TYPE_DOWNLOADER),
		Analyzers:       getModuleSummaries(registrar, module.TYPE_ANALYZER),
		Pipelines:       getModuleSummaries(registrar, module.TYPE_PIPELINE),
		ReqBufferPool:   getBufferPoolSummary(ss.sched.reqBufferPool),
		RespBufferPool:  getBufferPoolSummary(ss.sched.respBufferPool),
		ItemBufferPool:  getBufferPoolSummary(ss.sched.itemBufferPool),
		ErrorBufferPool: getBufferPoolSummary(ss.sched.errorBufferPool),
		// NumberURL:       len(ss.sched.urlMap),
	}
}

func (ss *mySchedSummary) String() string {
	b, err := json.MarshalIndent(ss.Struct(), "", "    ")
	if err != nil {
		fmt.Printf("An error occurs when generating scheduler summary: %s\n", err)
		return ""
	}
	return string(b)
}

// BufferPoolSummaryStruct 缓冲池的摘要类型。
type BufferPoolSummaryStruct struct {
	BufferCap       uint32 `json:"buffer_cap"`
	MaxBufferNumber uint32 `json:"max_buffer_number"`
	BufferNumber    uint32 `json:"buffer_number"`
	Total           uint64 `json:"total"`
}

// getBufferPoolSummary 生成和返回某个数据缓冲池的摘要信息。
func getBufferPoolSummary(bufferPool buffer.Pool) BufferPoolSummaryStruct {
	return BufferPoolSummaryStruct{
		BufferCap:       bufferPool.BufferCap(),
		MaxBufferNumber: bufferPool.MaxBufferNumber(),
		BufferNumber:    bufferPool.BufferNumber(),
		Total:           bufferPool.Total(),
	}
}

// getModuleSummaries 获取已注册的某类组件的摘要。
func getModuleSummaries(registrar module.Registrar, mType module.Type) []module.SummaryStruct {
	moduleMap, _ := registrar.GetAllByType(mType)
	summaries := []module.SummaryStruct{}
	if len(moduleMap) > 0 {
		for _, module := range moduleMap {
			summaries = append(summaries, module.Summary())
		}
	}
	if len(summaries) > 0 {
		sort.Slice(summaries,
			func(i, j int) bool {
				return summaries[i].ID < summaries[j].ID
			})
	}
	return summaries
}
