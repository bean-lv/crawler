package scheduler

// Status 调度器状态的类型。
type Status uint8

const (
	// SCHED_STATUS_UNINITIALIZED 未初始化。
	SCHED_STATUS_UNINITIALIZED Status = 0
	// SCHED_STATUS_INITIALIZING 正在初始化。
	SCHED_STATUS_INITIALIZING Status = 1
	// SCHED_STATUS_INITIALIZED 已初始化。
	SCHED_STATUS_INITIALIZED Status = 2
	// SCHED_STATUS_STARTING 正在启动。
	SCHED_STATUS_STARTING Status = 3
	// SCHED_STATUS_STARTED 已启动。
	SCHED_STATUS_STARTED Status = 4
	// SCHED_STATUS_STOPPING 正在停止。
	SCHED_STATUS_STOPPING Status = 5
	// SCHED_STATUS_STOPPED 已停止。
	SCHED_STATUS_STOPPED Status = 6
)

// GetStatusDescription 获取状态的文字描述。
func GetStatusDescription(status Status) string {
	switch status {
	case SCHED_STATUS_UNINITIALIZED:
		return "uninitialized"
	case SCHED_STATUS_INITIALIZING:
		return "initializing"
	case SCHED_STATUS_INITIALIZED:
		return "initialized"
	case SCHED_STATUS_STARTING:
		return "starting"
	case SCHED_STATUS_STARTED:
		return "started"
	case SCHED_STATUS_STOPPING:
		return "stopping"
	case SCHED_STATUS_STOPPED:
		return "stopped"
	default:
		return "unknown"
	}
}
