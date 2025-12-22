# Recasting Demo

This example demonstrates the `RecastToJSON` function from the goroutine package, which provides a powerful way to map fields between structs using struct tags.

## Overview

The `RecastToJSON` function maps fields from a source struct to a destination struct by matching `recast` tags in the source struct with `json` tags in the destination struct.

## Features Demonstrated

### 1. Basic Field Mapping
Maps fields between structs using matching tag names. When a source field's `recast` tag matches a destination field's `json` tag, the value is copied.

### 2. Type Conversion
Automatically converts between compatible types:
- **Int to Int**: Direct conversion between integer types
- **String to String**: Direct string copying
- **String to Int**: Parses string values and converts them to integers

### 3. Real-World Use Cases
Shows practical applications like:
- Mapping database models to API responses
- Converting form data to typed structures
- Preparing data for JSON serialization

### 4. Partial Mapping
Demonstrates that:
- Only fields with matching tags are mapped
- Fields without matching tags remain unchanged
- Unmapped fields in the destination retain their original values

## Running the Example

```bash
cd example/recasting_demo
go run main.go
```

## How It Works

### Source Struct
Define your source struct with `recast` tags:

```go
type Source struct {
    Name  string `recast:"name"`
    Age   int    `recast:"age"`
    Email string `recast:"email"`
}
```

### Destination Struct
Define your destination struct with `json` tags:

```go
type Destination struct {
    Name  string `json:"name"`
    Age   int    `json:"age"`
    Email string `json:"email"`
}
```

### Perform Recasting

```go
src := Source{Name: "John", Age: 30, Email: "john@example.com"}
dst := Destination{}

goroutine.RecastToJSON(&src, &dst)
// dst now contains: {Name:John Age:30 Email:john@example.com}
```

## Supported Type Conversions

| Source Type | Destination Type | Supported |
|-------------|------------------|-----------|
| string      | string          | ✓         |
| int         | int             | ✓         |
| string      | int             | ✓         |

## Use Cases

1. **API Response Mapping**: Convert internal database models to external API response structures
2. **Form Data Processing**: Parse string form data into typed structures
3. **Data Transfer Objects**: Create DTOs from domain models
4. **Configuration Mapping**: Map configuration structures between different representations

## Notes

- Both source and destination must be passed as pointers
- Only exported (capitalized) fields can be mapped
- Fields without matching tags are ignored
- Type conversion is automatic for supported types
- The function uses reflection for runtime type inspection and conversion
