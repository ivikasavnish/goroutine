package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ivikasavnish/goroutine"
)

// Message represents a demo message
type Message struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Content   string    `json:"content"`
	Sender    string    `json:"sender"`
}

func main() {
	backendType := flag.String("backend", "redis", "Backend type: redis or pipe")
	redisAddr := flag.String("redis-addr", "localhost:6379", "Redis address")
	pipePath := flag.String("pipe-path", "/tmp/safechannel.pipe", "Named pipe path")
	demo := flag.String("demo", "producer-consumer", "Demo type: producer-consumer, multiple-producers, subscribe")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n\nShutting down gracefully...")
		cancel()
	}()

	switch *backendType {
	case "redis":
		runRedisDemo(ctx, *redisAddr, *demo)
	case "pipe":
		runPipeDemo(ctx, *pipePath, *demo)
	default:
		fmt.Printf("Unknown backend: %s\n", *backendType)
		os.Exit(1)
	}
}

// runRedisDemo demonstrates Redis backend usage
func runRedisDemo(ctx context.Context, redisAddr string, demoType string) {
	fmt.Println("\n=== Redis Distributed SafeChannel Demo ===")
	fmt.Printf("Connecting to Redis at %s...\n", redisAddr)

	// Create Redis backend
	redisBackend := NewRedisBackend(redisAddr, "messages")
	defer redisBackend.Close()

	// Test connection
	testCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	testMsg := []byte("test")
	if err := redisBackend.Send(testCtx, "test-topic", testMsg); err != nil {
		log.Fatalf("Failed to connect to Redis: %v\n", err)
	}
	fmt.Println("✓ Connected to Redis successfully\n")

	// Create distributed channel
	channel := goroutine.NewDistributedSafeChannel[Message](
		redisBackend,
		"demo-messages",
		100,
		5*time.Second,
	)
	defer channel.Close()

	fmt.Printf("Backend Type: %s\n", channel.BackendType())
	fmt.Printf("Demo Type: %s\n\n", demoType)

	switch demoType {
	case "producer-consumer":
		redisProducerConsumerDemo(ctx, channel)
	case "multiple-producers":
		redisMultipleProducersDemo(ctx, channel)
	case "subscribe":
		redisSubscribeDemo(ctx, channel)
	default:
		fmt.Printf("Unknown demo: %s\n", demoType)
	}
}

// runPipeDemo demonstrates OS pipe backend usage
func runPipeDemo(ctx context.Context, pipePath string, demoType string) {
	fmt.Println("\n=== OS Named Pipe Distributed SafeChannel Demo ===")
	fmt.Printf("Using pipe: %s\n", pipePath)

	// Create pipe backend
	pipeBackend := NewPipeBackend(pipePath)
	defer pipeBackend.Close()

	// Create distributed channel
	channel := goroutine.NewDistributedSafeChannel[Message](
		pipeBackend,
		"demo-messages",
		100,
		5*time.Second,
	)
	defer channel.Close()

	fmt.Printf("Backend Type: %s\n", channel.BackendType())
	fmt.Printf("Demo Type: %s\n\n", demoType)

	switch demoType {
	case "producer-consumer":
		pipeProducerConsumerDemo(ctx, channel)
	case "multiple-producers":
		pipeMultipleProducersDemo(ctx, channel)
	case "subscribe":
		pipeSubscribeDemo(ctx, channel)
	default:
		fmt.Printf("Unknown demo: %s\n", demoType)
	}
}

// redisProducerConsumerDemo shows basic producer-consumer pattern
func redisProducerConsumerDemo(ctx context.Context, channel *goroutine.DistributedSafeChannel[Message]) {
	fmt.Println("Starting Producer-Consumer Demo...")
	fmt.Println("Producer: Sending 5 messages")
	fmt.Println("Consumer: Receiving and processing\n")

	done := make(chan bool)
	var wg sync.WaitGroup

	// Producer
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 1; i <= 5; i++ {
			msg := Message{
				ID:        fmt.Sprintf("redis-msg-%d", i),
				Timestamp: time.Now(),
				Content:   fmt.Sprintf("Hello from Redis #%d", i),
				Sender:    "RedisProducer",
			}

			if err := channel.Send(ctx, msg); err != nil {
				log.Printf("Producer error: %v\n", err)
				return
			}

			fmt.Printf("[Producer] Sent: %s - %s\n", msg.ID, msg.Content)
			time.Sleep(200 * time.Millisecond)
		}
		fmt.Println("[Producer] Finished sending all messages\n")
	}()

	// Consumer
	wg.Add(1)
	go func() {
		defer wg.Done()
		consumed := 0
		for consumed < 5 {
			msg, err := channel.Receive(ctx)
			if err != nil {
				log.Printf("Consumer error: %v\n", err)
				return
			}

			fmt.Printf("[Consumer] Received: %s - %s (from %s)\n", msg.ID, msg.Content, msg.Sender)
			consumed++
		}
		fmt.Println("[Consumer] Finished consuming all messages\n")
	}()

	// Wait with timeout
	wg.Wait()
	done <- true
}

