package utils

import (
	"fmt"
	"strconv"
)

// InterfaceToString converts primitive types (string, int, float64) or general interfaces to string
func InterfaceToString(val interface{}) string {
	if val == nil {
		return ""
	}
	switch v := val.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case float64:
		return fmt.Sprintf("%v", v)
	}
	return fmt.Sprintf("%v", val)
}
