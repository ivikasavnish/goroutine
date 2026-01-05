package goroutine

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

const (
	// cronSchedulerInterval is the interval at which the cron scheduler checks for tasks
	cronSchedulerInterval = 50 * time.Millisecond
)

// TaskFunc represents a function to be executed by workers
type TaskFunc func(ctx context.Context) error

// TaskPriority defines task priority levels
type TaskPriority int

const (
	PriorityLow TaskPriority = iota
	PriorityNormal
	PriorityHigh
	PriorityCritical
)

// TaskStatus represents the execution status of a task
type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusRunning   TaskStatus = "running"
	StatusSuccess   TaskStatus = "success"
	StatusFailed    TaskStatus = "failed"
	StatusRetrying  TaskStatus = "retrying"
	StatusExpired   TaskStatus = "expired"
)

// TaskResult stores the result of task execution
type TaskResult struct {
	TaskID    string
	Status    TaskStatus
	Result    interface{}
	Error     error
	StartTime time.Time
	EndTime   time.Time
	Attempts  int
}

// RetryPolicy defines how tasks should be retried on failure
type RetryPolicy struct {
	MaxRetries     int
	RetryDelay     time.Duration
	BackoffFactor  float64
	MaxRetryDelay  time.Duration
}

// DefaultRetryPolicy returns a default retry policy
func DefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxRetries:    3,
		RetryDelay:    time.Second,
		BackoffFactor: 2.0,
		MaxRetryDelay: 30 * time.Second,
	}
}

// WorkerTask represents a unit of work with optional delay and cron schedule
type WorkerTask struct {
	ID          string
	Func        TaskFunc
	Delay       time.Duration
	CronExpr    string
	cronParser  *CronSchedule
	nextRun     time.Time
	IsRecurring bool
	
	// Resque/Celery compatibility fields
	Priority    TaskPriority
	Queue       string
	RetryPolicy *RetryPolicy
	Timeout     time.Duration
	ExpiresAt   *time.Time
	Args        map[string]interface{}
	
	// Internal tracking
	attempts    int
	status      TaskStatus
	startTime   time.Time
}

// MarshalJSON serializes task metadata (Celery compatibility)
func (wt *WorkerTask) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"id":        wt.ID,
		"queue":     wt.Queue,
		"priority":  wt.Priority,
		"status":    wt.status,
		"args":      wt.Args,
		"attempts":  wt.attempts,
		"expiresAt": wt.ExpiresAt,
	})
}

// GetAttempts returns the number of execution attempts
func (wt *WorkerTask) GetAttempts() int {
	return wt.attempts
}

// GetStatus returns the current task status
func (wt *WorkerTask) GetStatus() TaskStatus {
	return wt.status
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
	
	// Resque/Celery compatibility
	queues       map[string]chan *WorkerTask  // Named queues (Celery mode)
	queuesMu     sync.RWMutex
	results      map[string]*TaskResult       // Result backend (both modes)
	resultsMu    sync.RWMutex
	deadLetter   []*WorkerTask                // Dead letter queue (Resque mode)
	deadLetterMu sync.Mutex
	maxDeadLetter int
	
	// Hooks
	onTaskStart    func(task *WorkerTask)
	onTaskComplete func(task *WorkerTask, result *TaskResult)
	onTaskFailed   func(task *WorkerTask, err error)
}

// WorkerPoolConfig configures the worker pool
type WorkerPoolConfig struct {
	NumWorkers      int
	MaxDeadLetter   int
	EnableQueues    bool  // Enable named queues (Celery mode)
	EnableResults   bool  // Enable result storage (both modes)
	OnTaskStart     func(task *WorkerTask)
	OnTaskComplete  func(task *WorkerTask, result *TaskResult)
	OnTaskFailed    func(task *WorkerTask, err error)
}

// DefaultWorkerPoolConfig returns default configuration
func DefaultWorkerPoolConfig() *WorkerPoolConfig {
	return &WorkerPoolConfig{
		NumWorkers:    1,
		MaxDeadLetter: 1000,
		EnableQueues:  true,
		EnableResults: true,
	}
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
	config := DefaultWorkerPoolConfig()
	config.NumWorkers = numWorkers
	return NewWorkerPoolWithConfig(config)
}

