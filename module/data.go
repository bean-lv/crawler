package module

import (
	"net/http"
)

// Data 数据接口类型
type Data interface {
	// Valid 判断数据有效性
	Valid() bool
}

// Request 数据请求类型
type Request struct {
	// httpReq http请求
	httpReq *http.Request
	// depth 请求深度
	depth uint32
}

// NewRequest 创建一个请求实例
func NewRequest(httpReq *http.Request, depth uint32) *Request {
	return &Request{httpReq: httpReq, depth: depth}
}

// HTTPReq 获取http请求
func (req *Request) HTTPReq() *http.Request {
	return req.httpReq
}

// Depth 获取请求深度
func (req *Request) Depth() uint32 {
	return req.depth
}

// Valid 判断请求是否有效
func (req *Request) Valid() bool {
	return req.httpReq != nil && req.httpReq.URL != nil
}

// Response 数据响应类型
type Response struct {
	// httpResp http响应
	httpResp *http.Response
	// depth 响应深度
	depth uint32
}

// NewResponse 创建一个响应实例
func NewResponse(httpResp *http.Response, depth uint32) *Response {
	return &Response{httpResp: httpResp, depth: depth}
}

// HTTPResp 获取http响应
func (resp *Response) HTTPResp() *http.Response {
	return resp.httpResp
}

// Depth 获取响应深度
func (resp *Response) Depth() uint32 {
	return resp.depth
}

// Valid 判断响应是否有效
func (resp *Response) Valid() bool {
	return resp.httpResp != nil && resp.httpResp.Body != nil
}

// Item 条目类型
type Item map[string]interface{}

// Valid 判断条目是否有效
func (item Item) Valid() bool {
	return item != nil
}
