package goroutine

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

// TestRegisterTaskHandler tests task handler registration
func TestRegisterTaskHandler(t *testing.T) {
	handlerCalled := false
	
	handler := func(ctx context.Context, args map[string]interface{}) error {
		handlerCalled = true
		return nil
	}
	
	RegisterTaskHandler("test-handler", handler)
	
	retrieved, err := GetTaskHandler("test-handler")
	if err != nil {
		t.Fatalf("Failed to get registered handler: %v", err)
	}
	
	if retrieved == nil {
		t.Fatal("Retrieved handler is nil")
	}
	
	// Test calling the handler
	retrieved(context.Background(), nil)
	
	if !handlerCalled {
		t.Error("Handler was not called")
	}
}

// TestGetTaskHandlerNotRegistered tests retrieving non-existent handler
func TestGetTaskHandlerNotRegistered(t *testing.T) {
	_, err := GetTaskHandler("non-existent-handler")
	if err == nil {
		t.Error("Expected error for non-existent handler")
	}
}

// TestEncodeToBroker tests encoding task to broker format
func TestEncodeToBroker(t *testing.T) {
	expiresAt := time.Now().Add(time.Hour)
	
	task := &WorkerTask{
		ID:          "test-task-123",
		HandlerName: "process-order",
		Queue:       "orders",
		Priority:    PriorityHigh,
		Args: map[string]interface{}{
			"order_id": 12345,
			"user_id":  "user-456",
		},
		RetryPolicy: &RetryPolicy{
			MaxRetries:    3,
			RetryDelay:    time.Second,
			BackoffFactor: 2.0,
		},
		Timeout:   30 * time.Second,
		Delay:     5 * time.Minute,
		ExpiresAt: &expiresAt,
	}
	
	data, err := task.EncodeToBroker()
	if err != nil {
		t.Fatalf("Failed to encode task: %v", err)
	}
	
	// Verify it's valid JSON
	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Encoded data is not valid JSON: %v", err)
	}
	
	// Check key fields
	if decoded["id"] != "test-task-123" {
		t.Errorf("Expected id 'test-task-123', got %v", decoded["id"])
	}
	
	if decoded["handler_name"] != "process-order" {
		t.Errorf("Expected handler_name 'process-order', got %v", decoded["handler_name"])
	}
	
	if decoded["queue"] != "orders" {
		t.Errorf("Expected queue 'orders', got %v", decoded["queue"])
	}
	
	// Check Resque compatibility field
	if decoded["class"] != "process-order" {
		t.Errorf("Expected class 'process-order' (Resque), got %v", decoded["class"])
	}
	
	// Check Celery compatibility field
	if decoded["task"] != "process-order" {
		t.Errorf("Expected task 'process-order' (Celery), got %v", decoded["task"])
	}
}

// TestEncodeToBrokerWithoutHandlerName tests encoding fails without handler name
func TestEncodeToBrokerWithoutHandlerName(t *testing.T) {
	task := &WorkerTask{
		ID:   "test-task",
		Func: func(ctx context.Context) error { return nil },
	}
	
	_, err := task.EncodeToBroker()
	if err == nil {
		t.Error("Expected error when encoding without HandlerName")
	}
}