// NewWorkerPoolWithConfig creates a new worker pool with custom configuration
func NewWorkerPoolWithConfig(config *WorkerPoolConfig) *WorkerPool {
	if config.NumWorkers <= 0 {
		config.NumWorkers = 1
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	wp := &WorkerPool{
		numWorkers:     config.NumWorkers,
		taskQueue:      make(chan *WorkerTask, config.NumWorkers*2),
		delayedQueue:   newDelayedQueue(),
		cronTasks:      make(map[string]*WorkerTask),
		ctx:            ctx,
		cancel:         cancel,
		running:        false,
		maxDeadLetter:  config.MaxDeadLetter,
		onTaskStart:    config.OnTaskStart,
		onTaskComplete: config.OnTaskComplete,
		onTaskFailed:   config.OnTaskFailed,
	}
	
	if config.EnableQueues {
		wp.queues = make(map[string]chan *WorkerTask)
	}
	
	if config.EnableResults {
		wp.results = make(map[string]*TaskResult)
	}
	
	wp.deadLetter = make([]*WorkerTask, 0, config.MaxDeadLetter)
	
	return wp
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
			
			wp.executeTask(id, task)
		}
	}
}

// executeTask executes a task with retry logic
func (wp *WorkerPool) executeTask(workerID int, task *WorkerTask) {
	// Check if task is expired
	if task.ExpiresAt != nil && time.Now().After(*task.ExpiresAt) {
		task.status = StatusExpired
		if wp.results != nil {
			wp.storeResult(task, nil, fmt.Errorf("task expired"))
		}
		if wp.onTaskFailed != nil {
			wp.onTaskFailed(task, fmt.Errorf("task expired"))
		}
		return
	}
	
	// Update task status
	task.status = StatusRunning
	task.startTime = time.Now()
	task.attempts++
	
	// Call onTaskStart hook
	if wp.onTaskStart != nil {
		wp.onTaskStart(task)
	}
	
	// Execute with timeout if specified
	ctx := wp.ctx
	var cancel context.CancelFunc
	if task.Timeout > 0 {
		ctx, cancel = context.WithTimeout(wp.ctx, task.Timeout)
		defer cancel()
	}
	
	// Execute the task
	err := task.Func(ctx)
	endTime := time.Now()
	
	if err != nil {
		// Task failed
		task.status = StatusFailed
		log.Printf("Worker %d: Task %s failed (attempt %d): %v", workerID, task.ID, task.attempts, err)
		
		// Try to retry if policy exists
		if task.RetryPolicy != nil && task.attempts < task.RetryPolicy.MaxRetries {
			task.status = StatusRetrying
			wp.scheduleRetry(task)
		} else {
			// Max retries exceeded or no retry policy
			if wp.results != nil {
				wp.storeResult(task, nil, err)
			}
			
			// Add to dead letter queue
			wp.addToDeadLetter(task)
			
			if wp.onTaskFailed != nil {
				wp.onTaskFailed(task, err)
			}
		}
	} else {
		// Task succeeded
		task.status = StatusSuccess
		
		if wp.results != nil {
			wp.storeResult(task, nil, nil)
		}
		
		if wp.onTaskComplete != nil {
			result := &TaskResult{
				TaskID:    task.ID,
				Status:    StatusSuccess,
				StartTime: task.startTime,
				EndTime:   endTime,
				Attempts:  task.attempts,
			}
			wp.onTaskComplete(task, result)
		}
	}
}

// scheduleRetry schedules a task for retry with exponential backoff
func (wp *WorkerPool) scheduleRetry(task *WorkerTask) {
	if task.RetryPolicy == nil {
		return
	}
	
	// Calculate delay with exponential backoff
	delay := task.RetryPolicy.RetryDelay
	for i := 1; i < task.attempts; i++ {
		delay = time.Duration(float64(delay) * task.RetryPolicy.BackoffFactor)
	}
	
	// Cap at max retry delay
	if delay > task.RetryPolicy.MaxRetryDelay {
		delay = task.RetryPolicy.MaxRetryDelay
	}
	
	log.Printf("Task %s will retry in %v (attempt %d/%d)", 
		task.ID, delay, task.attempts, task.RetryPolicy.MaxRetries)
	
	// Schedule retry as delayed task
	go func() {
		time.Sleep(delay)
		select {
		case wp.taskQueue <- task:
		case <-wp.ctx.Done():
		}
	}()
}

// storeResult stores task execution result
func (wp *WorkerPool) storeResult(task *WorkerTask, result interface{}, err error) {
	wp.resultsMu.Lock()
	defer wp.resultsMu.Unlock()
	
	// Use task's current status if already set (e.g., for expired tasks)
	status := task.status
	if status == "" || status == StatusRunning || status == StatusPending {
		if err != nil {
			status = StatusFailed
		} else {
			status = StatusSuccess
		}
	}
	
	wp.results[task.ID] = &TaskResult{
		TaskID:    task.ID,
		Status:    status,
		Result:    result,
		Error:     err,
		StartTime: task.startTime,
		EndTime:   time.Now(),
		Attempts:  task.attempts,
	}
}

