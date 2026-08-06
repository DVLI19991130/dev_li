package funcs

import (
	"fmt"
	"math/rand"
	"strconv"
)

// Random generates random integer in specified range
// $(random,100)           -> integer in [0, 100)
// $(random,10,100)        -> integer in [10, 100)
// $(random,100,1000,10)   -> integer in [100, 1000) with step of 10
func Random(args ...string) string {
	mn, mx, step := 0, 100, 1

	switch len(args) {
	case 0:
		// Default [0, 100)
		return fmt.Sprintf("%d", rand.Intn(100))
	case 1:
		// $(random,100) -> [0, 100)
		mx, _ = strconv.Atoi(args[0])
	case 2:
		// $(random,10,100) -> [10, 100)
		mn, _ = strconv.Atoi(args[0])
		mx, _ = strconv.Atoi(args[1])
	case 3:
		// $(random,100,1000,10) -> [100, 1000) step 10
		mn, _ = strconv.Atoi(args[0])
		mx, _ = strconv.Atoi(args[1])
		step, _ = strconv.Atoi(args[2])
	default:
		return fmt.Sprintf("%d", rand.Intn(100))
	}

	if step <= 0 {
		step = 1
	}

	if mx <= mn {
		return fmt.Sprintf("%d", mn)
	}

	n := rand.Intn((mx-mn)/step)*step + mn
	return fmt.Sprintf("%d", n)
}
