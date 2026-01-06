package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ivikasavnish/goroutine"
)

func main() {
	fmt.Println("=== Worker Pool: Broker Encoding/Decoding Demo ===")
	fmt.Println()

	// Example 1: Register task handlers
	fmt.Println("Example 1: Registering Task Handlers")
	registerHandlers()

	fmt.Println()
	fmt.Println("============================================================")
	fmt.Println()

	// Example 2: Encode task to broker format
	fmt.Println("Example 2: Encode Task to Broker Format")
	exampleEncodeToBroker()

	fmt.Println()
	fmt.Println("============================================================")
	fmt.Println()

	// Example 3: Decode task from broker and execute
	fmt.Println("Example 3: Decode Task from Broker and Execute")
	exampleDecodeAndExecute()

	fmt.Println()
	fmt.Println("============================================================")
	fmt.Println()

	// Example 4: Resque-style broker format
	fmt.Println("Example 4: Resque-Style Format Compatibility")
	exampleResqueFormat()

	fmt.Println()
	fmt.Println("============================================================")
	fmt.Println()

	// Example 5: Celery-style broker format
	fmt.Println("Example 5: Celery-Style Format Compatibility")
	exampleCeleryFormat()
}

// registerHandlers registers task handlers that can be retrieved by name
func registerHandlers() {
	// Register a simple handler
	goroutine.RegisterTaskHandler("send-email", func(ctx context.Context, args map[string]interface{}) error {
		recipient := args["recipient"].(string)
		subject := args["subject"].(string)
		fmt.Printf("  📧 Sending email to: %s\n", recipient)
		fmt.Printf("  📧 Subject: %s\n", subject)
		return nil
	})
	fmt.Println("  ✓ Registered handler: send-email")

	// Register a data processing handler
	goroutine.RegisterTaskHandler("process-order", func(ctx context.Context, args map[string]interface{}) error {
		orderID := int(args["order_id"].(float64))
		fmt.Printf("  📦 Processing order #%d\n", orderID)
		return nil
	})
	fmt.Println("  ✓ Registered handler: process-order")

	// Register a report generation handler
	goroutine.RegisterTaskHandler("generate-report", func(ctx context.Context, args map[string]interface{}) error {
		reportType := args["type"].(string)
		fmt.Printf("  📊 Generating %s report\n", reportType)
		return nil
	})
	fmt.Println("  ✓ Registered handler: generate-report")
}

// exampleEncodeToBroker demonstrates encoding a task to broker format
func exampleEncodeToBroker() {
	task := &goroutine.WorkerTask{
		ID:          "email-task-001",
		HandlerName: "send-email",
		Queue:       "notifications",
		Priority:    goroutine.PriorityHigh,
		Args: map[string]interface{}{
			"recipient": "user@example.com",
			"subject":   "Welcome to our service!",
			"body":      "Thank you for signing up.",
		},
		RetryPolicy: goroutine.DefaultRetryPolicy(),
		Timeout:     30 * time.Second,
	}

	// Encode to broker format
	data, err := task.EncodeToBroker()
	if err != nil {
		fmt.Printf("  ❌ Error encoding task: %v\n", err)
		return
	}

	fmt.Println("  ✓ Task encoded to broker format")
	fmt.Println()
	fmt.Println("  JSON representation:")
	
	// Pretty print the JSON
	var prettyJSON map[string]interface{}
	json.Unmarshal(data, &prettyJSON)
	prettyData, _ := json.MarshalIndent(prettyJSON, "  ", "  ")
	fmt.Println(string(prettyData))

	fmt.Println()
	fmt.Printf("  📏 Size: %d bytes\n", len(data))
	fmt.Println("  💾 Ready to save to Redis, RabbitMQ, or other broker")
}

