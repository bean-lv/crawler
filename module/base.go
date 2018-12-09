package module

// Module 组件的基础接口类型
// 该接口的实现类型必须是并发安全的
type Module interface {
	// ID 获取当前组件的ID
	ID() MID
	// Addr 获取当前组件的网络地址
	Addr() string
	// Score 获取当前组件的评分
	Score() uint64
	// SetScore 设置当前组件的评分
	SetScore(score uint64)
	// ScoreCalculator 获取评分计数器
	ScoreCalculator() CalculateScore
	// CalledCount 获取当前组件被调用的计数
	CalledCount() uint64
	// AcceptedCount 获取当前组件接受调用的计数
	AcceptedCount() uint64
	// CompletedCount 获取当前组件成功完成的计数
	CompletedCount() uint64
	// HandlingNumber 获取当前组件正在处理的调用的数量
	HandlingNumber() uint64
	// Counts 一次性获取所有计数
	Counts() Counts
	// Summary 获取组件摘要
	Summary() SummaryStruct
}
