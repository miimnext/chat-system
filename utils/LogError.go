package utils

import (
	"log"
)

// LogError 用于记录错误信息
func LogError(message string, err error) {
	if err != nil {
		log.Printf("[ERROR] %s: %v", message, err)
	}
}
