// dynamic/flow_No.go - sequence number generator
package funcs

import (
	"fmt"
	"sync/atomic"
	"time"
)

var sequenceNum uint64 = 0

// FlowNo generates order number
// Format: yyMMddHHmmss + 8-digit sequence = 20 digits
// Example: 26033010385954006127
// Uses CAS for lock-free atomic operations, ensuring concurrency safety
func FlowNo(args ...string) string {
	now := time.Now()
	timeStr := now.Format("060102150405")

	sn := atomic.AddUint64(&sequenceNum, 1) % 100000000
	return fmt.Sprintf("%s%08d", timeStr, sn)
}