// redisMultipleProducersDemo shows multiple producers scenario
func redisMultipleProducersDemo(ctx context.Context, channel *goroutine.DistributedSafeChannel[Message]) {
	fmt.Println("Starting Multiple Producers Demo...")
	fmt.Println("3 Producers: Each sending 3 messages")
	fmt.Println("1 Consumer: Processing all messages\n")

	done := make(chan bool)
	var wg sync.WaitGroup
	mu := sync.Mutex{}
	messageCount := 0

	// Multiple producers
	numProducers := 3
	messagesPerProducer := 3

	for p := 1; p <= numProducers; p++ {
		wg.Add(1)
		go func(producerID int) {
			defer wg.Done()
			for i := 1; i <= messagesPerProducer; i++ {
				msg := Message{
					ID:        fmt.Sprintf("redis-p%d-msg%d", producerID, i),
					Timestamp: time.Now(),
					Content:   fmt.Sprintf("Message from producer %d", producerID),
					Sender:    fmt.Sprintf("RedisProducer-%d", producerID),
				}

				if err := channel.Send(ctx, msg); err != nil {
					log.Printf("Producer %d error: %v\n", producerID, err)
					return
				}

				fmt.Printf("[Producer-%d] Sent: %s\n", producerID, msg.ID)
				time.Sleep(100 * time.Millisecond)
			}
		}(p)
	}

	// Consumer
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			msg, err := channel.Receive(ctx)
			if err != nil {
				log.Printf("Consumer error: %v\n", err)
				return
			}

			mu.Lock()
			messageCount++
			currentCount := messageCount
			mu.Unlock()

			fmt.Printf("[Consumer] [%d] Received: %s from %s\n", currentCount, msg.ID, msg.Sender)

			if currentCount >= numProducers*messagesPerProducer {
				fmt.Println("[Consumer] All messages received\n")
				return
			}
		}
	}()

	// Wait with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		fmt.Println("Demo completed successfully!")
	case <-timeoutCtx.Done():
		fmt.Println("Demo timeout - not all messages received")
	}
}

// redisSubscribeDemo shows subscription pattern
func redisSubscribeDemo(ctx context.Context, channel *goroutine.DistributedSafeChannel[Message]) {
	fmt.Println("Starting Subscribe Demo...")
	fmt.Println("Subscriber: Listening for messages")
	fmt.Println("Producer: Sending 3 messages\n")

	done := make(chan bool)
	var wg sync.WaitGroup

	// Subscriber
	wg.Add(1)
	go func() {
		defer wg.Done()
		received := 0
		maxMessages := 3

		for received < maxMessages {
			msg, err := channel.Receive(ctx)
			if err != nil {
				log.Printf("Subscriber error: %v\n", err)
				return
			}

			fmt.Printf("[Subscriber] Received: %s - %s\n", msg.ID, msg.Content)
			received++
		}
		fmt.Println("[Subscriber] Subscription complete\n")
	}()

	// Producer
	time.Sleep(500 * time.Millisecond) // Let subscriber start listening
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 1; i <= 3; i++ {
			msg := Message{
				ID:        fmt.Sprintf("redis-sub-msg-%d", i),
				Timestamp: time.Now(),
				Content:   fmt.Sprintf("Subscription message #%d", i),
				Sender:    "RedisPublisher",
			}

			if err := channel.Send(ctx, msg); err != nil {
				log.Printf("Publisher error: %v\n", err)
				return
			}

			fmt.Printf("[Publisher] Published: %s\n", msg.ID)
			time.Sleep(200 * time.Millisecond)
		}
		fmt.Println("[Publisher] Publishing complete\n")
	}()

	// Wait with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		fmt.Println("Demo completed successfully!")
	case <-timeoutCtx.Done():
		fmt.Println("Demo timeout")
	}
}

// pipeProducerConsumerDemo shows basic producer-consumer pattern with pipes
func pipeProducerConsumerDemo(ctx context.Context, channel *goroutine.DistributedSafeChannel[Message]) {
	fmt.Println("Starting Producer-Consumer Demo...")
	fmt.Println("Note: OS pipes require reader and writer to be coordinated")
	fmt.Println("Producer: Sending 3 messages")
	fmt.Println("Consumer: Receiving and processing\n")

	done := make(chan bool)
	var wg sync.WaitGroup

	// Producer
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(500 * time.Millisecond) // Give consumer time to open for reading
		for i := 1; i <= 3; i++ {
			msg := Message{
				ID:        fmt.Sprintf("pipe-msg-%d", i),
				Timestamp: time.Now(),
				Content:   fmt.Sprintf("Message through pipe #%d", i),
				Sender:    "PipeProducer",
			}

			if err := channel.Send(ctx, msg); err != nil {
				fmt.Printf("[Producer] Error: %v\n", err)
				return
			}

			fmt.Printf("[Producer] Sent: %s\n", msg.ID)
			time.Sleep(300 * time.Millisecond)
		}
		fmt.Println("[Producer] Finished\n")
	}()

	// Consumer
	wg.Add(1)
	go func() {
		defer wg.Done()
		consumed := 0
		for consumed < 3 {
			msg, err := channel.Receive(ctx)
			if err != nil {
				fmt.Printf("[Consumer] Error: %v\n", err)
				return
			}

			fmt.Printf("[Consumer] Received: %s - %s\n", msg.ID, msg.Content)
			consumed++
		}
		fmt.Println("[Consumer] Finished\n")
	}()

	// Wait with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		fmt.Println("Demo completed successfully!")
	case <-timeoutCtx.Done():
		fmt.Println("Demo timeout")
	}
}

