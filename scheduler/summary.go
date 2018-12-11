package scheduler

import (
	"BeanGithub/crawler/module"
)

// SchedSummary 调度器摘要的接口类型。
type SchedSummary interface {
	// Struct 获得摘要信息的结构化形式。
	Struct() SummaryStruct
	// String 获得摘要信息的字符串形式。
	String() string
}

// SummaryStruct 调度器摘要的结构。
type SummaryStruct struct {
	RequestArgs     RequestArgs             `json:"request_args"`
	DataArgs        DataArgs                `json:"data_args"`
	ModuleArgs      ModuleArgs              `json:"module_args"`
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

// BufferPoolSummaryStruct 缓冲池的摘要类型。
type BufferPoolSummaryStruct struct{}
