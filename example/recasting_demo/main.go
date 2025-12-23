package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/ivikasavnish/goroutine"
)

func main() {
	fmt.Println("=== Recasting Examples ===\n")

	demo1BasicFieldMapping()
	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	demo2TypeConversion()
	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	demo3RealWorldDatabaseToAPI()
	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	demo4ComplexStructMapping()
	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	demo5StringToIntConversion()
	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	demo6MixedTypes()
	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	demo7PartialMapping()
	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	demo8JSONOutput()
	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	demo9FieldOmission()
	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	demo10FunctionalTransformation()
	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	demo11FieldRenaming()
	fmt.Println("\n" + strings.Repeat("=", 50) + "\n")

	demo12MapOutput()
}

// Example 1: Basic field mapping
func demo1BasicFieldMapping() {
	fmt.Println("Demo 1: Basic Field Mapping")
	fmt.Println("Demonstrates basic field mapping between structs using recast and json tags\n")

	// Source struct with recast tags
	type Source struct {
		FirstName string `recast:"first_name"`
		LastName  string `recast:"last_name"`
		Email     string `recast:"email"`
	}

	// Destination struct with json tags
	type Destination struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
	}

	// Create source with data
	src := Source{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
	}

	// Create destination
	dst := Destination{}

	fmt.Printf("Before recasting:\n")
	fmt.Printf("  Source: %+v\n", src)
	fmt.Printf("  Destination: %+v\n\n", dst)

	// Perform recasting
	goroutine.RecastToJSON(&src, &dst)

	fmt.Printf("After recasting:\n")
	fmt.Printf("  Source: %+v\n", src)
	fmt.Printf("  Destination: %+v\n", dst)
}

// Example 2: Type conversion (int to int)
func demo2TypeConversion() {
	fmt.Println("Demo 2: Type Conversion")
	fmt.Println("Demonstrates automatic type conversion between compatible types\n")

	type Source struct {
		Age   int `recast:"age"`
		Count int `recast:"count"`
		Score int `recast:"score"`
	}

	type Destination struct {
		Age   int `json:"age"`
		Count int `json:"count"`
		Score int `json:"score"`
	}

	src := Source{
		Age:   30,
		Count: 100,
		Score: 95,
	}

	dst := Destination{}

	fmt.Printf("Before recasting:\n")
	fmt.Printf("  Source: %+v\n", src)
	fmt.Printf("  Destination: %+v\n\n", dst)

	goroutine.RecastToJSON(&src, &dst)

	fmt.Printf("After recasting:\n")
	fmt.Printf("  Source: %+v\n", src)
	fmt.Printf("  Destination: %+v\n", dst)
}

// Example 3: Real-world scenario - Database model to API response
func demo3RealWorldDatabaseToAPI() {
	fmt.Println("Demo 3: Real-World Scenario - Database Model to API Response")
	fmt.Println("Maps a database model to an API response structure\n")

	// Database model (internal representation)
	type UserModel struct {
		ID        int       `recast:"id"`
		Username  string    `recast:"username"`
		Email     string    `recast:"email"`
		IsActive  string    `recast:"is_active"`
		CreatedAt time.Time `recast:"-"` // Not mapped
	}

	// API response (external representation)
	type UserAPIResponse struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
		Email    string `json:"email"`
		IsActive string `json:"is_active"`
	}

	// Simulated database record
	dbUser := UserModel{
		ID:        42,
		Username:  "johndoe",
		Email:     "john@example.com",
		IsActive:  "true",
		CreatedAt: time.Now(),
	}

	// Prepare API response
	apiResponse := UserAPIResponse{}

	fmt.Printf("Database Model:\n")
	fmt.Printf("  %+v\n\n", dbUser)

	// Convert database model to API response
	goroutine.RecastToJSON(&dbUser, &apiResponse)

	fmt.Printf("API Response:\n")
	fmt.Printf("  %+v\n", apiResponse)
	fmt.Printf("\nNote: CreatedAt field was not mapped (tag is '-')\n")
}

