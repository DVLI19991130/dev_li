package slb

import "math/rand"

// WeightedRandom weighted random load balancing
// addrs backend address list
// weights weight list, one-to-one correspondence with addrs
func WeightedRandom(addrs []string, weights []int) string {
	if len(addrs) == 1 {
		return addrs[0]
	}

	if len(weights) == 0 {
		return addrs[rand.Intn(len(addrs))]
	}

	total := 0
	for _, w := range weights {
		total += w
	}

	// When total weight is 0, uniformly distributed selection
	if total == 0 {
		return addrs[rand.Intn(len(addrs))]
	}

	// Weighted random selection
	r := rand.Intn(total)
	current := 0
	for i, w := range weights {
		current += w
		if r < current {
			return addrs[i]
		}
	}
	return addrs[0]
}
