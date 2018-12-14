package module

import (
	"errors"
)

// ErrNotFoundModuleInstance 未找到组件实例的错误类型。
var ErrNotFoundModuleInstance = errors.New("not found module instance")