// TestDecodeFromBroker tests decoding task from broker format
func TestDecodeFromBroker(t *testing.T) {
	// Register a test handler
	var handlerArgs map[string]interface{}
	RegisterTaskHandler("decode-test-handler", func(ctx context.Context, args map[string]interface{}) error {
		handlerArgs = args
		return nil
	})
	
	expiresAt := time.Now().Add(time.Hour)
	brokerData := BrokerJobData{
		ID:          "decoded-task-456",
		HandlerName: "decode-test-handler",
		Queue:       "test-queue",
		Priority:    PriorityNormal,
		Args: map[string]interface{}{
			"key1": "value1",
			"key2": float64(42),
		},
		RetryPolicy: &RetryPolicy{
			MaxRetries: 2,
		},
		Timeout:   15000, // 15 seconds in milliseconds
		Delay:     5000,  // 5 seconds in milliseconds
		ExpiresAt: &expiresAt,
		CreatedAt: time.Now(),
	}
	
	data, err := json.Marshal(brokerData)
	if err != nil {
		t.Fatalf("Failed to marshal broker data: %v", err)
	}
	
	task, err := DecodeFromBroker(data)
	if err != nil {
		t.Fatalf("Failed to decode from broker: %v", err)
	}
	
	// Verify decoded fields
	if task.ID != "decoded-task-456" {
		t.Errorf("Expected ID 'decoded-task-456', got %s", task.ID)
	}
	
	if task.HandlerName != "decode-test-handler" {
		t.Errorf("Expected HandlerName 'decode-test-handler', got %s", task.HandlerName)
	}
	
	if task.Queue != "test-queue" {
		t.Errorf("Expected Queue 'test-queue', got %s", task.Queue)
	}
	
	if task.Priority != PriorityNormal {
		t.Errorf("Expected Priority Normal, got %d", task.Priority)
	}
	
	if task.Timeout != 15*time.Second {
		t.Errorf("Expected Timeout 15s, got %v", task.Timeout)
	}
	
	if task.Delay != 5*time.Second {
		t.Errorf("Expected Delay 5s, got %v", task.Delay)
	}
	
	// Test that the function works
	if task.Func == nil {
		t.Fatal("Task function is nil")
	}
	
	if err := task.Func(context.Background()); err != nil {
		t.Errorf("Task function failed: %v", err)
	}
	
	// Verify args were passed correctly
	if handlerArgs == nil {
		t.Fatal("Handler args are nil")
	}
	
	if handlerArgs["key1"] != "value1" {
		t.Errorf("Expected key1='value1', got %v", handlerArgs["key1"])
	}
}

// TestDecodeFromBrokerResqueFormat tests decoding from Resque-style format
func TestDecodeFromBrokerResqueFormat(t *testing.T) {
	RegisterTaskHandler("resque-handler", func(ctx context.Context, args map[string]interface{}) error {
		return nil
	})
	
	// Resque-style format uses "class" instead of "handler_name"
	resqueData := map[string]interface{}{
		"id":    "resque-task",
		"class": "resque-handler",
		"queue": "resque-queue",
		"args": map[string]interface{}{
			"data": "test",
		},
	}
	
	data, _ := json.Marshal(resqueData)
	
	task, err := DecodeFromBroker(data)
	if err != nil {
		t.Fatalf("Failed to decode Resque format: %v", err)
	}
	
	if task.HandlerName != "resque-handler" {
		t.Errorf("Expected HandlerName 'resque-handler', got %s", task.HandlerName)
	}
}

// TestDecodeFromBrokerCeleryFormat tests decoding from Celery-style format
func TestDecodeFromBrokerCeleryFormat(t *testing.T) {
	RegisterTaskHandler("celery-handler", func(ctx context.Context, args map[string]interface{}) error {
		return nil
	})
	
	eta := time.Now().Add(10 * time.Minute)
	
	// Celery-style format uses "task" and "kwargs"
	celeryData := map[string]interface{}{
		"id":   "celery-task",
		"task": "celery-handler",
		"kwargs": map[string]interface{}{
			"param1": "value1",
		},
		"eta": eta,
	}
	
	data, _ := json.Marshal(celeryData)
	
	task, err := DecodeFromBroker(data)
	if err != nil {
		t.Fatalf("Failed to decode Celery format: %v", err)
	}
	
	if task.HandlerName != "celery-handler" {
		t.Errorf("Expected HandlerName 'celery-handler', got %s", task.HandlerName)
	}
	
	if task.Args["param1"] != "value1" {
		t.Errorf("Expected param1='value1', got %v", task.Args["param1"])
	}
	
	// Check that ETA was converted to delay
	if task.Delay <= 0 {
		t.Error("Expected positive delay from ETA")
	}
}

// TestDecodeFromBrokerNoHandler tests decoding fails without handler
func TestDecodeFromBrokerNoHandler(t *testing.T) {
	brokerData := map[string]interface{}{
		"id": "no-handler-task",
	}
	
	data, _ := json.Marshal(brokerData)
	
	_, err := DecodeFromBroker(data)
	if err == nil {
		t.Error("Expected error when decoding without handler name")
	}
}

