package goroutine

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

// TestTaskPriority tests task priority levels
func TestTaskPriority(t *testing.T) {
	priorities := []TaskPriority{PriorityLow, PriorityNormal, PriorityHigh, PriorityCritical}
	
	for _, p := range priorities {
		task := &WorkerTask{
			ID:       fmt.Sprintf("task-priority-%d", p),
			Priority: p,
			Func: func(ctx context.Context) error {
				return nil
			},
		}
		
		if task.Priority != p {
			t.Errorf("Expected priority %d, got %d", p, task.Priority)
		}
	}
}

// TestRetryPolicy tests task retry with exponential backoff
func TestRetryPolicy(t *testing.T) {
	config := DefaultWorkerPoolConfig()
	config.NumWorkers = 2
	config.EnableResults = true
	
	pool := NewWorkerPoolWithConfig(config)
	pool.Start()
	defer pool.Stop()
	
	var attempts int32
	
	task := &WorkerTask{
		ID: "retry-task",
		Func: func(ctx context.Context) error {
			count := atomic.AddInt32(&attempts, 1)
			if count < 3 {
				return fmt.Errorf("simulated failure %d", count)
			}
			return nil
		},
		RetryPolicy: &RetryPolicy{
			MaxRetries:    3,
			RetryDelay:    50 * time.Millisecond,
			BackoffFactor: 1.5,
			MaxRetryDelay: 500 * time.Millisecond,
		},
	}
	
	if err := pool.Submit(task); err != nil {
		t.Fatalf("Failed to submit task: %v", err)
	}
	
	// Wait for retries
	time.Sleep(500 * time.Millisecond)
	
	finalAttempts := atomic.LoadInt32(&attempts)
	if finalAttempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", finalAttempts)
	}
	
	// Check result
	result, exists := pool.GetResult("retry-task")
	if !exists {
		t.Error("Result not found")
	}
	if result.Status != StatusSuccess {
		t.Errorf("Expected success status, got %s", result.Status)
	}
	if result.Attempts != 3 {
		t.Errorf("Expected 3 attempts in result, got %d", result.Attempts)
	}
}

// TestDeadLetterQueue tests dead letter queue for failed tasks
func TestDeadLetterQueue(t *testing.T) {
	config := DefaultWorkerPoolConfig()
	config.NumWorkers = 2
	config.MaxDeadLetter = 10
	
	pool := NewWorkerPoolWithConfig(config)
	pool.Start()
	defer pool.Stop()
	
	// Task that always fails
	task := &WorkerTask{
		ID: "failing-task",
		Func: func(ctx context.Context) error {
			return fmt.Errorf("always fails")
		},
		RetryPolicy: &RetryPolicy{
			MaxRetries:    2,
			RetryDelay:    10 * time.Millisecond,
			BackoffFactor: 1.0,
		},
	}
	
	if err := pool.Submit(task); err != nil {
		t.Fatalf("Failed to submit task: %v", err)
	}
	
	// Wait for task to fail and exhaust retries
	time.Sleep(200 * time.Millisecond)
	
	// Check dead letter queue
	deadTasks := pool.GetDeadLetterTasks()
	if len(deadTasks) != 1 {
		t.Errorf("Expected 1 task in dead letter queue, got %d", len(deadTasks))
	}
	
	if len(deadTasks) > 0 && deadTasks[0].ID != "failing-task" {
		t.Errorf("Expected task ID 'failing-task', got '%s'", deadTasks[0].ID)
	}
}

// TestNamedQueues tests Celery-style named queues
func TestNamedQueues(t *testing.T) {
	config := DefaultWorkerPoolConfig()
	config.NumWorkers = 2
	config.EnableQueues = true
	
	pool := NewWorkerPoolWithConfig(config)
	pool.Start()
	defer pool.Stop()
	
	var highPrioExecuted, lowPrioExecuted int32
	
	highPrioTask := &WorkerTask{
		ID:       "high-prio-task",
		Priority: PriorityHigh,
		Func: func(ctx context.Context) error {
			atomic.StoreInt32(&highPrioExecuted, 1)
			return nil
		},
	}
	
	lowPrioTask := &WorkerTask{
		ID:       "low-prio-task",
		Priority: PriorityLow,
		Func: func(ctx context.Context) error {
			atomic.StoreInt32(&lowPrioExecuted, 1)
			return nil
		},
	}
	
	// Submit to different queues
	if err := pool.SubmitToQueue("high-priority", highPrioTask); err != nil {
		t.Fatalf("Failed to submit high priority task: %v", err)
	}
	
	if err := pool.SubmitToQueue("low-priority", lowPrioTask); err != nil {
		t.Fatalf("Failed to submit low priority task: %v", err)
	}
	
	// Wait for execution
	time.Sleep(100 * time.Millisecond)
	
	if atomic.LoadInt32(&highPrioExecuted) != 1 {
		t.Error("High priority task not executed")
	}
	
	if atomic.LoadInt32(&lowPrioExecuted) != 1 {
		t.Error("Low priority task not executed")
	}
}

