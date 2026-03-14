package utils

import (
	"encoding/json"
	"fmt"
)

func Dump(data any, args ...any) {
	fmt.Print(ToString(data) + " ")
	if len(args) > 0 {
		for _, arg := range args {
			fmt.Print(ToString(arg) + " ")
		}
	}
	fmt.Println("")
}

func ToString(data any) string {
	switch data.(type) {
	case int:
		return fmt.Sprintf("%d", data)
	case float32:
		return fmt.Sprintf("%f", data)
	case float64:
		return fmt.Sprintf("%f", data)
	case string:
		return data.(string)
	default:
	}
	text, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return ""
	}
	return string(text)
}
