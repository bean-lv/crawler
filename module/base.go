package module

import (
	"net/http"
)

// Counts 汇集组件内部计数的类型。
type Counts struct {
	// CalledCount 调用计数。
	CalledCount uint64
	// AcceptedCount 接受计数。
	AcceptedCount uint64
	// CompletedCount 成功完成计数。
	CompletedCount uint64
	// HandlingNumber 实时处理数。
	HandlingNumber uint64
}

// SummaryStruct 组件摘要结构类型。
type SummaryStruct struct {
	ID        MID         `json:"id"`
	Called    uint64      `json:"called"`
	Accepted  uint64      `json:"accepted"`
	Completed uint64      `json:"completed"`
	Handling  uint64      `json:"handling"`
	Extra     interface{} `json:"extra,omitempty"`
}

// Module 组件的基础接口类型。
// 该接口的实现类型必须是并发安全的！
type Module interface {
	// ID 获取当前组件的ID。
	ID() MID
	// Addr 获取当前组件的网络地址。
	Addr() string
	// Score 获取当前组件的评分。
	Score() uint64
	// SetScore 设置当前组件的评分。
	SetScore(score uint64)
	// ScoreCalculator 获取评分计数器。
	ScoreCalculator() CalculateScore
	// CalledCount 获取当前组件被调用的计数。
	CalledCount() uint64
	// AcceptedCount 获取当前组件接受调用的计数。
	AcceptedCount() uint64
	// CompletedCount 获取当前组件成功完成的计数。
	CompletedCount() uint64
	// HandlingNumber 获取当前组件正在处理的调用的数量。
	HandlingNumber() uint64
	// Counts 一次性获取所有计数。
	Counts() Counts
	// Summary 获取组件摘要。
	Summary() SummaryStruct
}

// Downloader 下载器的接口类型。
// 该接口的实现类型必须是并发安全的！
type Downloader interface {
	Module
	// Download 根据请求获取内容并返回响应。
	Download(req *Request) (*Response, error)
}

// Analyzer 分析器的接口类型。
// 该接口的实现类型必须是并发安全的！
type Analyzer interface {
	Module
	// RespParsers 返回当前分析器使用的响应解析函数列表。
	RespParsers() []ParseResponse
	// Analyze 根据规则分析响应并返回请求和条目。
	// 响应需要分别经过若干响应解析函数处理，然后合并结果。
	Analyze(resp *Response) ([]Data, []error)
}

// ParseResponse 用于解析HTTP响应的函数类型。
type ParseResponse func(httpResp *http.Response, respDepth uint32) ([]Data, []error)

// Pipeline 条目处理管道的接口类型。
// 该接口的实现类型必须是并发安全的！
type Pipeline interface {
	Module
	// ItemProcessors 返回当前条目处理管道使用的条目处理函数列表。
	ItemProcessors() []ProcessItem
	// Send 向条目处理管道发送条目
	// 条目需要依次经过若干条目处理函数的处理。
	Send(item Item) []error
	// FailFast 当前条目处理管道是否是快速失败的。
	// 这里快速失败指：只要在处理某个条目时在某一步骤上出错，
	// 那么条目处理管道就会忽略后续的所有处理步骤，并报告错误。
	FailFast() bool
	// SetFailFast 设置是否快速失败
	SetFailFast(failFast bool)
}

// ProcessItem 处理条目的函数类型
type ProcessItem func(item Item) (result Item, err error)
