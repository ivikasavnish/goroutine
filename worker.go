package goroutine

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TaskFunc represents a function to be executed by workers
type TaskFunc func(ctx context.Context) error

// WorkerTask represents a unit of work with optional delay and cron schedule
type WorkerTask struct {
	ID          string
	Func        TaskFunc
	Delay       time.Duration
	CronExpr    string
	cronParser  *CronSchedule
	nextRun     time.Time
	IsRecurring bool
}

// CronSchedule parses and manages cron schedules
type CronSchedule struct {
	expr     string
	interval time.Duration
}

// NewCronSchedule creates a new cron schedule from an expression
// Supported formats:
// - "@every 5s" - Run every 5 seconds
// - "@every 1m" - Run every 1 minute
// - "@every 1h" - Run every 1 hour
// - "@hourly" - Run every hour
// - "@daily" - Run every day at midnight
func NewCronSchedule(expr string) (*CronSchedule, error) {
	var interval time.Duration
	
	switch expr {
	case "@hourly":
		interval = time.Hour
	case "@daily":
		interval = 24 * time.Hour
	case "@weekly":
		interval = 7 * 24 * time.Hour
	default:
		// Parse @every format
		var d time.Duration
		var err error
		if len(expr) > 7 && expr[:7] == "@every " {
			d, err = time.ParseDuration(expr[7:])
			if err != nil {
				return nil, fmt.Errorf("invalid cron expression: %w", err)
			}
			interval = d
		} else {
			return nil, fmt.Errorf("unsupported cron expression: %s", expr)
		}
	}
	
	return &CronSchedule{
		expr:     expr,
		interval: interval,
	}, nil
}

// Next returns the next execution time
func (cs *CronSchedule) Next(from time.Time) time.Time {
	return from.Add(cs.interval)
}

// WorkerPool manages a pool of workers that execute tasks
type WorkerPool struct {
	numWorkers   int
	taskQueue    chan *WorkerTask
	delayedQueue *delayedQueue
	cronTasks    map[string]*WorkerTask
	cronMu       sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	stopOnce     sync.Once
	running      bool
	runningMu    sync.RWMutex
}

// delayedQueue manages tasks with delays
type delayedQueue struct {
	tasks    []*WorkerTask
	mu       sync.Mutex
	notifier chan struct{}
}

func newDelayedQueue() *delayedQueue {
	return &delayedQueue{
		tasks:    make([]*WorkerTask, 0),
		notifier: make(chan struct{}, 1),
	}
}

// add adds a task to the delayed queue
func (dq *delayedQueue) add(task *WorkerTask) {
	dq.mu.Lock()
	defer dq.mu.Unlock()
	
	task.nextRun = time.Now().Add(task.Delay)
	dq.tasks = append(dq.tasks, task)
	
	// Notify the scheduler
	select {
	case dq.notifier <- struct{}{}:
	default:
	}
}

// pop returns tasks that are ready to run
func (dq *delayedQueue) pop() []*WorkerTask {
	dq.mu.Lock()
	defer dq.mu.Unlock()
	
	now := time.Now()
	ready := make([]*WorkerTask, 0)
	remaining := make([]*WorkerTask, 0)
	
	for _, task := range dq.tasks {
		if task.nextRun.Before(now) || task.nextRun.Equal(now) {
			ready = append(ready, task)
		} else {
			remaining = append(remaining, task)
		}
	}
	
	dq.tasks = remaining
	return ready
}

// nextDelay returns the duration until the next task is ready
func (dq *delayedQueue) nextDelay() time.Duration {
	dq.mu.Lock()
	defer dq.mu.Unlock()
	
	if len(dq.tasks) == 0 {
		return time.Hour // Default wait time when no tasks
	}
	
	// Find the earliest task
	var earliest time.Time
	for i, task := range dq.tasks {
		if i == 0 || task.nextRun.Before(earliest) {
			earliest = task.nextRun
		}
	}
	
	delay := time.Until(earliest)
	if delay < 0 {
		return 0
	}
	return delay
}

// NewWorkerPool creates a new worker pool with the specified number of workers
func NewWorkerPool(numWorkers int) *WorkerPool {
	if numWorkers <= 0 {
		numWorkers = 1
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	return &WorkerPool{
		numWorkers:   numWorkers,
		taskQueue:    make(chan *WorkerTask, numWorkers*2),
		delayedQueue: newDelayedQueue(),
		cronTasks:    make(map[string]*WorkerTask),
		ctx:          ctx,
		cancel:       cancel,
		running:      false,
	}
}

// Start starts the worker pool
func (wp *WorkerPool) Start() {
	wp.runningMu.Lock()
	if wp.running {
		wp.runningMu.Unlock()
		return
	}
	wp.running = true
	wp.runningMu.Unlock()
	
	// Start workers
	for i := 0; i < wp.numWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
	
	// Start delayed task scheduler
	wp.wg.Add(1)
	go wp.delayedScheduler()
	
	// Start cron scheduler
	wp.wg.Add(1)
	go wp.cronScheduler()
}

// worker is the main worker goroutine
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()
	
	for {
		select {
		case <-wp.ctx.Done():
			return
		case task := <-wp.taskQueue:
			if task == nil {
				return
			}
			
			// Execute the task
			if err := task.Func(wp.ctx); err != nil {
				// Task failed, but we continue processing
				// In production, you might want to log this
			}
		}
	}
}

