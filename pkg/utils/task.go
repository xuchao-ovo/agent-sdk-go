package utils

import (
	"errors"
	"sync"
)

// TaskIDPool 任务ID池
type TaskIDPool struct {
	segments    []*IDSegment // 存储 ID 段
	segmentSize int          // 每个段的大小
}

type IDSegment struct {
	mu       sync.Mutex // 每个段的锁
	nextID   int        // 当前段的下一个可用任务 ID
	maxID    int        // 当前段的最大任务 ID
	recycled chan int   // 回收的任务 ID
}

// NewTaskIDPool 创建任务ID池，分段管理任务ID
func NewTaskIDPool(segmentCount, segmentSize int) *TaskIDPool {
	pool := &TaskIDPool{
		segmentSize: segmentSize,
		segments:    make([]*IDSegment, segmentCount),
	}

	// 初始化各个段
	for i := 0; i < segmentCount; i++ {
		pool.segments[i] = &IDSegment{
			nextID:   1,
			maxID:    segmentSize,
			recycled: make(chan int, segmentSize),
		}
	}
	return pool
}

// GetTaskID 获取一个新的任务 ID
func (pool *TaskIDPool) GetTaskID() (int, error) {

	// 遍历所有段，尝试获取任务 ID
	for _, segment := range pool.segments {
		segment.mu.Lock()
		if segment.nextID <= segment.maxID {
			taskID := segment.nextID
			segment.nextID++
			segment.mu.Unlock()
			return taskID, nil
		}
		segment.mu.Unlock()
	}

	// 如果所有段都没有可用 ID，尝试从回收池获取 ID
	for _, segment := range pool.segments {
		select {
		case taskID := <-segment.recycled:
			return taskID, nil
		default:
		}
	}

	// 如果没有可用的 ID，返回错误
	return 0, errors.New("任务 ID 池已用完，且回收池为空")
}

// RecycleTaskID 回收一个已完成的任务 ID
func (pool *TaskIDPool) RecycleTaskID(taskID int) {
	// 将任务 ID 放回某个段的回收池
	for _, segment := range pool.segments {
		// 在锁外进行通道操作，防止死锁
		segment.mu.Lock()
		if taskID >= segment.nextID-1 && taskID <= segment.maxID {
			segment.mu.Unlock()
			// 使用非阻塞通道写入
			select {
			case segment.recycled <- taskID:
			default:
			}
			return
		}
		segment.mu.Unlock()
	}
}
