package goroutine

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// TransformFunc is a function type for custom field transformations
type TransformFunc func(interface{}) interface{}

// RecastOptions holds configuration for recasting operations
type RecastOptions struct {
	// TransformFuncs maps field names to transformation functions
	TransformFuncs map[string]TransformFunc
	// OmitFields lists fields to exclude from output (supports negation)
	OmitFields []string
	// RenameFields maps source field names to destination field names
	RenameFields map[string]string
}

// createOmitMap creates a map for quick field omission lookups
func createOmitMap(opts *RecastOptions) map[string]bool {
	omitMap := make(map[string]bool)
	if opts != nil {
		for _, field := range opts.OmitFields {
			omitMap[field] = true
		}
	}
	return omitMap
}

// RecastToJSON maps source struct fields to destination struct using recast and json tags
func RecastToJSON(src, dst interface{}) {
	RecastToJSONWithOptions(src, dst, nil)
}

// RecastToJSONWithOptions provides advanced struct-to-struct mapping with transformations
func RecastToJSONWithOptions(src, dst interface{}, opts *RecastOptions) {
	srcVal := reflect.Indirect(reflect.ValueOf(src))
	dstVal := reflect.Indirect(reflect.ValueOf(dst))

	if opts == nil {
		opts = &RecastOptions{}
	}

	omitMap := createOmitMap(opts)

	for i := 0; i < srcVal.NumField(); i++ {
		srcType := srcVal.Type().Field(i)
		recastTag := srcType.Tag.Get("recast")
		
		// Skip if no recast tag or marked to skip
		if recastTag == "" || recastTag == "-" {
			continue
		}
		
		// Parse recast tag
		tagParts := strings.Split(recastTag, ",")
		originalFieldName := tagParts[0]
		fieldName := originalFieldName
		
		// Skip if field is in omit list
		if omitMap[originalFieldName] {
			continue
		}
		
		// Apply field renaming if configured
		if renamed, ok := opts.RenameFields[originalFieldName]; ok {
			fieldName = renamed
		}

		for j := 0; j < dstVal.NumField(); j++ {
			dstType := dstVal.Type().Field(j)

			if fieldName == dstType.Tag.Get("json") {
				srcField := srcVal.Field(i)
				dstField := dstVal.Field(j)

				// Check for functional transformation in tag
				var transformedValue interface{}
				hasFunc := false
				
				if len(tagParts) > 1 {
					for _, part := range tagParts[1:] {
						if strings.HasPrefix(part, "func:") {
							funcName := part[5:]
							method := srcVal.MethodByName(funcName)
							if method.IsValid() {
								result := method.Call([]reflect.Value{})
								if len(result) == 1 {
									transformedValue = result[0].Interface()
									hasFunc = true
								}
							}
						}
					}
				}
				
				// Apply custom transformation function if provided (use original field name)
				if opts.TransformFuncs != nil {
					if transformFunc, ok := opts.TransformFuncs[originalFieldName]; ok {
						var valueToTransform interface{}
						if hasFunc {
							valueToTransform = transformedValue
						} else {
							valueToTransform = srcField.Interface()
						}
						transformedValue = transformFunc(valueToTransform)
						hasFunc = true
					}
				}

				// Set the value
				if hasFunc {
					setValue(dstField, transformedValue)
				} else {
					// Basic type conversions
					switch dstField.Kind() {
					case reflect.String:
						dstField.SetString(srcField.String())
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						if srcField.Kind() == reflect.String {
							if val, err := strconv.Atoi(srcField.String()); err == nil {
								dstField.SetInt(int64(val))
							}
						} else {
							dstField.SetInt(srcField.Int())
						}
					default:
						if srcField.Type().AssignableTo(dstField.Type()) {
							dstField.Set(srcField)
						}
					}
				}
			}
		}
	}
}

// RecastToMap converts a struct to a map[string]interface{} with field mapping and transformations
func RecastToMap(src interface{}, opts *RecastOptions) map[string]interface{} {
	result := make(map[string]interface{})
	srcVal := reflect.Indirect(reflect.ValueOf(src))

	if opts == nil {
		opts = &RecastOptions{}
	}

	omitMap := createOmitMap(opts)

	for i := 0; i < srcVal.NumField(); i++ {
		srcType := srcVal.Type().Field(i)
		recastTag := srcType.Tag.Get("recast")
		
		if recastTag == "" || recastTag == "-" {
			continue
		}
		
		// Parse recast tag
		tagParts := strings.Split(recastTag, ",")
		fieldName := tagParts[0]
		
		// Skip if field is in omit list
		if omitMap[fieldName] {
			continue
		}
		
		// Apply field renaming if configured
		outputName := fieldName
		if renamed, ok := opts.RenameFields[fieldName]; ok {
			outputName = renamed
		}

		srcField := srcVal.Field(i)
		var value interface{}
		
		// Check for functional transformation in tag
		hasFunc := false
		if len(tagParts) > 1 {
			for _, part := range tagParts[1:] {
				if strings.HasPrefix(part, "func:") {
					funcName := part[5:]
					method := srcVal.MethodByName(funcName)
					if method.IsValid() {
						result := method.Call([]reflect.Value{})
						if len(result) == 1 {
							value = result[0].Interface()
							hasFunc = true
						}
					}
				}
			}
		}
		
		if !hasFunc {
			value = srcField.Interface()
		}
		
		// Apply custom transformation function if provided
		if opts.TransformFuncs != nil {
			if transformFunc, ok := opts.TransformFuncs[fieldName]; ok {
				value = transformFunc(value)
			}
		}
		
		result[outputName] = value
	}
	
	return result
}

// RecastToJSONBytes converts a struct to JSON bytes with field mapping and transformations
func RecastToJSONBytes(src interface{}, opts *RecastOptions) ([]byte, error) {
	resultMap := RecastToMap(src, opts)
	return json.Marshal(resultMap)
}

// RecastToJSONString converts a struct to JSON string with field mapping and transformations
func RecastToJSONString(src interface{}, opts *RecastOptions) (string, error) {
	bytes, err := RecastToJSONBytes(src, opts)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// setValue sets a reflect.Value with proper type conversion
func setValue(field reflect.Value, value interface{}) {
	if !field.CanSet() {
		return
	}
	
	val := reflect.ValueOf(value)
	if val.Type().AssignableTo(field.Type()) {
		field.Set(val)
	} else {
		// Try type conversion
		switch field.Kind() {
		case reflect.String:
			field.SetString(fmt.Sprint(value))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if v, ok := value.(int); ok {
				field.SetInt(int64(v))
			} else if v, ok := value.(int64); ok {
				field.SetInt(v)
			} else if v, ok := value.(string); ok {
				if intVal, err := strconv.Atoi(v); err == nil {
					field.SetInt(int64(intVal))
				}
			}
		case reflect.Bool:
			if v, ok := value.(bool); ok {
				field.SetBool(v)
			}
		case reflect.Float32, reflect.Float64:
			if v, ok := value.(float64); ok {
				field.SetFloat(v)
			} else if v, ok := value.(float32); ok {
				field.SetFloat(float64(v))
			}
		}
	}
}
