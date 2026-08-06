package pkg

import "github.com/google/uuid"

// GenerateTraceID generates trace ID: microsecond timestamp
func GenerateTraceID() string {
	return uuid.New().String()
}
