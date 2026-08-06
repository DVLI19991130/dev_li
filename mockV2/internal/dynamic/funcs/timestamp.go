package funcs

import (
	"fmt"
	"time"
)

// Timestamp generates timestamp
// $(timestamp)       -> second-level timestamp (10 digits)
// $(timestamp,ms)    -> millisecond-level timestamp (13 digits)
// $(timestamp,ns)    -> nanosecond-level timestamp (19 digits)
func Timestamp(args ...string) string {
	if len(args) > 0 && args[0] == "ms" {
		return fmt.Sprintf("%d", time.Now().UnixMilli())
	}
	if len(args) > 0 && args[0] == "ns" {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("%d", time.Now().Unix())
}