// TestTaskResult tests result storage and retrieval
func TestTaskResult(t *testing.T) {
	config := DefaultWorkerPoolConfig()
	config.EnableResults = true
	
	pool := NewWorkerPoolWithConfig(config)
	pool.Start()
	defer pool.Stop()
	
	task := &WorkerTask{
		ID: "result-task",
		Func: func(ctx context.Context) error {
			time.Sleep(50 * time.Millisecond)
			return nil
		},
	}
	
	if err := pool.Submit(task); err != nil {
		t.Fatalf("Failed to submit task: %v", err)
	}
	
	// Wait for task to complete
	time.Sleep(150 * time.Millisecond)
	
	// Retrieve result
	result, exists := pool.GetResult("result-task")
	if !exists {
		t.Fatal("Result not found")
	}
	
	if result.Status != StatusSuccess {
		t.Errorf("Expected success status, got %s", result.Status)
	}
	
	if result.TaskID != "result-task" {
		t.Errorf("Expected task ID 'result-task', got '%s'", result.TaskID)
	}
	
	if result.Attempts != 1 {
		t.Errorf("Expected 1 attempt, got %d", result.Attempts)
	}
}

// TestTaskTimeout tests task timeout functionality
func TestTaskTimeout(t *testing.T) {
	config := DefaultWorkerPoolConfig()
	config.EnableResults = true
	
	pool := NewWorkerPoolWithConfig(config)
	pool.Start()
	defer pool.Stop()
	
	task := &WorkerTask{
		ID:      "timeout-task",
		Timeout: 50 * time.Millisecond,
		Func: func(ctx context.Context) error {
			select {
			case <-time.After(200 * time.Millisecond):
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		},
	}
	
	if err := pool.Submit(task); err != nil {
		t.Fatalf("Failed to submit task: %v", err)
	}
	
	// Wait for task to timeout
	time.Sleep(150 * time.Millisecond)
	
	// Check result
	result, exists := pool.GetResult("timeout-task")
	if !exists {
		t.Fatal("Result not found")
	}
	
	if result.Status != StatusFailed {
		t.Errorf("Expected failed status, got %s", result.Status)
	}
	
	if result.Error == nil {
		t.Error("Expected timeout error")
	}
}

// TestTaskExpiration tests task expiration
func TestTaskExpiration(t *testing.T) {
	config := DefaultWorkerPoolConfig()
	config.EnableResults = true
	
	pool := NewWorkerPoolWithConfig(config)
	pool.Start()
	defer pool.Stop()
	
	expiresAt := time.Now().Add(50 * time.Millisecond)
	
	task := &WorkerTask{
		ID:        "expiring-task",
		ExpiresAt: &expiresAt,
		Func: func(ctx context.Context) error {
			return nil
		},
	}
	
	// Wait for expiration
	time.Sleep(100 * time.Millisecond)
	
	// Submit expired task
	if err := pool.Submit(task); err != nil {
		t.Fatalf("Failed to submit task: %v", err)
	}
	
	// Wait for processing
	time.Sleep(50 * time.Millisecond)
	
	// Check result
	result, exists := pool.GetResult("expiring-task")
	if !exists {
		t.Fatal("Result not found")
	}
	
	if result.Status != StatusExpired {
		t.Errorf("Expected expired status, got %s", result.Status)
	}
}

// TestRequeueDeadLetter tests requeuing tasks from dead letter queue
func TestRequeueDeadLetter(t *testing.T) {
	config := DefaultWorkerPoolConfig()
	config.NumWorkers = 2
	config.EnableResults = true
	
	pool := NewWorkerPoolWithConfig(config)
	pool.Start()
	defer pool.Stop()
	
	var shouldFail int32 = 1
	
	task := &WorkerTask{
		ID: "requeue-task",
		Func: func(ctx context.Context) error {
			if atomic.LoadInt32(&shouldFail) == 1 {
				return fmt.Errorf("failing as requested")
			}
			return nil
		},
		RetryPolicy: &RetryPolicy{
			MaxRetries:    1,
			RetryDelay:    10 * time.Millisecond,
			BackoffFactor: 1.0,
		},
	}
	
	if err := pool.Submit(task); err != nil {
		t.Fatalf("Failed to submit task: %v", err)
	}
	
	// Wait for task to fail
	time.Sleep(100 * time.Millisecond)
	
	// Check dead letter queue
	deadTasks := pool.GetDeadLetterTasks()
	if len(deadTasks) != 1 {
		t.Fatalf("Expected 1 task in dead letter queue, got %d", len(deadTasks))
	}
	
	// Change condition so task will succeed on requeue
	atomic.StoreInt32(&shouldFail, 0)
	
	// Requeue the task
	if err := pool.RequeueDeadLetter("requeue-task"); err != nil {
		t.Fatalf("Failed to requeue task: %v", err)
	}
	
	// Wait for requeued task to succeed
	time.Sleep(100 * time.Millisecond)
	
	// Check result
	result, exists := pool.GetResult("requeue-task")
	if !exists {
		t.Fatal("Result not found after requeue")
	}
	
	if result.Status != StatusSuccess {
		t.Errorf("Expected success status after requeue, got %s", result.Status)
	}
	
	// Check dead letter queue is now empty
	deadTasks = pool.GetDeadLetterTasks()
	if len(deadTasks) != 0 {
		t.Errorf("Expected empty dead letter queue after requeue, got %d tasks", len(deadTasks))
	}
}

// TestTaskHooks tests task lifecycle hooks
func TestTaskHooks(t *testing.T) {
	var started, completed, failed bool
	
	config := DefaultWorkerPoolConfig()
	config.OnTaskStart = func(task *WorkerTask) {
		started = true
	}
	config.OnTaskComplete = func(task *WorkerTask, result *TaskResult) {
		completed = true
	}
	config.OnTaskFailed = func(task *WorkerTask, err error) {
		failed = true
	}
	
	pool := NewWorkerPoolWithConfig(config)
	pool.Start()
	defer pool.Stop()
	
	// Successful task
	successTask := &WorkerTask{
		ID: "success-task",
		Func: func(ctx context.Context) error {
			return nil
		},
	}
	
	if err := pool.Submit(successTask); err != nil {
		t.Fatalf("Failed to submit task: %v", err)
	}
	
	time.Sleep(50 * time.Millisecond)
	
	if !started {
		t.Error("OnTaskStart hook not called")
	}
	
	if !completed {
		t.Error("OnTaskComplete hook not called")
	}
	
	// Reset flags
	started, completed, failed = false, false, false
	
	// Failing task
	failTask := &WorkerTask{
		ID: "fail-task",
		Func: func(ctx context.Context) error {
			return fmt.Errorf("intentional failure")
		},
		RetryPolicy: &RetryPolicy{
			MaxRetries: 0,
		},
	}
	
	if err := pool.Submit(failTask); err != nil {
		t.Fatalf("Failed to submit task: %v", err)
	}
	
	time.Sleep(50 * time.Millisecond)
	
	if !started {
		t.Error("OnTaskStart hook not called for failing task")
	}
	
	if !failed {
		t.Error("OnTaskFailed hook not called")
	}
}

// TestGetQueueLength tests queue length retrieval
func TestGetQueueLength(t *testing.T) {
	config := DefaultWorkerPoolConfig()
	config.NumWorkers = 1
	config.EnableQueues = true
	
	pool := NewWorkerPoolWithConfig(config)
	pool.Start()
	defer pool.Stop()
	
	// Submit tasks to queue
	for i := 0; i < 5; i++ {
		task := &WorkerTask{
			ID: fmt.Sprintf("queue-task-%d", i),
			Func: func(ctx context.Context) error {
				time.Sleep(50 * time.Millisecond)
				return nil
			},
		}
		pool.SubmitToQueue("test-queue", task)
	}
	
	// Check queue length (may be less than 5 as workers process them)
	length := pool.GetQueueLength("test-queue")
	if length < 0 || length > 5 {
		t.Errorf("Unexpected queue length: %d", length)
	}
}

// BenchmarkWorkerPoolWithRetry benchmarks task execution with retry
func BenchmarkWorkerPoolWithRetry(b *testing.B) {
	config := DefaultWorkerPoolConfig()
	config.NumWorkers = 10
	config.EnableResults = true
	
	pool := NewWorkerPoolWithConfig(config)
	pool.Start()
	defer pool.Stop()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		task := &WorkerTask{
			ID: fmt.Sprintf("bench-task-%d", i),
			Func: func(ctx context.Context) error {
				return nil
			},
			RetryPolicy: DefaultRetryPolicy(),
		}
		pool.Submit(task)
	}
}

// BenchmarkNamedQueues benchmarks named queue operations
func BenchmarkNamedQueues(b *testing.B) {
	config := DefaultWorkerPoolConfig()
	config.NumWorkers = 10
	config.EnableQueues = true
	
	pool := NewWorkerPoolWithConfig(config)
	pool.Start()
	defer pool.Stop()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		task := &WorkerTask{
			ID: fmt.Sprintf("bench-queue-task-%d", i),
			Func: func(ctx context.Context) error {
				return nil
			},
		}
		pool.SubmitToQueue("bench-queue", task)
	}
}
