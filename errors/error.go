package errors

import (
	"fmt"
	"strings"
)

// ErrorType 错误类型
type ErrorType string

const (
	// ERROR_TYPE_DOWNLOADER 下载器错误
	ERROR_TYPE_DOWNLOADER ErrorType = "downloader error"
	// ERROR_TYPE_ANALYZER 分析器错误
	ERROR_TYPE_ANALYZER ErrorType = "analyzer error"
	// ERROR_TYPE_PIPELINE 条目处理管道错误
	ERROR_TYPE_PIPELINE ErrorType = "pipeline error"
	// ERROR_TYPE_SCHEDULER 调度器错误
	ERROR_TYPE_SCHEDULER ErrorType = "scheduler error"
)

// CrawlerError 爬虫错误接口类型
type CrawlerError interface {
	// Type 获取错误类型
	Type() ErrorType
	// Error 获取错误提示信息
	Error() string
}

// myCrawlerError 爬虫错误类型实现
type myCrawlerError struct {
	// errType 错误类型
	errType ErrorType
	// errMsg 错误提示信息
	errMsg string
	// fullErrMsg 完整的错误提示信息
	fullErrMsg string
}

// NewCrawlerError 创建新的爬虫错误
func NewCrawlerError(errType ErrorType, errMsg string) CrawlerError {
	return &myCrawlerError{
		errType: errType,
		errMsg:  errMsg,
	}
}

// NewCrawlerErrorBy 根据给定的错误创建新的爬虫错误
func NewCrawlerErrorBy(errType ErrorType, err error) CrawlerError {
	return NewCrawlerError(errType, err.Error())
}

// Type 获取错误类型
func (ce *myCrawlerError) Type() ErrorType {
	return ce.errType
}

// Error 获取错误提示信息
func (ce *myCrawlerError) Error() string {
	if ce.fullErrMsg == "" {
		ce.genFullErrMsg()
	}
	return ce.fullErrMsg
}

// getFullErrMsg 生成错误信息，并赋值给相应字段
func (ce *myCrawlerError) genFullErrMsg() {
	var builder strings.Builder
	builder.WriteString("crawler error: ")
	if ce.errMsg != "" {
		builder.WriteString(string(ce.errType))
		builder.WriteString(": ")
	}
	builder.WriteString(ce.errMsg)
	ce.fullErrMsg = fmt.Sprintf("%s", builder.String())
	return
}

// IllegalParameterError 非法的参数错误类型
type IllegalParameterError struct {
	msg string
}

// NewIllegalParameterError 新建一个非法参数错误实例
func NewIllegalParameterError(errMsg string) IllegalParameterError {
	return IllegalParameterError{
		msg: fmt.Sprintf("illegal parameter: %s", strings.TrimSpace(errMsg)),
	}
}

func (ipe IllegalParameterError) Error() string {
	return ipe.msg
}
