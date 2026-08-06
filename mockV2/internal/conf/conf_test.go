package conf

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoad(t *testing.T) {
	fmt.Println("-----------------------mock-dubbo.json-------------------------")
	c, err := Load("../../mock-dubbo.json")
	assert.NoError(t, err)
	fmt.Println(c)

	fmt.Println("-----------------------mock-http.json-------------------------")
	c, err = Load("../../mock-http.json")
	assert.NoError(t, err)
	fmt.Println(c)

	fmt.Println("-----------------------mock-http-nacos.json-------------------------")
	c, err = Load("../../mock-http-nacos.json")
	assert.NoError(t, err)
	fmt.Println(c)

	fmt.Println("-----------------------proxy-http.json-------------------------")
	c, err = Load("../../proxy-http.json")
	assert.NoError(t, err)
	fmt.Println(c)
	fmt.Println("------------------------------------------------")
}
