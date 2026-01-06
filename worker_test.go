package goroutine

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestNewWorkerPool tests worker pool creation
func TestNewWorkerPool(t *testing.T) {
	wp := NewWorkerPool(5)
	if wp == nil {
		t.Fatal("Expected worker pool, got nil")
	}
	
	if wp.WorkerCount() != 5 {
		t.Errorf("Expected 5 workers, got %d", wp.WorkerCount())
	}
	
	if wp.IsRunning() {
		t.Error("Worker pool should not be running initially")
	}
}

// TestWorkerPoolStartStop tests starting and stopping the worker pool
func TestWorkerPoolStartStop(t *testing.T) {
	wp := NewWorkerPool(3)
	
	if wp.IsRunning() {
		t.Error("Worker pool should not be running initially")
	}
	
	wp.Start()
	
	if !wp.IsRunning() {
		t.Error("Worker pool should be running after Start()")
	}
	
	// Give it a moment to start
	time.Sleep(10 * time.Millisecond)
	
	wp.Stop()
	
	if wp.IsRunning() {
		t.Error("Worker pool should not be running after Stop()")
	}
}

// TestWorkerPoolSubmit tests immediate task submission
func TestWorkerPoolSubmit(t *testing.T) {
	wp := NewWorkerPool(2)
	wp.Start()
	defer wp.Stop()
	
	var counter int32
	var wg sync.WaitGroup
	
	numTasks := 10
	wg.Add(numTasks)
	
	for i := 0; i < numTasks; i++ {
		task := &WorkerTask{
			ID: fmt.Sprintf("task-%d", i),
			Func: func(ctx context.Context) error {
				atomic.AddInt32(&counter, 1)
				wg.Done()
				return nil
			},
		}
		
		if err := wp.Submit(task); err != nil {
			t.Fatalf("Failed to submit task: %v", err)
		}
	}
	
	// Wait for all tasks to complete
	wg.Wait()
	
	if atomic.LoadInt32(&counter) != int32(numTasks) {
		t.Errorf("Expected %d tasks executed, got %d", numTasks, counter)
	}
}

// TestWorkerPoolSubmitDelayed tests delayed task execution
func TestWorkerPoolSubmitDelayed(t *testing.T) {
	wp := NewWorkerPool(2)
	wp.Start()
	defer wp.Stop()
	
	var executed bool
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(1)
	
	task := &WorkerTask{
		ID: "delayed-task",
		Func: func(ctx context.Context) error {
			mu.Lock()
			executed = true
			mu.Unlock()
			wg.Done()
			return nil
		},
	}
	
	delay := 200 * time.Millisecond
	start := time.Now()
	
	if err := wp.SubmitDelayed(task, delay); err != nil {
		t.Fatalf("Failed to submit delayed task: %v", err)
	}
	
	// Task should not execute immediately
	time.Sleep(50 * time.Millisecond)
	mu.Lock()
	if executed {
		t.Error("Task executed too early")
	}
	mu.Unlock()
	
	// Wait for task to complete
	wg.Wait()
	elapsed := time.Since(start)
	
	mu.Lock()
	if !executed {
		t.Error("Task was not executed")
	}
	mu.Unlock()
	
	if elapsed < delay {
		t.Errorf("Task executed too early: %v < %v", elapsed, delay)
	}
}

// TestWorkerPoolSubmitCron tests cron job execution
func TestWorkerPoolSubmitCron(t *testing.T) {
	wp := NewWorkerPool(2)
	wp.Start()
	defer wp.Stop()
	
	var counter int32
	
	task := &WorkerTask{
		ID: "cron-task",
		Func: func(ctx context.Context) error {
			atomic.AddInt32(&counter, 1)
			return nil
		},
	}
	
	// Run every 100ms
	if err := wp.SubmitCron(task, "@every 100ms"); err != nil {
		t.Fatalf("Failed to submit cron task: %v", err)
	}
	
	// Wait for multiple executions
	time.Sleep(550 * time.Millisecond)
	
	// Cancel the cron task
	wp.CancelCron("cron-task")
	
	executions := atomic.LoadInt32(&counter)
	
	// Should have executed at least 3-4 times in 550ms
	if executions < 3 {
		t.Errorf("Expected at least 3 executions, got %d", executions)
	}
	
	// Wait a bit more to ensure it stopped
	prevCount := executions
	time.Sleep(200 * time.Millisecond)
	newCount := atomic.LoadInt32(&counter)
	
	if newCount > prevCount {
		t.Error("Task continued executing after cancellation")
	}
}

// TestCronSchedule tests cron schedule parsing
func TestCronSchedule(t *testing.T) {
	tests := []struct {
		expr     string
		wantErr  bool
		expected time.Duration
	}{
		{"@hourly", false, time.Hour},
		{"@daily", false, 24 * time.Hour},
		{"@weekly", false, 7 * 24 * time.Hour},
		{"@every 5s", false, 5 * time.Second},
		{"@every 1m", false, time.Minute},
		{"@every 1h", false, time.Hour},
		{"invalid", true, 0},
	}
	
	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			schedule, err := NewCronSchedule(tt.expr)
			
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			if schedule.interval != tt.expected {
				t.Errorf("Expected interval %v, got %v", tt.expected, schedule.interval)
			}
			
			// Test Next()
			now := time.Now()
			next := schedule.Next(now)
			expectedNext := now.Add(tt.expected)
			
			if next != expectedNext {
				t.Errorf("Expected next time %v, got %v", expectedNext, next)
			}
		})
	}
}

