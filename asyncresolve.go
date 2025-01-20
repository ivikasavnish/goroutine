package goroutine

import (
	"sync"
	"time"
)

type Task[T any] struct {
	fn     func() T
	result T
	done   bool
}

type Group struct {
	tasks []*Task[any]
	mu    sync.Mutex
	wg    sync.WaitGroup
}

func NewGroup() *Group {
	return &Group{
		tasks: make([]*Task[any], 0),
	}
}

func (g *Group) Assign(result *any, fn func() any) {
	task := &Task[any]{
		fn: fn,
	}

	g.mu.Lock()
	g.tasks = append(g.tasks, task)
	g.mu.Unlock()

	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		task.result = task.fn()
		task.done = true
		*result = task.result
	}()
}

func (g *Group) Resolve() {
	g.wg.Wait()
}

func (g *Group) ResolveWithTimeout(timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		g.wg.Wait()
	}()

	select {
	case <-c:
		return true
	case <-time.After(timeout):
		return false
	}
}
