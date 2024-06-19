package main

import (
	"fmt"
	"goroutine"
	"log"
	"strings"
	"time"
)

type FuncType func()

func sum(i int) error {
	time.Sleep(time.Duration(i) * time.Second)
	log.Printf("%d being done", i)
	return nil
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
	manager := goroutine.NewGoManager()
	alphabet := "abcdefghijklmnop"
	for i := 0; i < 7; i++ {
		manager.GO(fmt.Sprintf("abc-%d", i), sum, i)
		manager.GO(fmt.Sprintf("def-%d", i), times, string(alphabet[i]), i)
	}
	manager.Cancel("abc-2")
	manager.Cancel("def-3")
	time.Sleep(3 * time.Second)

	log.Println("Hello World")
}
