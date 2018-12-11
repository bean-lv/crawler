package module

// Registrar 组件注册器接口。
type Registrar interface {
	// Register 注册组件实例。
	Register(modle Module) (bool, error)
	// Unregister 注销组件实例。
	Unregister(mid MID) (bool, error)
	// Get 获取一个指定类型的组件实例。
	// 基于负载均衡策略返回实例。
	Get(moduleType Type) (Module, error)
	// GetAllByType 获取指定类型的所有组件实例。
	GetAllByType(moduleType Type) (map[MID]Module, error)
	// GetAll 获取所有组件实例。
	GetAll() map[MID]Module
	// Clear 清除所有的组件注册记录。
	Clear()
}
