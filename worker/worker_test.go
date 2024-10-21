package worker

import (
	"runtime"
	"strconv"
	"testing"
	"time"
)

func BenchmarkWorkerPool(b *testing.B) {
	wp := NewWorkerPool(runtime.NumCPU(), 1000)
	wp.Start()

	// Reset timer before starting the actual benchmark measurement
	b.ResetTimer()

	// Submit tasks
	for i := 0; i < b.N; i++ {
		wp.Submit(Task{
			TraceID: "Task-" + strconv.Itoa(i),
			Action: func() error {
				time.Sleep(1 * time.Second)
				return nil
			},
		})
	}

	// Shutdown the worker pool
	wp.Stop()

	// Report allocations and stop the timer
	b.ReportAllocs()
	b.StopTimer()
}
