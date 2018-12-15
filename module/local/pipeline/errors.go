package pipeline

import (
	"BeanGithub/crawler/errors"
)

// genError 生成爬虫错误值。
func genError(errMsg string) error {
	return errors.NewCrawlerError(errors.ERROR_TYPE_PIPELINE, errMsg)
}

// genParameterError 生成爬虫参数错误值。
func genParameterError(errMsg string) error {
	return errors.NewCrawlerErrorBy(errors.ERROR_TYPE_PIPELINE,
		errors.NewIllegalParameterError(errMsg))
}
