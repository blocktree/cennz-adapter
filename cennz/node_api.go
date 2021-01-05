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
	"github.com/blocktree/openwallet/v2/log"
	"github.com/imroc/req"
	"github.com/tidwall/gjson"
	"math/big"
	"strconv"
)

type ClientInterface interface {
	Call(path string, request []interface{}) (*gjson.Result, error)
}

// A Client is a Elastos RPC client. It performs RPCs over HTTP using JSON
// request and responses. A Client must be configured with a secret token
// to authenticate with other Cores on the network.
type Client struct {
	BaseURL     string
	AccessToken string
	Debug       bool
	client      *req.Req
	Symbol      string
}

type Response struct {
	Code    int         `json:"code,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Message string      `json:"message,omitempty"`
	Id      string      `json:"id,omitempty"`
}

func NewClient(url string /*token string,*/, debug bool, symbol string) *Client {
	c := Client{
		BaseURL: url,
		//	AccessToken: token,
		Debug: debug,
	}

	log.Debug("BaseURL : ", url)

	api := req.New()
	//trans, _ := api.Client().Transport.(*http.Transport)
	//trans.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	c.client = api
	c.Symbol = symbol

	return &c
}

// 用get方法获取内容
func (c *Client) PostCall(path string, v map[string]interface{}) (*gjson.Result, error) {
	if c.Debug {
		log.Debug("Start Request API...")
	}

	r, err := req.Post(c.BaseURL+path, req.BodyJSON(&v))

	if c.Debug {
		log.Std.Info("Request API Completed")
	}

	if c.Debug {
		log.Debugf("%+v\n", r)
	}

	if err != nil {
		return nil, err
	}

	resp := gjson.ParseBytes(r.Bytes())

	result := resp

	return &result, nil
}

// 用get方法获取内容
func (c *Client) GetCall(path string) (*gjson.Result, error) {

	if c.Debug {
		log.Debug("Start Request API...")
	}

	r, err := req.Get(c.BaseURL + path)

	if c.Debug {
		log.Std.Info("Request API Completed")
	}

	if c.Debug {
		log.Debugf("%+v\n", r)
	}

	if err != nil {
		return nil, err
	}

	resp := gjson.ParseBytes(r.Bytes())

	result := resp

	return &result, nil
}

// 获取当前最高区块
func (c *Client) getMetaData() (*Metadata, error) {
	resp, err := c.GetCall("/api/scan/metadata")

	if err != nil {
		return nil, err
	}

	metadata, err := GetMetadata(resp)
	if err!=nil {
		return nil, err
	}

	return metadata, nil
}

// 获取当前最高区块
func (c *Client) getBlockHeight() (uint64, error) {
	metadata, err := c.getMetaData()
	if err != nil {
		return 0, err
	}
	return metadata.BlockNum, nil
}

// 获取地址余额
func (c *Client) getBalance(address string, assetId string) (*AddrBalance, error) {
	body := map[string]interface{}{
		"address" : address,
	}

	r, err := c.PostCall("/api/scan/account", body);

	if err != nil {
		return nil, err
	}

	//{"statusCode":404,"message":"Not Found"}
	statusCode := gjson.Get(r.Raw, "statusCode").Int()
	message := gjson.Get(r.Raw, "message").String()
	if statusCode==404 && message=="Not Found" {
		result := AddrBalance{
			Address: address,
			AssetId: assetId,
			Balance: big.NewInt(0),
			Free:    big.NewInt(0),
			Freeze:  big.NewInt(0),
			Nonce:   0,
			index:   0,
			Actived: false,
		}
		return &result, nil
	}

	dataJSON, err := GetApiData(r)
	if err!=nil {
		return nil, err
	}

	if assetId=="" {
		assetId = "1"
	}

	addrBalance := AddrBalance{
		Address:address,
		Actived:true,
	}

	if gjson.Get(dataJSON.Raw, "nonce").Exists()==false {
		return nil, errors.New("nonce not found")
	}
	addrBalance.Nonce = gjson.Get(dataJSON.Raw, "nonce").Uint()

	if gjson.Get(dataJSON.Raw, "balances").Exists()==false {
		return nil, errors.New("balances not found")
	}
	balances := gjson.Get(dataJSON.Raw, "balances").Array()

	findBalance := false
	for _, balance := range balances {
		if gjson.Get(balance.Raw, "assetId").Exists()==false {
			continue
		}
		if gjson.Get(balance.Raw, "free").Exists()==false {
			continue
		}
		if gjson.Get(balance.Raw, "lock").Exists()==false {
			continue
		}

		itemAssetId := gjson.Get(balance.Raw, "assetId").String()
		free := big.NewInt(balance.Get("free").Int())
		lock := big.NewInt(balance.Get("lock").Int())

		if assetId==itemAssetId{
			addrBalance.AssetId = assetId
			addrBalance.Free = free
			addrBalance.Freeze = lock
			balanceBigInt := new(big.Int)
			addrBalance.Balance = balanceBigInt.Sub(free, lock)

			findBalance = true
			break
		}
	}
	if findBalance==false {
		return nil, errors.New("wrong balances")
	}

	return &addrBalance, nil
}

func (c *Client) getBlockByHeight(height uint64) (*Block, error) {
	resp, err := c.GetCall("/api/scan/block?block_num=" + strconv.FormatUint(height, 10))

	if err != nil {
		return nil, err
	}

	dataJSON, err := GetApiData(resp)
	if err!=nil {
		return nil, err
	}

	return NewBlock(dataJSON, c.Symbol), nil
}

//获取当前最新高度
func (c *Client) getMostHeightBlock() (*Block, error) {
	resp, err := c.GetCall("/api/scan/blocks?row=1&page=1")

	if err != nil {
		return nil, err
	}

	dataJSON, err := GetApiData(resp)
	if err!=nil {
		return nil, err
	}

	if gjson.Get(dataJSON.Raw, "blocks").Exists()==false {
		return nil, errors.New("blocks not found")
	}
	blocks := gjson.Get(dataJSON.Raw, "blocks").Array()

	if len(blocks)<1 {
		return nil, errors.New("blocks length not right")
	}

	return NewBlock(&blocks[0], c.Symbol), nil
}
