package scheduler

import (
	"BeanGithub/crawler/module"
)

// Args 参数容器的接口类型。
type Args interface {
	// Check 用于自检参数的有效性。
	// 若结果值为nil，则说明未发现问题，否则就意味着自检未通过。
	Check() error
}

// RequestArgs 请求相关的参数容器类型。
type RequestArgs struct {
	// AcceptedDomains 可以接受的URL主域名列表。
	// URL主域名不在列表中的请求都会被忽略。
	AcceptedDomains []string `json:"accepted_primary_domains"`
	// MaxDepth 需要被爬取的最大深度。
	// 实际深度大于此值的请求都会被忽略。
	MaxDepth uint32 `json:"max_depth"`
}

// DataArgs 数据相关的参数容器类型。
type DataArgs struct {
	// ReqBufferCap 请求缓冲器的容量。
	ReqBufferCap uint32 `json:"req_buffer_cap"`
	// ReqMaxBufferNumber 请求缓冲器的最大数量。
	ReqMaxBufferNumber uint32 `json:"req_max_buffer_number"`
	// RespBufferCap 响应缓冲器的容量。
	RespBufferCap uint32 `json:"resp_buffer_cap"`
	// RespMaxBufferNumber 响应缓冲器的最大数量。
	RespMaxBufferNumber uint32 `json:"resp_max_buffer_number"`
	// ItemBufferCap 条目缓冲器的容量。
	ItemBufferCap uint32 `json:"item_buffer_cap"`
	// ItemMaxBufferNumber 条目缓冲器的最大数量。
	ItemMaxBufferNumber uint32 `json:"item_max_buffer_number"`
	// ErrorBufferCap 错误缓冲器的容量。
	ErrorBufferCap uint32 `json:"error_buffer_cap"`
	// ErrorMaxBufferNumber 错误缓冲器的最大数量。
	ErrorMaxBufferNumber uint32 `json:"error_max_buffer_number"`
}

// ModuleArgs 组件相关的参数容器类型。
type ModuleArgs struct {
	// 下载器列表。
	Downloaders []module.Downloader
	// 分析器列表。
	Analyzers []module.Analyzer
	// 条目处理管道列表。
	Pipelines []module.Pipeline
}