// TestWorkerPoolMultipleDelayed tests multiple delayed tasks
func TestWorkerPoolMultipleDelayed(t *testing.T) {
	wp := NewWorkerPool(3)
	wp.Start()
	defer wp.Stop()
	
	var wg sync.WaitGroup
	results := make(map[string]time.Time)
	var mu sync.Mutex
	
	tasks := []struct {
		id    string
		delay time.Duration
	}{
		{"task1", 100 * time.Millisecond},
		{"task2", 200 * time.Millisecond},
		{"task3", 50 * time.Millisecond},
	}
	
	start := time.Now()
	
	for _, tt := range tasks {
		wg.Add(1)
		taskID := tt.id // Capture for closure
		task := &WorkerTask{
			ID: taskID,
			Func: func(ctx context.Context) error {
				mu.Lock()
				results[taskID] = time.Now()
				mu.Unlock()
				wg.Done()
				return nil
			},
		}
		
		if err := wp.SubmitDelayed(task, tt.delay); err != nil {
			t.Fatalf("Failed to submit task %s: %v", tt.id, err)
		}
	}
	
	wg.Wait()
	
	// Verify execution order
	mu.Lock()
	defer mu.Unlock()
	
	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}
	
	// task3 should execute first (50ms), then task1 (100ms), then task2 (200ms)
	elapsed3 := results["task3"].Sub(start)
	elapsed1 := results["task1"].Sub(start)
	elapsed2 := results["task2"].Sub(start)
	
	if elapsed3 > elapsed1 {
		t.Error("task3 should execute before task1")
	}
	if elapsed1 > elapsed2 {
		t.Error("task1 should execute before task2")
	}
}

// TestWorkerPoolStopWithPendingTasks tests stopping with pending tasks
func TestWorkerPoolStopWithPendingTasks(t *testing.T) {
	wp := NewWorkerPool(1)
	wp.Start()
	
	var executed int32
	
	// Submit multiple tasks
	for i := 0; i < 5; i++ {
		task := &WorkerTask{
			ID: fmt.Sprintf("task-%d", i),
			Func: func(ctx context.Context) error {
				time.Sleep(50 * time.Millisecond)
				atomic.AddInt32(&executed, 1)
				return nil
			},
		}
		wp.Submit(task)
	}
	
	// Stop immediately
	wp.Stop()
	
	// Some tasks may have executed, but not necessarily all
	count := atomic.LoadInt32(&executed)
	t.Logf("Executed %d tasks before stop", count)
}

// TestWorkerPoolSubmitWhenStopped tests submitting to stopped pool
func TestWorkerPoolSubmitWhenStopped(t *testing.T) {
	wp := NewWorkerPool(2)
	
	task := &WorkerTask{
		ID: "task",
		Func: func(ctx context.Context) error {
			return nil
		},
	}
	
	// Should fail when not started
	if err := wp.Submit(task); err == nil {
		t.Error("Expected error when submitting to non-running pool")
	}
	
	// Start and stop
	wp.Start()
	wp.Stop()
	
	// Should fail after stop
	if err := wp.Submit(task); err == nil {
		t.Error("Expected error when submitting to stopped pool")
	}
}

// TestWorkerPoolConcurrentSubmission tests concurrent task submissions
func TestWorkerPoolConcurrentSubmission(t *testing.T) {
	wp := NewWorkerPool(5)
	wp.Start()
	defer wp.Stop()
	
	var counter int32
	var wg sync.WaitGroup
	
	numGoroutines := 10
	tasksPerGoroutine := 10
	totalTasks := numGoroutines * tasksPerGoroutine
	
	wg.Add(totalTasks)
	
	// Submit tasks from multiple goroutines
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < tasksPerGoroutine; j++ {
				task := &WorkerTask{
					ID: fmt.Sprintf("task-%d-%d", id, j),
					Func: func(ctx context.Context) error {
						atomic.AddInt32(&counter, 1)
						wg.Done()
						return nil
					},
				}
				wp.Submit(task)
			}
		}(i)
	}
	
	wg.Wait()
	
	if atomic.LoadInt32(&counter) != int32(totalTasks) {
		t.Errorf("Expected %d tasks executed, got %d", totalTasks, counter)
	}
}

// TestWorkerPoolZeroWorkers tests creating pool with zero workers
func TestWorkerPoolZeroWorkers(t *testing.T) {
	wp := NewWorkerPool(0)
	
	// Should default to 1 worker
	if wp.WorkerCount() != 1 {
		t.Errorf("Expected 1 worker, got %d", wp.WorkerCount())
	}
}

// BenchmarkWorkerPoolSubmit benchmarks task submission
func BenchmarkWorkerPoolSubmit(b *testing.B) {
	wp := NewWorkerPool(10)
	wp.Start()
	defer wp.Stop()
	
	var wg sync.WaitGroup
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		task := &WorkerTask{
			ID: fmt.Sprintf("task-%d", i),
			Func: func(ctx context.Context) error {
				wg.Done()
				return nil
			},
		}
		wp.Submit(task)
	}
	wg.Wait()
}

// BenchmarkWorkerPoolDelayed benchmarks delayed task submission
func BenchmarkWorkerPoolDelayed(b *testing.B) {
	wp := NewWorkerPool(10)
	wp.Start()
	defer wp.Stop()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		task := &WorkerTask{
			ID: fmt.Sprintf("task-%d", i),
			Func: func(ctx context.Context) error {
				return nil
			},
		}
		wp.SubmitDelayed(task, time.Millisecond)
	}
}
