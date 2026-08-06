package dynamic

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProcess(t *testing.T) {
	value := Process("aaa")
	assert.Equal(t, value, "aaa")

	value = Process("${aa}")
	assert.Equal(t, value, "${aa}")

	value = Process("$(createOrderNo)")
	fmt.Println(value)
}
