package main

import (
	"context"
	"github.com/ivikasavnish/goroutine"
	"log"
	"strings"
	"time"
)

type FuncType func()

func sum(i ...int) int {
	s := 0
	for _, i := range i {
		s = s + i
	}
	log.Println(s)
	return s
}

func times(str string, n int) string {
	var res strings.Builder
	for i := 0; i < n; i++ {
		res.WriteString(str)
	}
	log.Println(res.String())
	return res.String()
}

func main() {
	//manager := goroutine.NewGoManager()
	//alphabet := "abcdefghijklmnop"
	//for i := 0; i < 7; i++ {
	//	manager.GO(fmt.Sprintf("abc-%d", i), sum, i)
	//	manager.GO(fmt.Sprintf("def-%d", i), times, string(alphabet[i]), i)
	//}
	//manager.Cancel("abc-2")
	//manager.Cancel("def-3")
	//time.Sleep(3 * time.Second)

	//log.Println(<-manager.R("sum-app", sum, 1, 2))
	//time.Sleep(1)
	//log.Println("Hello World")
	//sum(1, 2, 3)
	safeCh := goroutine.NewSafeChannel[string](1, time.Second)

	// Context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Send with timeout
	err := safeCh.Send(ctx, "hello")
	if err != nil {
		log.Printf("Send error: %v", err)
	}

	// Receive with timeout
	value, err := safeCh.Receive(ctx)
	if err != nil {
		log.Printf("Receive error: %v", err)
	}
	log.Printf("Received: %s", value)

	// Non-blocking try operations
	err = safeCh.TrySend("world")
	value, err = safeCh.TryReceive()

	// Safe closure
	safeCh.Close()
}
