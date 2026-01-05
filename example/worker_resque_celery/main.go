package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ivikasavnish/goroutine"
)

func main() {
	fmt.Println("=== Worker Pool: Resque & Celery Mode Demo ===")
	fmt.Println()

	// Example 1: Resque mode - Retry with exponential backoff
	fmt.Println("Example 1: Resque Mode - Retry with Exponential Backoff")
	exampleResqueRetry()

	fmt.Println("\n" + strings.Repeat("=", 60) + "\n")

	// Example 2: Resque mode - Dead letter queue
	fmt.Println("Example 2: Resque Mode - Dead Letter Queue")
	exampleDeadLetterQueue()

	fmt.Println("\n" + strings.Repeat("=", 60) + "\n")

	// Example 3: Celery mode - Named queues
	fmt.Println("Example 3: Celery Mode - Named Queues & Priorities")
	exampleNamedQueues()

	fmt.Println("\n" + strings.Repeat("=", 60) + "\n")

	// Example 4: Combined features
	fmt.Println("Example 4: Combined Features - Production Example")
	exampleProductionUse()
}

// Example 1: Resque mode with retry and exponential backoff
func exampleResqueRetry() {
	config := goroutine.DefaultWorkerPoolConfig()
	config.NumWorkers = 3
	config.EnableResults = true
	config.OnTaskStart = func(task *goroutine.WorkerTask) {
		fmt.Printf("  [START] Task %s (attempt %d)\n", task.ID, task.GetAttempts()+1)
	}
	config.OnTaskComplete = func(task *goroutine.WorkerTask, result *goroutine.TaskResult) {
		fmt.Printf("  [SUCCESS] Task %s completed after %d attempts\n", 
			task.ID, result.Attempts)
	}
	
	pool := goroutine.NewWorkerPoolWithConfig(config)
	pool.Start()
	defer pool.Stop()
	
	attempts := 0
	task := &goroutine.WorkerTask{
		ID: "flaky-api-call",
		Func: func(ctx context.Context) error {
			attempts++
			if attempts < 3 {
				return fmt.Errorf("API timeout (attempt %d)", attempts)
			}
			return nil
		},
		RetryPolicy: &goroutine.RetryPolicy{
			MaxRetries:    3,
			RetryDelay:    100 * time.Millisecond,
			BackoffFactor: 2.0,
			MaxRetryDelay: 1 * time.Second,
		},
		Priority: goroutine.PriorityHigh,
	}
	
	fmt.Println("  Submitting task that fails twice before succeeding...")
	pool.Submit(task)
	
	time.Sleep(800 * time.Millisecond)
	
	result, _ := pool.GetResult("flaky-api-call")
	fmt.Printf("  Final status: %s\n", result.Status)
}

// Example 2: Dead letter queue for failed tasks
func exampleDeadLetterQueue() {
	config := goroutine.DefaultWorkerPoolConfig()
	config.NumWorkers = 2
	config.EnableResults = true
	config.MaxDeadLetter = 100
	
	pool := goroutine.NewWorkerPoolWithConfig(config)
	pool.Start()
	defer pool.Stop()
	
	// Task that always fails
	failingTask := &goroutine.WorkerTask{
		ID: "broken-task",
		Func: func(ctx context.Context) error {
			return fmt.Errorf("database connection failed")
		},
		RetryPolicy: &goroutine.RetryPolicy{
			MaxRetries:    2,
			RetryDelay:    50 * time.Millisecond,
			BackoffFactor: 1.5,
		},
	}
	
	fmt.Println("  Submitting task that always fails...")
	pool.Submit(failingTask)
	
	time.Sleep(300 * time.Millisecond)
	
	// Check dead letter queue
	deadTasks := pool.GetDeadLetterTasks()
	fmt.Printf("  Dead letter queue: %d tasks\n", len(deadTasks))
	
	if len(deadTasks) > 0 {
		fmt.Printf("  Failed task ID: %s\n", deadTasks[0].ID)
		fmt.Println("  Requeuing task after fixing the issue...")
		
		// Simulate fixing the issue by replacing the function
		deadTasks[0].Func = func(ctx context.Context) error {
			fmt.Println("  [FIXED] Database connection restored")
			return nil
		}
		
		// Requeue the task
		pool.RequeueDeadLetter("broken-task")
		time.Sleep(100 * time.Millisecond)
		
		result, _ := pool.GetResult("broken-task")
		fmt.Printf("  After requeue - Status: %s\n", result.Status)
	}
}

