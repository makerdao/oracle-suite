package gofer

// WorkerPool structure that contain Woker Pool implementation
type WorkerPool struct {
	workerCount int
}

// NewQueryPool create new worker pool for queries
func NewQueryPool(workerCount int) *WorkerPool {
	return &WorkerPool{
		workerCount: workerCount,
	}
}

// Start start worker pool
func (wp *WorkerPool) Start() {
	for w := 1; w <= wp.workerCount; w++ {
		//go worker(w, jobs, results)
	}
}

// Stop stop worker in pool
func (wp *WorkerPool) Stop() error {
	return nil
}
