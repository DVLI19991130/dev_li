package funcs

import "github.com/google/uuid"

// UUID generates UUID v4
// $(uuid) -> 550e8400-e29b-41d4-a716-446655440000
func UUID(args ...string) string {
	return uuid.New().String()
}