// Example 3: Celery mode with named queues
func exampleNamedQueues() {
	config := goroutine.DefaultWorkerPoolConfig()
	config.NumWorkers = 5
	config.EnableQueues = true
	config.EnableResults = true
	
	pool := goroutine.NewWorkerPoolWithConfig(config)
	pool.Start()
	defer pool.Stop()
	
	// Submit tasks to different queues
	queues := map[string]goroutine.TaskPriority{
		"critical": goroutine.PriorityCritical,
		"high":     goroutine.PriorityHigh,
		"normal":   goroutine.PriorityNormal,
		"low":      goroutine.PriorityLow,
	}
	
	fmt.Println("  Submitting tasks to different priority queues...")
	
	for queueName, priority := range queues {
		for i := 1; i <= 2; i++ {
			task := &goroutine.WorkerTask{
				ID:       fmt.Sprintf("%s-task-%d", queueName, i),
				Priority: priority,
				Queue:    queueName,
				Func: func(ctx context.Context) error {
					time.Sleep(50 * time.Millisecond)
					return nil
				},
			}
			pool.SubmitToQueue(queueName, task)
		}
	}
	
	time.Sleep(100 * time.Millisecond)
	
	// Check queue lengths
	fmt.Println("\n  Queue lengths:")
	for queueName := range queues {
		length := pool.GetQueueLength(queueName)
		fmt.Printf("    %s: %d pending\n", queueName, length)
	}
	
	time.Sleep(300 * time.Millisecond)
	fmt.Println("\n  All tasks completed")
}

// Example 4: Production use case combining features
func exampleProductionUse() {
	config := goroutine.DefaultWorkerPoolConfig()
	config.NumWorkers = 4
	config.EnableQueues = true
	config.EnableResults = true
	config.MaxDeadLetter = 1000
	
	// Add lifecycle hooks
	config.OnTaskStart = func(task *goroutine.WorkerTask) {
		fmt.Printf("  [%s] Starting: %s (priority: %d)\n", 
			time.Now().Format("15:04:05"), task.ID, task.Priority)
	}
	config.OnTaskComplete = func(task *goroutine.WorkerTask, result *goroutine.TaskResult) {
		duration := result.EndTime.Sub(result.StartTime)
		fmt.Printf("  [%s] Completed: %s (took %v)\n", 
			time.Now().Format("15:04:05"), task.ID, duration.Round(time.Millisecond))
	}
	config.OnTaskFailed = func(task *goroutine.WorkerTask, err error) {
		fmt.Printf("  [%s] Failed: %s - %v\n", 
			time.Now().Format("15:04:05"), task.ID, err)
	}
	
	pool := goroutine.NewWorkerPoolWithConfig(config)
	pool.Start()
	defer pool.Stop()
	
	fmt.Println("  Simulating production workload...")
	fmt.Println()
	
	// 1. High priority email notification
	emailTask := &goroutine.WorkerTask{
		ID:       "send-email-notification",
		Priority: goroutine.PriorityCritical,
		Func: func(ctx context.Context) error {
			time.Sleep(30 * time.Millisecond)
			return nil
		},
		Timeout: 5 * time.Second,
	}
	pool.SubmitToQueue("notifications", emailTask)
	
	// 2. Background report generation with retry
	reportTask := &goroutine.WorkerTask{
		ID:       "generate-monthly-report",
		Priority: goroutine.PriorityNormal,
		Func: func(ctx context.Context) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		},
		RetryPolicy: goroutine.DefaultRetryPolicy(),
		Timeout:     30 * time.Second,
	}
	pool.SubmitToQueue("reports", reportTask)
	
	// 3. Low priority data cleanup with expiration
	expiresAt := time.Now().Add(10 * time.Second)
	cleanupTask := &goroutine.WorkerTask{
		ID:        "cleanup-old-data",
		Priority:  goroutine.PriorityLow,
		ExpiresAt: &expiresAt,
		Func: func(ctx context.Context) error {
			time.Sleep(50 * time.Millisecond)
			return nil
		},
	}
	pool.SubmitToQueue("maintenance", cleanupTask)
	
	// Wait for tasks to complete
	time.Sleep(500 * time.Millisecond)
	
	fmt.Println("\n  Task Results:")
	for _, taskID := range []string{"send-email-notification", "generate-monthly-report", "cleanup-old-data"} {
		if result, exists := pool.GetResult(taskID); exists {
			fmt.Printf("    %s: %s (attempts: %d)\n", 
				taskID, result.Status, result.Attempts)
		}
	}
}