// TestDecodeFromBrokerUnregisteredHandler tests decoding with unregistered handler
func TestDecodeFromBrokerUnregisteredHandler(t *testing.T) {
	brokerData := BrokerJobData{
		ID:          "task-123",
		HandlerName: "unregistered-handler",
	}
	
	data, _ := json.Marshal(brokerData)
	
	_, err := DecodeFromBroker(data)
	if err == nil {
		t.Error("Expected error when decoding with unregistered handler")
	}
}

// TestRoundTripEncodeDecode tests encoding and then decoding a task
func TestRoundTripEncodeDecode(t *testing.T) {
	// Register handler
	var receivedData string
	RegisterTaskHandler("roundtrip-handler", func(ctx context.Context, args map[string]interface{}) error {
		if data, ok := args["data"].(string); ok {
			receivedData = data
		}
		return nil
	})
	
	// Create original task
	originalTask := &WorkerTask{
		ID:          "roundtrip-task",
		HandlerName: "roundtrip-handler",
		Queue:       "test-queue",
		Priority:    PriorityHigh,
		Args: map[string]interface{}{
			"data": "test-data-123",
		},
		RetryPolicy: DefaultRetryPolicy(),
		Timeout:     10 * time.Second,
	}
	
	// Encode to broker
	encoded, err := originalTask.EncodeToBroker()
	if err != nil {
		t.Fatalf("Failed to encode: %v", err)
	}
	
	// Decode from broker
	decodedTask, err := DecodeFromBroker(encoded)
	if err != nil {
		t.Fatalf("Failed to decode: %v", err)
	}
	
	// Verify key fields match
	if decodedTask.ID != originalTask.ID {
		t.Errorf("ID mismatch: expected %s, got %s", originalTask.ID, decodedTask.ID)
	}
	
	if decodedTask.HandlerName != originalTask.HandlerName {
		t.Errorf("HandlerName mismatch: expected %s, got %s", originalTask.HandlerName, decodedTask.HandlerName)
	}
	
	if decodedTask.Queue != originalTask.Queue {
		t.Errorf("Queue mismatch: expected %s, got %s", originalTask.Queue, decodedTask.Queue)
	}
	
	if decodedTask.Priority != originalTask.Priority {
		t.Errorf("Priority mismatch: expected %d, got %d", originalTask.Priority, decodedTask.Priority)
	}
	
	// Test that the function works with the args
	if err := decodedTask.Func(context.Background()); err != nil {
		t.Errorf("Decoded function failed: %v", err)
	}
	
	if receivedData != "test-data-123" {
		t.Errorf("Args not passed correctly: expected 'test-data-123', got '%s'", receivedData)
	}
}

// TestBrokerJobDataJSONCompleteness tests that BrokerJobData includes all necessary fields
func TestBrokerJobDataJSONCompleteness(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(time.Hour)
	eta := now.Add(10 * time.Minute)
	
	data := BrokerJobData{
		ID:          "complete-task",
		HandlerName: "handler",
		Queue:       "queue",
		Priority:    PriorityHigh,
		Args:        map[string]interface{}{"key": "value"},
		RetryPolicy: DefaultRetryPolicy(),
		Timeout:     5000,
		Delay:       1000,
		ExpiresAt:   &expiresAt,
		CronExpr:    "@hourly",
		IsRecurring: true,
		CreatedAt:   now,
		Class:       "handler",
		Task:        "handler",
		Kwargs:      map[string]interface{}{"key": "value"},
		ETA:         &eta,
		Expires:     &expiresAt,
	}
	
	encoded, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal BrokerJobData: %v", err)
	}
	
	var decoded BrokerJobData
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal BrokerJobData: %v", err)
	}
	
	// Verify all fields are preserved
	if decoded.ID != data.ID {
		t.Error("ID not preserved")
	}
	if decoded.HandlerName != data.HandlerName {
		t.Error("HandlerName not preserved")
	}
	if decoded.Queue != data.Queue {
		t.Error("Queue not preserved")
	}
	if decoded.Priority != data.Priority {
		t.Error("Priority not preserved")
	}
}
