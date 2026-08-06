package slb

import (
	"testing"
)

func TestWeightedRandom_SingleAddr(t *testing.T) {
	addrs := []string{"127.0.0.1:8080"}
	result := WeightedRandom(addrs, []int{100})

	if result != "127.0.0.1:8080" {
		t.Errorf("Expected 127.0.0.1:8080, got %s", result)
	}
}

func TestWeightedRandom_EmptyWeights(t *testing.T) {
	addrs := []string{"127.0.0.1:8080", "127.0.0.1:8081", "127.0.0.1:8082"}
	result := WeightedRandom(addrs, nil)

	found := false
	for _, addr := range addrs {
		if result == addr {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Result %s is not in address list", result)
	}
}

func TestWeightedRandom_ZeroTotalWeight(t *testing.T) {
	addrs := []string{"127.0.0.1:8080", "127.0.0.1:8081"}
	weights := []int{0, 0}

	// Run multiple times to verify uniform distribution
	counts := make(map[string]int)
	for i := 0; i < 1000; i++ {
		result := WeightedRandom(addrs, weights)
		counts[result]++
	}

	// When weight is 0, should be uniformly distributed
	if counts["127.0.0.1:8080"] == 0 || counts["127.0.0.1:8081"] == 0 {
		t.Errorf("When weight is 0, should be uniformly distributed, actual distribution: %v", counts)
	}
}

func TestWeightedRandom_EmptyAddrs(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Empty address list should panic")
		}
	}()

	WeightedRandom([]string{}, []int{})
}

func TestWeightedRandom_NormalCase(t *testing.T) {
	addrs := []string{"127.0.0.1:8080", "127.0.0.1:8081"}
	weights := []int{30, 70}

	// Run multiple times to verify weighted effect
	counts := make(map[string]int)
	for i := 0; i < 1000; i++ {
		result := WeightedRandom(addrs, weights)
		counts[result]++
	}

	// 30%:70% ratio should be within reasonable range
	// Allow some tolerance
	if counts["127.0.0.1:8080"] < 200 || counts["127.0.0.1:8080"] > 400 {
		t.Errorf("Weighted ratio abnormal, actual distribution: %v (expected 127.0.0.1:8080 between 200-400)", counts)
	}
	if counts["127.0.0.1:8081"] < 600 || counts["127.0.0.1:8081"] > 800 {
		t.Errorf("Weighted ratio abnormal, actual distribution: %v (expected 127.0.0.1:8081 between 600-800)", counts)
	}
}
