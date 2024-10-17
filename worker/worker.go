package worker

import (
	"sync"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/threading"
)

type Task struct {
	TraceID string       // Identifier for tracing task execution
	Action  func() error // The actual function to execute
}

// WorkerPool manages a pool of workers to process tasks.
type WorkerPool struct {
	tasks       chan Task // Channel to send tasks to workers
	rg          *threading.RoutineGroup
	numWorkers  int
	shutdownCh  chan struct{}
	mu          *sync.Mutex // pointer to the mutex
	cond        *sync.Cond  // shared data from mutex to synchronize
	isShutdown  bool
	isQueueFull bool
}

// NewWorkerPool creates a new worker pool.
func NewWorkerPool(numWorkers int, bufferSize int) *WorkerPool {
	logx.Info("[WORKER] Starting...")
	mu := new(sync.Mutex)
	wp := WorkerPool{
		tasks:       make(chan Task, bufferSize),
		rg:          threading.NewRoutineGroup(),
		numWorkers:  numWorkers,
		shutdownCh:  make(chan struct{}),
		isShutdown:  false,
		isQueueFull: false,
		cond:        sync.NewCond(mu),
		mu:          mu,
	}
	return &wp
}

// Start initializes the worker pool and starts the workers.
func (wp *WorkerPool) Start() {
	for i := 0; i < wp.numWorkers; i++ {
		// when run anything, it' always in waitGroup
		wp.rg.RunSafe(func() {
			for {
				select {
				case task, ok := <-wp.tasks: // Check for tasks
					if ok {
						wp.rg.RunSafe(func() {
							// Execute the task's action
							err := task.Action()
							if err != nil {
								logx.Errorf("[WORKER] Error task for reason %s with TraceID: %s", err.Error(), task.TraceID)
							} else {
								logx.Infof("[WORKER] Completed task with TraceID: %s", task.TraceID)
							}
						})
					}
				case <-wp.shutdownCh: // Check for shutdown signal
					logx.Infof("[WORKER] Worker received shutdown signal.")
					return // Exit the worker
				}
			}
		})
	}
}

// Submit adds a new task to the worker pool.
func (wp *WorkerPool) Submit(task Task) {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	// Do not add tasks if the pool is shutting down
	if wp.isShutdown {
		return
	}

	// Wait until there's space in the task queue (blocking behavior)
	if wp.isQueueFull {
		logx.Info("[WORKER] Queue is full. Task will be processed as space becomes available.")
		wp.cond.Wait() // Wait until signaled that space is available
	}

	wp.tasks <- task

	// Signal that there is a task available for processing
	if len(wp.tasks) == cap(wp.tasks) {
		wp.isQueueFull = true // Mark queue as full
	} else {
		wp.isQueueFull = false // Mark queue as not full
		wp.cond.Signal()       // Signal that space is available
	}

	logx.Infof("[WORKER] Task submitted successfully with TraceID: %v", task.TraceID)
}

// Shutdown waits for all workers to finish processing and stops the pool
func (wp *WorkerPool) Stop() {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if wp.isShutdown {
		return
	}
	logx.Info("[WORKER] Received termination signal, shutting down...")
	wp.isShutdown = true

	close(wp.tasks) // Signal  to stop
	close(wp.shutdownCh)
	wp.rg.Wait() // Wait for all workers to finish
	logx.Info("[WORKER] All workers have finished. Worker pool shutdown complete.")
}
