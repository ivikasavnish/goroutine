package goroutine

import (
	_ "context"
	_ "fmt"
	"reflect"
	"strconv"
	"strings"
)

func RecastToJSON(src, dst interface{}) {
	srcVal := reflect.Indirect(reflect.ValueOf(src))
	dstVal := reflect.Indirect(reflect.ValueOf(dst))

	for i := 0; i < srcVal.NumField(); i++ {
		srcType := srcVal.Type().Field(i)

		for j := 0; j < dstVal.NumField(); j++ {
			dstType := dstVal.Type().Field(j)

			if srcType.Tag.Get("recast") == dstType.Tag.Get("json") {
				srcField := srcVal.Field(i)
				dstField := dstVal.Field(j)

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
				}

				// Function resolution
				if tagParts := strings.Split(srcType.Tag.Get("recast"), ","); len(tagParts) > 1 && tagParts[1][0:5] == "func:" {
					funcName := tagParts[1][5:]
					method := srcVal.MethodByName(funcName)
					if method.IsValid() {
						result := method.Call([]reflect.Value{})
						if len(result) == 1 {
							dstField.Set(result[0])
						}
					}
				}
			}
		}
	}
}
