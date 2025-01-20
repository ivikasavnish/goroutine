package main

import (
	_ "github.com/ivikasavnish/goroutine"
)

func zmain() {
	// Example 1: Basic number processing
	//numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	//evenSquares := streamify.FromSlice(numbers).
	//	Filter(func(n int) bool { return n%2 == 0 }).
	//	Map(func(n int) int { return n * n }).
	//	Collect()
	//fmt.Println("Even squares:", evenSquares)
	//
	//// Example 2: String processing
	//words := []string{"hello", "world", "streamify", "processing", "go"}
	//streamify.FromSlice(words).
	//	Filter(func(s string) bool { return len(s) > 4 }).
	//	Map(func(s string) string { return strings.ToUpper(s) }).
	//	ForEach(func(s string) { fmt.Println("Long word:", s) })
	//
	//// Example 3: Custom type processing
	//type Person struct {
	//	Name string
	//	Age  int
	//}
	//
	//people := []Person{
	//	{"Alice", 25},
	//	{"Bob", 30},
	//	{"Charlie", 35},
	//	{"David", 20},
	//}
	//
	//// Calculate average age of people over 25
	//totalAge := streamify.FromSlice(people).
	//	Filter(func(p Person) bool { return p.Age > 25 }).
	//	Map(func(p Person) int { return p.Age }).
	//	Reduce(0, func(acc, curr int) int { return acc + curr })
	//
	//fmt.Println("Total age of filtered people:", totalAge)
}
