package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ivikasavnish/goroutine"
)

func main() {
	fmt.Println("=== Worker Pool Demo ===")
	fmt.Println()

	// Example 1: Basic worker pool with immediate tasks
	fmt.Println("Example 1: Immediate Task Execution")
	example1ImmediateTasks()

	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	// Example 2: Delayed task execution
	fmt.Println("Example 2: Delayed Task Execution")
	example2DelayedTasks()

	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	// Example 3: Cron job scheduling
	fmt.Println("Example 3: Cron Job Scheduling")
	example3CronJobs()

	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	// Example 4: Mixed tasks
	fmt.Println("Example 4: Mixed Task Types")
	example4MixedTasks()
}

// Example 1: Immediate task execution
func example1ImmediateTasks() {
	// Create a worker pool with 3 workers
	pool := goroutine.NewWorkerPool(3)
	pool.Start()
	defer pool.Stop()

	fmt.Println("Starting 5 immediate tasks...")
	start := time.Now()

	// Submit multiple tasks
	for i := 1; i <= 5; i++ {
		taskID := i
		task := &goroutine.WorkerTask{
			ID: fmt.Sprintf("task-%d", taskID),
			Func: func(ctx context.Context) error {
				fmt.Printf("  [%s] Task %d executing (worker processing)\n", 
					time.Since(start).Round(time.Millisecond), taskID)
				time.Sleep(100 * time.Millisecond) // Simulate work
				fmt.Printf("  [%s] Task %d completed\n", 
					time.Since(start).Round(time.Millisecond), taskID)
				return nil
			},
		}
		
		if err := pool.Submit(task); err != nil {
			fmt.Printf("Error submitting task: %v\n", err)
		}
	}

	// Wait for all tasks to complete
	time.Sleep(500 * time.Millisecond)
	fmt.Printf("All tasks completed in %v\n", time.Since(start).Round(time.Millisecond))
}

// Example 2: Delayed task execution
func example2DelayedTasks() {
	pool := goroutine.NewWorkerPool(2)
	pool.Start()
	defer pool.Stop()

	fmt.Println("Scheduling delayed tasks...")
	start := time.Now()

	delays := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		50 * time.Millisecond,
	}

	for i, delay := range delays {
		taskID := i + 1
		currentDelay := delay // Capture for closure
		task := &goroutine.WorkerTask{
			ID: fmt.Sprintf("delayed-task-%d", taskID),
			Func: func(ctx context.Context) error {
				elapsed := time.Since(start).Round(time.Millisecond)
				fmt.Printf("  [%s] Delayed task %d executing (delayed by %v)\n", 
					elapsed, taskID, currentDelay)
				return nil
			},
		}
		
		if err := pool.SubmitDelayed(task, delay); err != nil {
			fmt.Printf("Error submitting delayed task: %v\n", err)
		}
		fmt.Printf("  Scheduled task %d to run after %v\n", taskID, delay)
	}

	// Wait for all delayed tasks
	time.Sleep(400 * time.Millisecond)
	fmt.Printf("All delayed tasks completed\n")
}

// Example 3: Cron job scheduling
func example3CronJobs() {
	pool := goroutine.NewWorkerPool(2)
	pool.Start()
	defer pool.Stop()

	fmt.Println("Starting cron jobs...")
	start := time.Now()

	// Cron job that runs every 200ms
	task := &goroutine.WorkerTask{
		ID: "report-cron",
		Func: func(ctx context.Context) error {
			elapsed := time.Since(start).Round(time.Millisecond)
			fmt.Printf("  [%s] Cron job executed - Generating report...\n", elapsed)
			return nil
		},
	}

	if err := pool.SubmitCron(task, "@every 200ms"); err != nil {
		fmt.Printf("Error submitting cron task: %v\n", err)
		return
	}

	fmt.Println("Cron job scheduled to run every 200ms")
	fmt.Println("Let it run for 1 second...")

	// Let cron job run for a while
	time.Sleep(1 * time.Second)

	// Cancel the cron job
	pool.CancelCron("report-cron")
	fmt.Println("Cron job cancelled")

	// Verify it stopped
	time.Sleep(300 * time.Millisecond)
	fmt.Println("Confirmed cron job stopped")
}

// Example 4: Mixed task types
func example4MixedTasks() {
	pool := goroutine.NewWorkerPool(4)
	pool.Start()
	defer pool.Stop()

	fmt.Println("Demonstrating mixed task types...")
	start := time.Now()

	// Immediate task
	immediateTask := &goroutine.WorkerTask{
		ID: "immediate",
		Func: func(ctx context.Context) error {
			fmt.Printf("  [%s] Immediate task: Processing urgent request\n", 
				time.Since(start).Round(time.Millisecond))
			return nil
		},
	}
	pool.Submit(immediateTask)

	// Delayed task
	delayedTask := &goroutine.WorkerTask{
		ID: "delayed",
		Func: func(ctx context.Context) error {
			fmt.Printf("  [%s] Delayed task: Sending scheduled email\n", 
				time.Since(start).Round(time.Millisecond))
			return nil
		},
	}
	pool.SubmitDelayed(delayedTask, 300*time.Millisecond)

	// Recurring cron task
	cronTask := &goroutine.WorkerTask{
		ID: "health-check",
		Func: func(ctx context.Context) error {
			fmt.Printf("  [%s] Cron task: Health check ping\n", 
				time.Since(start).Round(time.Millisecond))
			return nil
		},
	}
	pool.SubmitCron(cronTask, "@every 250ms")

	fmt.Println("Tasks scheduled:")
	fmt.Println("  - Immediate task (runs now)")
	fmt.Println("  - Delayed task (runs in 300ms)")
	fmt.Println("  - Cron task (runs every 250ms)")
	fmt.Println("\nWatching execution...")

	// Let tasks run
	time.Sleep(1 * time.Second)

	// Cleanup
	pool.CancelCron("health-check")
	fmt.Printf("\nAll tasks completed after %v\n", time.Since(start).Round(time.Millisecond))
}
