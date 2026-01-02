// Package goroutine provides advanced concurrent processing utilities for Go.
//
// This package offers powerful tools for managing concurrent operations including:
//   - Async task resolution (Promise-like pattern with Group)
//   - Safe channel operations with timeout support (SafeChannel)
//   - Parallel slice processing with automatic parallelization (SuperSlice)
//   - Flexible goroutine management with cancellation (GoManager)
//   - Type conversion utilities (Recasting)
//
// Key Features:
//
// Async Resolve: Launch multiple async operations and wait for all to complete,
// similar to Promise.all() in JavaScript. Supports timeout-based operations.
//
// SuperSlice: Process large slices efficiently with automatic parallelization
// based on configurable thresholds. Includes worker pool management, in-place
// updates, and support for map/filter/forEach operations.
//
// SafeChannel: Thread-safe channel wrapper with timeout capabilities and
// distributed backend support for building robust concurrent systems.
//
// GoManager: Manage named goroutines with context-based lifecycle management
// and dynamic cancellation support.
//
// Example usage:
//
//	// Async Resolve
//	group := goroutine.NewGroup()
//	var result1, result2 any
//	group.Assign(&result1, func() any { return "done" })
//	group.Assign(&result2, func() any { return 42 })
//	group.Resolve()
//
//	// SuperSlice
//	numbers := []int{1, 2, 3, 4, 5}
//	ss := goroutine.NewSuperSlice(numbers)
//	doubled := ss.Process(func(i int, n int) int { return n * 2 })
//
//	// SafeChannel
//	sc := goroutine.NewSafeChannel[int](10, 5*time.Second)
//	sc.Send(42)
//	value, _ := sc.Receive()
//
//	// GoManager
//	manager := goroutine.NewGoManager()
//	manager.GO("worker", func() { /* work */ })
//	manager.Cancel("worker")
package goroutine

import (
	"context"
	"log"
	"reflect"
)

func init() {
	log.SetFlags(log.Llongfile)
}

type GoManager struct {
	items map[string]context.CancelFunc
}

type Func struct {
	Name     string
	FuncName interface{}
	FuncArgs []interface{}
	args     []reflect.Value
}

func NewFunc(f Func) *Func {
	funcValue := reflect.ValueOf(f.FuncName)
	if funcValue.Kind() != reflect.Func {
		log.Println("fn is not a function", funcValue.Kind())

	}
	args := make([]reflect.Value, funcValue.Type().NumIn())
	for i := 0; i < funcValue.Type().NumIn(); i++ {
		args[i] = reflect.ValueOf(f.FuncArgs[i])
	}
	return &Func{
		Name:     f.Name,
		FuncName: funcValue,
		args:     args,
	}
}

func NewGoManager() *GoManager {
	return &GoManager{
		items: make(map[string]context.CancelFunc),
	}
}

func (GR *GoManager) AddCancelFunc(name string, cancelFunc context.CancelFunc) {
	GR.items[name] = cancelFunc

}

func (GR *GoManager) Cancel(name string) {
	defer log.Printf("%s being cancelled", name)
	gr, ok := GR.items[name]
	if ok {
		gr()
	}

}

func (GR *GoManager) GO(name string, fn interface{}, argv ...interface{}) {
	//generic_ctx, cancel := context.
	cancelCtx, cancel := context.WithCancel(context.Background())
	//defer GR.Cancel(name)
	GR.AddCancelFunc(name, cancel)

	funcValue := reflect.ValueOf(fn)
	if funcValue.Kind() != reflect.Func {
		log.Println("fn is not a function", funcValue.Kind())
		return
	}
	args := make([]reflect.Value, funcValue.Type().NumIn())
	for i := 0; i < funcValue.Type().NumIn(); i++ {
		args[i] = reflect.ValueOf(argv[i])
	}

	log.Printf("%s being staring", name)
	go func(ctx context.Context) {

		funcValue.Call(args)
	}(cancelCtx)

}

func handleContext(ctxgroup ...context.Context) bool {
	if len(ctxgroup) > 0 {
		if ctx, ok := ctxgroup[0].(context.Context); ok {
			// Check if context has cancellation function
			if ctxDone := ctx.Done(); ctxDone == nil {
				return false
			}
			return true
		}
	}
	return false
}

func (gr *GoManager) Poll(f ...Func) {

}
func Null() {

}
