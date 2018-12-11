package buffer

// Pool 数据缓冲池的接口类型。
type Pool interface {
	// BufferCap 获取池中缓冲器的容量。
	BufferCap() uint32
	// MaxBufferNumber 获取池中缓冲器的最大数量。
	MaxBufferNumber() uint32
	// BufferNumber 获取池中缓冲器的数量。
	BufferNumber() uint32
	// Total 获取缓冲池中数据的总数。
	Total() uint64
	// Put 向缓冲池放入数据。
	// 注意！本方法应该是阻塞的！
	// 若缓冲池已关闭，则直接返回非nil的错误值。
	Put(datum interface{}) error
	// Get 从缓冲池获取数据。
	// 注意！本方法应该是阻塞的！
	// 若缓冲池已关闭，则直接返回非nil的错误值。
	Get() (datum interface{}, err error)
	// Close 关闭缓冲池。
	// 若缓冲池之前已关闭则返回false，否则返回true。
	Close() bool
	// Closed 判断缓冲池是否已关闭。
	Closed() bool
}
