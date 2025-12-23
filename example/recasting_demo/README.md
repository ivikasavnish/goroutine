# Recasting Demo

This example demonstrates the recasting functionality from the goroutine package, which provides powerful ways to map, transform, and output struct data using struct tags and custom transformations.

## Overview

The recasting package offers several functions for flexible struct manipulation:
- **RecastToJSON**: Maps fields from source struct to destination struct
- **RecastToJSONString**: Converts struct directly to JSON string
- **RecastToMap**: Converts struct to map[string]interface{} for flexibility
- **RecastToJSONWithOptions**: Advanced mapping with transformations and field control

## Features Demonstrated

### 1. Basic Field Mapping
Maps fields between structs using matching tag names. When a source field's `recast` tag matches a destination field's `json` tag, the value is copied.

### 2. Type Conversion
Automatically converts between compatible types:
- **Int to Int**: Direct conversion between integer types
- **String to String**: Direct string copying
- **String to Int**: Parses string values and converts them to integers

### 3. JSON Output
Directly convert structs to JSON strings without intermediate destination structs:
```go
jsonStr, err := goroutine.RecastToJSONString(&source, nil)
```

### 4. Field Omission (Negation)
Exclude sensitive or unnecessary fields from output:
```go
opts := &goroutine.RecastOptions{
    OmitFields: []string{"password", "api_key"},
}
```

### 5. Functional Transformation
Apply custom transformation functions to fields during mapping:
```go
opts := &goroutine.RecastOptions{
    TransformFuncs: map[string]goroutine.TransformFunc{
        "price": func(val interface{}) interface{} {
            // Convert cents to dollars
            return float64(val.(int)) / 100.0
        },
    },
}
```

### 6. Field Renaming
Rename fields in the output for API consistency:
```go
opts := &goroutine.RecastOptions{
    RenameFields: map[string]string{
        "user_id": "id",
        "user_email": "email",
    },
}
```

### 7. Map Output
Convert to flexible map[string]interface{} for dynamic manipulation:
```go
resultMap := goroutine.RecastToMap(&source, opts)
```

## Running the Example

```bash
cd example/recasting_demo
go run main.go
```

## Basic Usage

### Struct-to-Struct Mapping

```go
type Source struct {
    Name  string `recast:"name"`
    Age   int    `recast:"age"`
    Email string `recast:"email"`
}

type Destination struct {
    Name  string `json:"name"`
    Age   int    `json:"age"`
    Email string `json:"email"`
}

src := Source{Name: "John", Age: 30, Email: "john@example.com"}
dst := Destination{}

goroutine.RecastToJSON(&src, &dst)
// dst now contains: {Name:John Age:30 Email:john@example.com}
```

### Struct-to-JSON

```go
type Product struct {
    ID    int    `recast:"id"`
    Name  string `recast:"name"`
    Price int    `recast:"price"`
}

product := Product{ID: 1, Name: "Laptop", Price: 1299}
jsonStr, err := goroutine.RecastToJSONString(&product, nil)
// jsonStr: {"id":1,"name":"Laptop","price":1299}
```

## Advanced Usage with Options

### Complete Example

```go
type InternalModel struct {
    UserID       int    `recast:"user_id"`
    UserEmail    string `recast:"user_email"`
    PasswordHash string `recast:"password_hash"`
    PriceCents   int    `recast:"price_cents"`
}

internal := InternalModel{
    UserID:       42,
    UserEmail:    "user@example.com",
    PasswordHash: "secret_hash",
    PriceCents:   129900,
}

opts := &goroutine.RecastOptions{
    // Exclude sensitive fields
    OmitFields: []string{"password_hash"},
    
    // Rename fields for API
    RenameFields: map[string]string{
        "user_id":    "id",
        "user_email": "email",
    },
    
    // Transform price from cents to dollars
    TransformFuncs: map[string]goroutine.TransformFunc{
        "price_cents": func(val interface{}) interface{} {
            return float64(val.(int)) / 100.0
        },
    },
}

jsonStr, err := goroutine.RecastToJSONString(&internal, opts)
// Result: {"email":"user@example.com","id":42,"price_cents":1299}
```

## Supported Type Conversions

| Source Type | Destination Type | Supported |
|-------------|------------------|-----------|
| string      | string          | ✓         |
| int         | int             | ✓         |
| string      | int             | ✓         |
| float       | float           | ✓         |
| bool        | bool            | ✓         |

## Use Cases

1. **API Response Mapping**: Convert internal database models to external API response structures
2. **Form Data Processing**: Parse string form data into typed structures
3. **Data Transfer Objects**: Create DTOs from domain models with field transformations
4. **Configuration Mapping**: Map configuration structures between different representations
5. **Sensitive Data Filtering**: Exclude passwords, API keys, and other sensitive fields from output
6. **Dynamic JSON Generation**: Create flexible JSON responses without rigid struct definitions
7. **Price Conversions**: Transform monetary values between different units (cents to dollars)
8. **Field Standardization**: Rename internal field names to API-standard names

## Notes

- Source must be passed as a pointer for struct-to-struct mapping
- Only exported (capitalized) fields can be mapped
- Fields without matching tags are ignored
- Type conversion is automatic for supported types
- The function uses reflection for runtime type inspection and conversion
- Transform functions have access to the field value and can return any type
- Multiple options can be combined for complex transformations

