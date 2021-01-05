package cennz

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

const (
	//testNodeAPI = "https://service.eks.centralityapp.com/cennznet-explorer-api" //官方
	testNodeAPI = "http://xxx.xxx.xxx.xxx:xxxxx" //local
	symbol = "CENNZ"
)

func PrintJsonLog(t *testing.T, logCont string){
	if strings.HasPrefix(logCont, "{") {
		var str bytes.Buffer
		_ = json.Indent(&str, []byte(logCont), "", "    ")
		t.Logf("Get Call Result return: \n\t%+v\n", str.String())
	}else{
		t.Logf("Get Call Result return: \n\t%+v\n", logCont)
	}
}

func TestGetCall(t *testing.T) {
	tw := NewClient(testNodeAPI, true, symbol)

	if r, err := tw.GetCall("/api/scan/blocks?row=1&page=1" ); err != nil {
		t.Errorf("Get Call Result failed: %v\n", err)
	} else {
		PrintJsonLog(t, r.String())
	}
}

func TestPostCall(t *testing.T) {
	tw := NewClient(testNodeAPI, true, symbol)

	body := map[string]interface{}{
		"address" : "xxxx",
	}

	if r, err := tw.PostCall("/api/scan/account", body); err != nil {
		t.Errorf("Post Call Result failed: %v\n", err)
	} else {
		PrintJsonLog(t, r.String())
	}
}

func Test_getBlockHeight(t *testing.T) {

	c := NewClient(testNodeAPI, true, symbol)

	r, err := c.getBlockHeight()

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("height:", r)
	}
}

func Test_getBalance(t *testing.T) {

	c := NewClient(testNodeAPI, true, symbol)

	address := "xxxx"

	r, err := c.getBalance(address, "")

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(r)
	}

	address = "xxxx"

	r, err = c.getBalance(address, "")

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(r)
	}
}

func Test_getBlockByHeight(t *testing.T) {
	c := NewClient(testNodeAPI, true, symbol)
	r, err := c.getBlockByHeight(4286826)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(r)
	}
}

func Test_getMostHeightBlock(t *testing.T) {
	c := NewClient(testNodeAPI, true, symbol)
	r, err := c.getMostHeightBlock()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(r)
	}
}