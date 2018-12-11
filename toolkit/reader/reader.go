package reader

import (
	"io"
)

// MultipleReader 多重读取器的接口。
type MultipleReader interface {
	// Reader 获取一个可关闭读取器的实例。
	// 后者会持有本多重读取器中的数据。
	Reader() io.ReadCloser
}