// addToDeadLetter adds a failed task to the dead letter queue
func (wp *WorkerPool) addToDeadLetter(task *WorkerTask) {
	wp.deadLetterMu.Lock()
	defer wp.deadLetterMu.Unlock()
	
	if len(wp.deadLetter) < wp.maxDeadLetter {
		wp.deadLetter = append(wp.deadLetter, task)
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
	
	ticker := time.NewTicker(cronSchedulerInterval)
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

// SubmitToQueue submits a task to a named queue (Celery mode)
func (wp *WorkerPool) SubmitToQueue(queueName string, task *WorkerTask) error {
	wp.runningMu.RLock()
	if !wp.running {
		wp.runningMu.RUnlock()
		return fmt.Errorf("worker pool is not running")
	}
	wp.runningMu.RUnlock()
	
	if wp.queues == nil {
		return fmt.Errorf("named queues not enabled")
	}
	
	task.Queue = queueName
	task.status = StatusPending
	
	wp.queuesMu.Lock()
	queue, exists := wp.queues[queueName]
	if !exists {
		// Create new queue
		queue = make(chan *WorkerTask, wp.numWorkers*2)
		wp.queues[queueName] = queue
		
		// Start worker for this queue
		wp.wg.Add(1)
		go wp.queueWorker(queueName, queue)
	}
	wp.queuesMu.Unlock()
	
	select {
	case queue <- task:
		return nil
	case <-wp.ctx.Done():
		return fmt.Errorf("worker pool is stopped")
	}
}

// queueWorker processes tasks from a named queue
func (wp *WorkerPool) queueWorker(queueName string, queue chan *WorkerTask) {
	defer wp.wg.Done()
	
	for {
		select {
		case <-wp.ctx.Done():
			return
		case task := <-queue:
			if task == nil {
				return
			}
			wp.executeTask(0, task)
		}
	}
}

// GetResult retrieves the result of a task (both modes)
func (wp *WorkerPool) GetResult(taskID string) (*TaskResult, bool) {
	if wp.results == nil {
		return nil, false
	}
	
	wp.resultsMu.RLock()
	defer wp.resultsMu.RUnlock()
	
	result, exists := wp.results[taskID]
	return result, exists
}

// GetDeadLetterTasks returns tasks that failed after max retries (Resque mode)
func (wp *WorkerPool) GetDeadLetterTasks() []*WorkerTask {
	wp.deadLetterMu.Lock()
	defer wp.deadLetterMu.Unlock()
	
	// Return a copy to avoid race conditions
	tasks := make([]*WorkerTask, len(wp.deadLetter))
	copy(tasks, wp.deadLetter)
	return tasks
}

// RequeueDeadLetter requeues a task from dead letter queue (Resque mode)
func (wp *WorkerPool) RequeueDeadLetter(taskID string) error {
	wp.deadLetterMu.Lock()
	var taskToRequeue *WorkerTask
	
	for i, task := range wp.deadLetter {
		if task.ID == taskID {
			// Remove from dead letter
			wp.deadLetter = append(wp.deadLetter[:i], wp.deadLetter[i+1:]...)
			
			// Reset task state
			task.attempts = 0
			task.status = StatusPending
			taskToRequeue = task
			break
		}
	}
	wp.deadLetterMu.Unlock()
	
	if taskToRequeue == nil {
		return fmt.Errorf("task %s not found in dead letter queue", taskID)
	}
	
	// Requeue (outside the lock)
	return wp.Submit(taskToRequeue)
}

// AckTask acknowledges task completion (Celery mode)
func (wp *WorkerPool) AckTask(taskID string) {
	// In a full implementation, this would interact with message broker
	// For now, we just mark it in results
	if wp.results != nil {
		wp.resultsMu.Lock()
		if result, exists := wp.results[taskID]; exists {
			result.Status = StatusSuccess
		}
		wp.resultsMu.Unlock()
	}
}

// RejectTask rejects a task and optionally requeues it (Celery mode)
func (wp *WorkerPool) RejectTask(taskID string, requeue bool) error {
	if !requeue {
		// Just mark as failed
		if wp.results != nil {
			wp.resultsMu.Lock()
			if result, exists := wp.results[taskID]; exists {
				result.Status = StatusFailed
			}
			wp.resultsMu.Unlock()
		}
		return nil
	}
	
	// Requeue from dead letter if exists
	return wp.RequeueDeadLetter(taskID)
}

// ClearResults clears all stored task results
func (wp *WorkerPool) ClearResults() {
	if wp.results == nil {
		return
	}
	
	wp.resultsMu.Lock()
	defer wp.resultsMu.Unlock()
	
	wp.results = make(map[string]*TaskResult)
}

// GetQueueLength returns the number of pending tasks in a queue (Celery mode)
func (wp *WorkerPool) GetQueueLength(queueName string) int {
	if wp.queues == nil {
		return len(wp.taskQueue)
	}
	
	wp.queuesMu.RLock()
	defer wp.queuesMu.RUnlock()
	
	if queue, exists := wp.queues[queueName]; exists {
		return len(queue)
	}
	
	return 0
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