// Example 4: Complex struct with multiple types
func demo4ComplexStructMapping() {
	fmt.Println("Demo 4: Complex Struct Mapping")
	fmt.Println("Demonstrates mapping structs with various field types\n")

	type Product struct {
		ProductID   int    `recast:"product_id"`
		Name        string `recast:"name"`
		Description string `recast:"description"`
		Price       int    `recast:"price"`
		InStock     int    `recast:"in_stock"`
		Category    string `recast:"category"`
	}

	type ProductResponse struct {
		ProductID   int    `json:"product_id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Price       int    `json:"price"`
		InStock     int    `json:"in_stock"`
		Category    string `json:"category"`
	}

	product := Product{
		ProductID:   1001,
		Name:        "Laptop",
		Description: "High-performance laptop with 16GB RAM",
		Price:       1299,
		InStock:     25,
		Category:    "Electronics",
	}

	response := ProductResponse{}

	fmt.Printf("Product (internal):\n")
	fmt.Printf("  %+v\n\n", product)

	goroutine.RecastToJSON(&product, &response)

	fmt.Printf("Product Response (API):\n")
	fmt.Printf("  %+v\n", response)
}

// Example 5: String to int conversion
func demo5StringToIntConversion() {
	fmt.Println("Demo 5: String to Int Conversion")
	fmt.Println("Demonstrates automatic conversion from string to integer\n")

	type FormData struct {
		Age   string `recast:"age"`
		Count string `recast:"count"`
		Score string `recast:"score"`
	}

	type ProcessedData struct {
		Age   int `json:"age"`
		Count int `json:"count"`
		Score int `json:"score"`
	}

	// Form data comes as strings
	formData := FormData{
		Age:   "25",
		Count: "42",
		Score: "98",
	}

	processed := ProcessedData{}

	fmt.Printf("Form Data (strings):\n")
	fmt.Printf("  %+v\n\n", formData)

	goroutine.RecastToJSON(&formData, &processed)

	fmt.Printf("Processed Data (integers):\n")
	fmt.Printf("  %+v\n", processed)
	fmt.Printf("\nNote: String values were automatically converted to integers\n")
}

// Example 6: Mixed types and selective mapping
func demo6MixedTypes() {
	fmt.Println("Demo 6: Mixed Types")
	fmt.Println("Demonstrates mapping with multiple different types in one struct\n")

	type UserProfile struct {
		Username    string `recast:"username"`
		Email       string `recast:"email"`
		Age         int    `recast:"age"`
		IsVerified  string `recast:"is_verified"`
		Score       int    `recast:"score"`
		Level       int    `recast:"level"`
	}

	type ProfileResponse struct {
		Username   string `json:"username"`
		Email      string `json:"email"`
		Age        int    `json:"age"`
		IsVerified string `json:"is_verified"`
		Score      int    `json:"score"`
		Level      int    `json:"level"`
	}

	profile := UserProfile{
		Username:   "alice_wonderland",
		Email:      "alice@example.com",
		Age:        28,
		IsVerified: "yes",
		Score:      1500,
		Level:      10,
	}

	response := ProfileResponse{}

	fmt.Printf("User Profile:\n")
	fmt.Printf("  %+v\n\n", profile)

	goroutine.RecastToJSON(&profile, &response)

	fmt.Printf("Profile Response:\n")
	fmt.Printf("  %+v\n", response)
}

// Example 7: Partial mapping - not all fields need to match
func demo7PartialMapping() {
	fmt.Println("Demo 7: Partial Mapping")
	fmt.Println("Demonstrates that only matching fields are mapped; others remain unchanged\n")

	type SourceData struct {
		FieldA string `recast:"field_a"`
		FieldB string `recast:"field_b"`
		FieldC int    `recast:"field_c"`
		FieldD string `recast:"field_d"` // This won't match anything in destination
	}

	type DestinationData struct {
		FieldA string `json:"field_a"`
		FieldB string `json:"field_b"`
		FieldC int    `json:"field_c"`
		FieldE string `json:"field_e"` // This won't match anything in source
	}

	source := SourceData{
		FieldA: "Value A",
		FieldB: "Value B",
		FieldC: 42,
		FieldD: "Value D (won't be mapped)",
	}

	// Pre-populate destination with some values
	dest := DestinationData{
		FieldE: "Existing Value E",
	}

	fmt.Printf("Source Data:\n")
	fmt.Printf("  %+v\n\n", source)

	fmt.Printf("Destination (before recasting):\n")
	fmt.Printf("  %+v\n\n", dest)

	goroutine.RecastToJSON(&source, &dest)

	fmt.Printf("Destination (after recasting):\n")
	fmt.Printf("  %+v\n", dest)
	fmt.Printf("\nNote: FieldD from source was not mapped (no matching json tag)\n")
	fmt.Printf("      FieldE in destination kept its original value (no matching recast tag)\n")
}

// Example 8: JSON output
func demo8JSONOutput() {
	fmt.Println("Demo 8: JSON Output")
	fmt.Println("Demonstrates converting struct directly to JSON string\n")

	type Product struct {
		ID          int    `recast:"id"`
		Name        string `recast:"name"`
		Price       int    `recast:"price"`
		InStock     int    `recast:"in_stock"`
		Description string `recast:"description"`
	}

	product := Product{
		ID:          101,
		Name:        "Wireless Mouse",
		Price:       2999,
		InStock:     150,
		Description: "Ergonomic wireless mouse with 6 buttons",
	}

	fmt.Printf("Product struct:\n")
	fmt.Printf("  %+v\n\n", product)

	// Convert to JSON string
	jsonStr, err := goroutine.RecastToJSONString(&product, nil)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("JSON Output:\n")
	fmt.Printf("  %s\n", jsonStr)
}

// Example 9: Field omission
func demo9FieldOmission() {
	fmt.Println("Demo 9: Field Omission (Negation)")
	fmt.Println("Demonstrates excluding specific fields from output\n")

	type User struct {
		Username  string `recast:"username"`
		Email     string `recast:"email"`
		Password  string `recast:"password"`
		APIKey    string `recast:"api_key"`
		FirstName string `recast:"first_name"`
		LastName  string `recast:"last_name"`
	}

	user := User{
		Username:  "johndoe",
		Email:     "john@example.com",
		Password:  "secret123",
		APIKey:    "key_12345_secret",
		FirstName: "John",
		LastName:  "Doe",
	}

	fmt.Printf("Original User (with sensitive data):\n")
	fmt.Printf("  %+v\n\n", user)

	// Omit sensitive fields
	opts := &goroutine.RecastOptions{
		OmitFields: []string{"password", "api_key"},
	}

	jsonStr, err := goroutine.RecastToJSONString(&user, opts)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("JSON Output (sensitive fields omitted):\n")
	fmt.Printf("  %s\n", jsonStr)
	fmt.Printf("\nNote: password and api_key fields were excluded\n")
}

// Example 10: Functional transformation
func demo10FunctionalTransformation() {
	fmt.Println("Demo 10: Functional Transformation")
	fmt.Println("Demonstrates applying custom transformation functions to fields\n")

	type Product struct {
		Name       string `recast:"name"`
		PriceCents int    `recast:"price"`
		Quantity   int    `recast:"quantity"`
	}

	product := Product{
		Name:       "Laptop",
		PriceCents: 129900, // Price in cents
		Quantity:   5,
	}

	fmt.Printf("Original Product:\n")
	fmt.Printf("  %+v\n\n", product)

	// Transform price from cents to dollars
	opts := &goroutine.RecastOptions{
		TransformFuncs: map[string]goroutine.TransformFunc{
			"price": func(val interface{}) interface{} {
				if priceInCents, ok := val.(int); ok {
					return float64(priceInCents) / 100.0
				}
				return val
			},
			"quantity": func(val interface{}) interface{} {
				if qty, ok := val.(int); ok {
					if qty > 0 {
						return fmt.Sprintf("%d items available", qty)
					}
					return "Out of stock"
				}
				return val
			},
		},
	}

	resultMap := goroutine.RecastToMap(&product, opts)

	fmt.Printf("Transformed Output:\n")
	fmt.Printf("  name: %v\n", resultMap["name"])
	fmt.Printf("  price: $%.2f (converted from cents)\n", resultMap["price"])
	fmt.Printf("  quantity: %v (formatted)\n", resultMap["quantity"])
}

// Example 11: Field renaming
func demo11FieldRenaming() {
	fmt.Println("Demo 11: Field Renaming")
	fmt.Println("Demonstrates renaming fields in the output\n")

	type InternalModel struct {
		UserID    int    `recast:"user_id"`
		UserEmail string `recast:"user_email"`
		UserName  string `recast:"user_name"`
		Status    string `recast:"status"`
	}

	internal := InternalModel{
		UserID:    42,
		UserEmail: "user@example.com",
		UserName:  "alice",
		Status:    "active",
	}

	fmt.Printf("Internal Model:\n")
	fmt.Printf("  %+v\n\n", internal)

	// Rename fields for API output
	opts := &goroutine.RecastOptions{
		RenameFields: map[string]string{
			"user_id":    "id",
			"user_email": "email",
			"user_name":  "username",
		},
	}

	jsonStr, err := goroutine.RecastToJSONString(&internal, opts)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("API Output (with renamed fields):\n")
	fmt.Printf("  %s\n", jsonStr)
	fmt.Printf("\nNote: user_id → id, user_email → email, user_name → username\n")
}

// Example 12: Map output for flexible structures
func demo12MapOutput() {
	fmt.Println("Demo 12: Map Output")
	fmt.Println("Demonstrates converting to map[string]interface{} for flexible manipulation\n")

	type Event struct {
		EventID   int    `recast:"event_id"`
		EventName string `recast:"event_name"`
		Location  string `recast:"location"`
		Capacity  int    `recast:"capacity"`
		Available int    `recast:"available"`
	}

	event := Event{
		EventID:   201,
		EventName: "Tech Conference 2024",
		Location:  "Convention Center",
		Capacity:  500,
		Available: 125,
	}

	fmt.Printf("Event struct:\n")
	fmt.Printf("  %+v\n\n", event)

	// Convert to map with transformations
	opts := &goroutine.RecastOptions{
		RenameFields: map[string]string{
			"event_id":   "id",
			"event_name": "name",
		},
		TransformFuncs: map[string]goroutine.TransformFunc{
			"available": func(val interface{}) interface{} {
				// Calculate percentage
				available, ok := val.(int)
				if !ok {
					return val
				}
				capacity := event.Capacity
				percentage := float64(available) / float64(capacity) * 100
				return fmt.Sprintf("%d (%.1f%% available)", available, percentage)
			},
		},
	}

	resultMap := goroutine.RecastToMap(&event, opts)

	fmt.Printf("Map Output:\n")
	for key, value := range resultMap {
		fmt.Printf("  %s: %v\n", key, value)
	}
	
	fmt.Printf("\nNote: Result is a flexible map[string]interface{} that can be\n")
	fmt.Printf("      further manipulated or used to construct dynamic responses\n")
}