// pipeMultipleProducersDemo shows multiple producers with pipes (sequential)
func pipeMultipleProducersDemo(ctx context.Context, channel *goroutine.DistributedSafeChannel[Message]) {
	fmt.Println("Starting Multiple Producers Demo (Sequential)...")
	fmt.Println("Note: Pipes are FIFO and work best sequentially")
	fmt.Println("2 Producers: Sending messages sequentially")
	fmt.Println("1 Consumer: Processing all messages\n")

	done := make(chan bool)
	var wg sync.WaitGroup
	mu := sync.Mutex{}
	messageCount := 0

	// Consumer starts listening
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(300 * time.Millisecond) // Let producers queue up

		for {
			msg, err := channel.Receive(ctx)
			if err != nil {
				fmt.Printf("[Consumer] Error: %v\n", err)
				return
			}

			mu.Lock()
			messageCount++
			currentCount := messageCount
			mu.Unlock()

			fmt.Printf("[Consumer] [%d] Received: %s from %s\n", currentCount, msg.ID, msg.Sender)

			if currentCount >= 4 { // 2 producers * 2 messages each
				fmt.Println("[Consumer] All messages received\n")
				return
			}
		}
	}()

	// Producers send sequentially
	wg.Add(1)
	go func() {
		defer wg.Done()
		for p := 1; p <= 2; p++ {
			for i := 1; i <= 2; i++ {
				msg := Message{
					ID:        fmt.Sprintf("pipe-p%d-msg%d", p, i),
					Timestamp: time.Now(),
					Content:   fmt.Sprintf("Message from producer %d", p),
					Sender:    fmt.Sprintf("PipeProducer-%d", p),
				}

				if err := channel.Send(ctx, msg); err != nil {
					fmt.Printf("[Producer-%d] Error: %v\n", p, err)
					return
				}

				fmt.Printf("[Producer-%d] Sent: %s\n", p, msg.ID)
				time.Sleep(200 * time.Millisecond)
			}
		}
	}()

	// Wait with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		fmt.Println("Demo completed successfully!")
	case <-timeoutCtx.Done():
		fmt.Println("Demo timeout")
	}
}

// pipeSubscribeDemo shows subscription pattern with pipes
func pipeSubscribeDemo(ctx context.Context, channel *goroutine.DistributedSafeChannel[Message]) {
	fmt.Println("Starting Subscribe Demo (Pipes)...")
	fmt.Println("Note: Pipes are streaming, no pub/sub semantics")
	fmt.Println("Listener: Reading from pipe")
	fmt.Println("Sender: Writing to pipe\n")

	done := make(chan bool)
	var wg sync.WaitGroup

	// Listener
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(500 * time.Millisecond) // Let sender prepare

		received := 0
		for received < 4 {
			msg, err := channel.Receive(ctx)
			if err != nil {
				fmt.Printf("[Listener] Error: %v\n", err)
				return
			}

			fmt.Printf("[Listener] Got: %s - %s\n", msg.ID, msg.Content)
			received++
		}
		fmt.Println("[Listener] Done listening\n")
	}()

	// Sender
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 1; i <= 4; i++ {
			msg := Message{
				ID:        fmt.Sprintf("pipe-stream-msg-%d", i),
				Timestamp: time.Now(),
				Content:   fmt.Sprintf("Stream message #%d", i),
				Sender:    "PipeSender",
			}

			if err := channel.Send(ctx, msg); err != nil {
				fmt.Printf("[Sender] Error: %v\n", err)
				return
			}

			fmt.Printf("[Sender] Sent: %s\n", msg.ID)
			time.Sleep(300 * time.Millisecond)
		}
		fmt.Println("[Sender] Done sending\n")
	}()

	// Wait with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		fmt.Println("Demo completed successfully!")
	case <-timeoutCtx.Done():
		fmt.Println("Demo timeout")
	}
}

// Helper function to print message as JSON
func printMessage(msg Message) string {
	data, _ := json.MarshalIndent(msg, "", "  ")
	return string(data)
}
