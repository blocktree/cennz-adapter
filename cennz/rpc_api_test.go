package cennz

import (
	"fmt"
	"testing"
)

const (
	testRpcAPI = "http://xxx.xxx.xxx.xxx:xxxxx"
)

func TestRpcGetCall(t *testing.T) {
	client := NewRpcClient(testRpcAPI, true, "cenzz")
	method := "chain_getBlockHash"

	params := []interface{}{
		0,
	}

	//for i := 0; i <= 10; i++ {
	result, err := client.Call(method, params)
	if err != nil {
		t.Logf("Get Call Result return: \n\t%+v\n", err)
	}

	if result != nil {
		fmt.Println(method, ", result:", result.String() )
	}
	//}
}

func Test_GetRuntimeVersion(t *testing.T) {

	c := NewRpcClient(testRpcAPI, true, symbol)

	r, err := c.GetRuntimeVersion()

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("r:", r)
	}
}

func Test_SendTransaction(t *testing.T) {

	c := NewRpcClient(testRpcAPI, true, symbol)

	r, err := c.sendTransaction("0xa1a1a1a1a")

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("r:", r)
	}
}