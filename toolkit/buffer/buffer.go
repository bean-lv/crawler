package buffer

// Buffer FIFO的缓冲器的接口类型。
type Buffer interface {
	// Cap 获取缓冲器的容量。
	Cap() uint32
	// Len 获取缓冲器中的数据数量。
	Len() uint32
	// Put 向缓冲器放入数据。
	// 注意！本方法应该是非阻塞的！
	// 若缓冲器已关闭，则直接返回非nil的错误值。
	Put(datum interface{}) (bool, error)
	// Get 从缓冲器获取数据。
	// 注意！本方法应该是非阻塞的！
	// 若缓冲器已关闭，则直接返回非nil的错误值。
	Get() (interface{}, error)
	// Close 关闭缓冲器。
	// 若缓冲器之前已关闭则返回false，否则返回true。
	Close() bool
	// Closed 判断缓冲器是否已关闭。
	Closed() bool
}