// exampleDecodeAndExecute demonstrates decoding and executing a task
func exampleDecodeAndExecute() {
	// Simulate receiving this JSON from a broker (e.g., Redis, RabbitMQ)
	brokerJSON := `{
		"id": "order-task-456",
		"handler_name": "process-order",
		"queue": "orders",
		"priority": 2,
		"args": {
			"order_id": 12345,
			"customer_id": 789
		},
		"retry_policy": {
			"MaxRetries": 3,
			"RetryDelay": 1000000000,
			"BackoffFactor": 2.0,
			"MaxRetryDelay": 30000000000
		},
		"timeout": 15000,
		"created_at": "2024-01-01T10:00:00Z"
	}`

	fmt.Println("  📥 Received job from broker:")
	var prettyJSON map[string]interface{}
	json.Unmarshal([]byte(brokerJSON), &prettyJSON)
	prettyData, _ := json.MarshalIndent(prettyJSON, "  ", "  ")
	fmt.Println(string(prettyData))
	fmt.Println()

	// Decode from broker
	task, err := goroutine.DecodeFromBroker([]byte(brokerJSON))
	if err != nil {
		fmt.Printf("  ❌ Error decoding task: %v\n", err)
		return
	}

	fmt.Println("  ✓ Task decoded successfully")
	fmt.Printf("  🔧 Task ID: %s\n", task.ID)
	fmt.Printf("  🔧 Handler: %s\n", task.HandlerName)
	fmt.Printf("  🔧 Queue: %s\n", task.Queue)
	fmt.Println()

	// Execute the task
	fmt.Println("  ▶️  Executing task...")
	if err := task.Func(context.Background()); err != nil {
		fmt.Printf("  ❌ Task execution failed: %v\n", err)
	} else {
		fmt.Println("  ✅ Task executed successfully")
	}
}

// exampleResqueFormat demonstrates Resque-style format compatibility
func exampleResqueFormat() {
	// Resque uses "class" instead of "handler_name"
	resqueJSON := `{
		"id": "resque-task-789",
		"class": "generate-report",
		"queue": "reports",
		"args": {
			"type": "monthly",
			"month": "December"
		}
	}`

	fmt.Println("  📥 Resque-style job format (uses 'class'):")
	var prettyJSON map[string]interface{}
	json.Unmarshal([]byte(resqueJSON), &prettyJSON)
	prettyData, _ := json.MarshalIndent(prettyJSON, "  ", "  ")
	fmt.Println(string(prettyData))
	fmt.Println()

	// Decode - should work with Resque format
	task, err := goroutine.DecodeFromBroker([]byte(resqueJSON))
	if err != nil {
		fmt.Printf("  ❌ Error decoding Resque format: %v\n", err)
		return
	}

	fmt.Println("  ✓ Resque format decoded successfully")
	fmt.Printf("  🔧 Mapped 'class' -> HandlerName: %s\n", task.HandlerName)
	fmt.Println()

	// Execute
	fmt.Println("  ▶️  Executing task...")
	if err := task.Func(context.Background()); err != nil {
		fmt.Printf("  ❌ Task failed: %v\n", err)
	} else {
		fmt.Println("  ✅ Task executed successfully")
	}
}

// exampleCeleryFormat demonstrates Celery-style format compatibility
func exampleCeleryFormat() {
	// Celery uses "task" and "kwargs" instead of "handler_name" and "args"
	eta := time.Now().Add(5 * time.Minute)
	celeryData := map[string]interface{}{
		"id":   "celery-task-abc",
		"task": "send-email",
		"kwargs": map[string]interface{}{
			"recipient": "admin@example.com",
			"subject":   "System Alert",
		},
		"eta":      eta,
		"priority": 3,
	}

	celeryJSON, _ := json.Marshal(celeryData)

	fmt.Println("  📥 Celery-style job format (uses 'task' and 'kwargs'):")
	var prettyJSON map[string]interface{}
	json.Unmarshal(celeryJSON, &prettyJSON)
	prettyData, _ := json.MarshalIndent(prettyJSON, "  ", "  ")
	fmt.Println(string(prettyData))
	fmt.Println()

	// Decode - should work with Celery format
	task, err := goroutine.DecodeFromBroker(celeryJSON)
	if err != nil {
		fmt.Printf("  ❌ Error decoding Celery format: %v\n", err)
		return
	}

	fmt.Println("  ✓ Celery format decoded successfully")
	fmt.Printf("  🔧 Mapped 'task' -> HandlerName: %s\n", task.HandlerName)
	fmt.Printf("  🔧 Mapped 'kwargs' -> Args: %v\n", task.Args)
	fmt.Printf("  🔧 Mapped 'eta' -> Delay: %v\n", task.Delay.Round(time.Second))
	fmt.Println()

	// Execute
	fmt.Println("  ▶️  Executing task...")
	if err := task.Func(context.Background()); err != nil {
		fmt.Printf("  ❌ Task failed: %v\n", err)
	} else {
		fmt.Println("  ✅ Task executed successfully")
	}
}
