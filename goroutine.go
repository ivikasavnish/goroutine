package goroutine

import (
	"context"
	"log"
	"reflect"
)

type GoManager struct {
	items map[string]context.CancelFunc
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
