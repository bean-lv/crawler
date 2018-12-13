package reader

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
)

// MultipleReader 多重读取器的接口。
type MultipleReader interface {
	// Reader 获取一个可关闭读取器的实例。
	// 后者会持有本多重读取器中的数据。
	Reader() io.ReadCloser
}

// myMultipleReader 多重读取器的实现类型。
type myMultipleReader struct {
	data []byte
}

// NewMultipleReader 新建并返回一个多重读取器的实例。
func NewMultipleReader(reader io.Reader) (MultipleReader, error) {
	var data []byte
	var err error
	if reader != nil {
		data, err = ioutil.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("multiple reader: couldn't create a new one: %s", err)
		}
	} else {
		data = []byte{}
	}
	return &myMultipleReader{
		data: data,
	}, nil
}

func (mr *myMultipleReader) Reader() io.ReadCloser {
	return ioutil.NopCloser(bytes.NewReader(mr.data))
}
