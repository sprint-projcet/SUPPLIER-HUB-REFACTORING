package utils

func InterfaceToString(value interface{}) string {
	if str, ok := value.(string); ok {
		return str
	}
	return ""
}
