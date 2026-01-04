package utils

import "os"

func Getenv(key string, template string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return template
}