// delayedScheduler manages delayed tasks
func (wp *WorkerPool) delayedScheduler() {
	defer wp.wg.Done()
	
	for {
		delay := wp.delayedQueue.nextDelay()
		timer := time.NewTimer(delay)
		
		select {
		case <-wp.ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			// Check for ready tasks
			ready := wp.delayedQueue.pop()
			for _, task := range ready {
				select {
				case wp.taskQueue <- task:
				case <-wp.ctx.Done():
					timer.Stop()
					return
				}
			}
		case <-wp.delayedQueue.notifier:
			timer.Stop()
			// New task added, recalculate delay
		}
	}
}

// cronScheduler manages recurring cron tasks
func (wp *WorkerPool) cronScheduler() {
	defer wp.wg.Done()
	
	ticker := time.NewTicker(50 * time.Millisecond) // Check more frequently
	defer ticker.Stop()
	
	for {
		select {
		case <-wp.ctx.Done():
			return
		case now := <-ticker.C:
			wp.cronMu.Lock()
			for id, task := range wp.cronTasks {
				if task.nextRun.Before(now) || task.nextRun.Equal(now) {
					// Create a copy of the task for execution
					taskCopy := &WorkerTask{
						ID:   id,
						Func: task.Func,
					}
					
					select {
					case wp.taskQueue <- taskCopy:
						// Schedule next run
						if task.cronParser != nil {
							task.nextRun = task.cronParser.Next(now)
						}
					case <-wp.ctx.Done():
						wp.cronMu.Unlock()
						return
					}
				}
			}
			wp.cronMu.Unlock()
		}
	}
}

// Submit submits a task for immediate execution
func (wp *WorkerPool) Submit(task *WorkerTask) error {
	wp.runningMu.RLock()
	if !wp.running {
		wp.runningMu.RUnlock()
		return fmt.Errorf("worker pool is not running")
	}
	wp.runningMu.RUnlock()
	
	select {
	case wp.taskQueue <- task:
		return nil
	case <-wp.ctx.Done():
		return fmt.Errorf("worker pool is stopped")
	}
}

// SubmitDelayed submits a task for delayed execution
func (wp *WorkerPool) SubmitDelayed(task *WorkerTask, delay time.Duration) error {
	wp.runningMu.RLock()
	if !wp.running {
		wp.runningMu.RUnlock()
		return fmt.Errorf("worker pool is not running")
	}
	wp.runningMu.RUnlock()
	
	task.Delay = delay
	wp.delayedQueue.add(task)
	return nil
}

// SubmitCron submits a recurring task with a cron schedule
func (wp *WorkerPool) SubmitCron(task *WorkerTask, cronExpr string) error {
	wp.runningMu.RLock()
	if !wp.running {
		wp.runningMu.RUnlock()
		return fmt.Errorf("worker pool is not running")
	}
	wp.runningMu.RUnlock()
	
	schedule, err := NewCronSchedule(cronExpr)
	if err != nil {
		return err
	}
	
	task.CronExpr = cronExpr
	task.cronParser = schedule
	task.nextRun = schedule.Next(time.Now())
	task.IsRecurring = true
	
	wp.cronMu.Lock()
	wp.cronTasks[task.ID] = task
	wp.cronMu.Unlock()
	
	return nil
}

// CancelCron cancels a recurring cron task
func (wp *WorkerPool) CancelCron(taskID string) {
	wp.cronMu.Lock()
	delete(wp.cronTasks, taskID)
	wp.cronMu.Unlock()
}

// Stop stops the worker pool and waits for all workers to finish
func (wp *WorkerPool) Stop() {
	wp.stopOnce.Do(func() {
		wp.runningMu.Lock()
		wp.running = false
		wp.runningMu.Unlock()
		
		wp.cancel()
		
		// Close task queue
		close(wp.taskQueue)
		
		// Wait for all workers to finish
		wp.wg.Wait()
	})
}

// IsRunning returns whether the worker pool is running
func (wp *WorkerPool) IsRunning() bool {
	wp.runningMu.RLock()
	defer wp.runningMu.RUnlock()
	return wp.running
}

// WorkerCount returns the number of workers in the pool
func (wp *WorkerPool) WorkerCount() int {
	return wp.numWorkers
}
