package application

import "github.com/dujiao-next/internal/shared/jsonmap"

func schemaRequiresSession(schema jsonmap.JSON) bool {
	fields, ok := schema["fields"].([]interface{})
	if !ok {
		return false
	}
	for _, value := range fields {
		field, ok := value.(map[string]interface{})
		if !ok {
			continue
		}
		key, _ := field["key"].(string)
		required, _ := field["required"].(bool)
		if key == "session" && required {
			return true
		}
	}
	return false
}
