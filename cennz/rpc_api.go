/*
 * Copyright 2018 The openwallet Authors
 * This file is part of the openwallet library.
 *
 * The openwallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The openwallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */
package cennz

import (
	"errors"
	"fmt"
	"github.com/blocktree/openwallet/v2/log"
	"github.com/imroc/req"
	"github.com/tidwall/gjson"
	"time"
)

type RpcClient struct {
	BaseURL string
	Debug   bool
}

func NewRpcClient(url string, debug bool, symbol string) *RpcClient {
	c := RpcClient{
		BaseURL: url,
		Debug: debug,
	}

	log.Debug("BaseURL : ", url)

	return &c
}

func (c *RpcClient) Call(method string, params []interface{}) (*gjson.Result, error) {
	authHeader := req.Header{
		"Accept":       "application/json",
		"Content-Type": "application/json",
	}
	body := make(map[string]interface{}, 0)
	body["jsonrpc"] = "2.0"
	body["id"] = 1
	body["method"] = method
	body["params"] = params

	if c.Debug {
		log.Debug("url : ", c.BaseURL, ", body : ", body)
	}

	r, err := req.Post(c.BaseURL, req.BodyJSON(&body), authHeader)

	if c.Debug {
		log.Debugf("%+v\n", r)
	}

	if err != nil {
		return nil, err
	}

	resp := gjson.ParseBytes(r.Bytes())
	err = isError(&resp)
	if err != nil {
		return nil, err
	}

	result := resp.Get("result")

	return &result, nil
}

//isError 是否报错
func isError(result *gjson.Result) error {
	var (
		err error
	)

	if !result.Get("error").IsObject() {

		if !result.Get("result").Exists() {
			return fmt.Errorf("Response is empty! ")
		}

		return nil
	}

	errInfo := fmt.Sprintf("[%d]%s",
		result.Get("error.code").Int(),
		result.Get("error.message").String()+" - "+result.Get("error.data").String())
	err = errors.New(errInfo)

	return err
}

func (c *RpcClient) GetRuntimeVersion() (*RuntimeVersion, error) {
	method := "state_getRuntimeVersion"

	params := []interface{}{
	}

	resp, err := c.Call(method, params)
	if err != nil {
		return nil, err
	}

	result, err := GetRuntimeVersion(resp)
	if err!=nil {
		return nil, err
	}

	return result, nil
}

func (c *RpcClient) GetGenesisHash() (string, error) {
	method := "chain_getBlockHash"

	params := []interface{}{
		0,
	}

	resp, err := c.Call(method, params)
	if err != nil {
		return "", err
	}

	return resp.String(), nil
}

func (c *RpcClient) sendTransaction(rawTx string) (string, error) {
	method := "author_submitExtrinsic"

	params := []interface{}{
		rawTx,
	}

	resp, err := c.Call(method, params)
	if err != nil {
		return "", err
	}

	time.Sleep(time.Duration(1) * time.Second)

	log.Debug("sendTransaction result : ", resp)

	if resp.Get("error").String() != "" && resp.Get("cause").String() != "" {
		return "", errors.New("Submit transaction with error: " + resp.Get("error").String() + "," + resp.Get("cause").String())
	}

	return resp.String(), nil
}

// 获取当前最高区块
func (c *RpcClient) GetBlockHash(height uint64) (string, error) {
	method := "chain_getBlockHash"

	params := []interface{}{
		height,
	}

	resp, err := c.Call(method, params)
	if err != nil {
		return "", err
	}

	return resp.String(), nil
}
