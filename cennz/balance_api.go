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

// A Client is a Elastos RPC client. It performs RPCs over HTTP using JSON
// request and responses. A Client must be configured with a secret token
// to authenticate with other Cores on the network.
type BalanceApiClient struct {
	BaseURL     string
	AccessToken string
	Debug       bool
	client      *req.Req
	Symbol      string
}

func NewBalanceClient(url string /*token string,*/, debug bool, symbol string) *BalanceApiClient {
	c := BalanceApiClient{
		BaseURL: url,
		//	AccessToken: token,
		Debug: debug,
	}

	log.Debug("Balance BaseURL : ", url)

	api := req.New()
	//trans, _ := api.Client().Transport.(*http.Transport)
	//trans.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	c.client = api
	c.Symbol = symbol

	return &c
}

// 用get方法获取内容
func (c *BalanceApiClient) BalanceApiGetCall(path string) (*gjson.Result, error) {

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

// 获取地址余额
func (c *BalanceApiClient) getApiBalance(addrBalance *AddrBalance, blockHash string) (*AddrBalance, error) {
	url := "/account/balance?address=" + addrBalance.Address + "&assetid=" + addrBalance.AssetId
	if len(blockHash)>0 {
		url += "&blockhash=" + blockHash
	}

	r, err := c.BalanceApiGetCall(url);

	if err != nil {
		return nil, err
	}

	//{"message":"Invalid request"}
	//{"balance":"190311967221"}
	message := gjson.Get(r.Raw, "message").String()
	balanceStr := gjson.Get(r.Raw, "balance").String()

	if len(message) > 0 {
		return nil, errors.New(message)
	}
	freeBalance, ok := new(big.Int).SetString(balanceStr, 10)
	if !ok {
		return nil, errors.New("wrong balance " + balanceStr)
	}

	addrBalance.Free = freeBalance
	addrBalance.Freeze = big.NewInt(0)
	addrBalance.Balance = freeBalance

	return addrBalance, nil
}

// 获取地址余额，加上nonce
func (c *BalanceApiClient) getApiBalanceWithNonce(address, blockHash, assetId string) (*AddrBalance, error) {
	if assetId=="" {
		assetId = "1"
	}

	addrBalance := &AddrBalance{
		Address: address,
		AssetId: assetId,
		Balance: big.NewInt(0),
		Free:    big.NewInt(0),
		Freeze:  big.NewInt(0),
		Nonce:   0,
		index:   0,
		Actived: false,
	}

	url := "/account/balance?address=" + addrBalance.Address + "&assetid=" + addrBalance.AssetId
	if len(blockHash)>0 {
		url += "&blockhash=" + blockHash
	}

	r, err := c.BalanceApiGetCall(url);

	if err != nil {
		return nil, err
	}

	//{"message":"Invalid request"}
	//{"balance":"190311967221"}
	message := gjson.Get(r.Raw, "message").String()
	balanceStr := gjson.Get(r.Raw, "balance").String()
	nonceStr := gjson.Get(r.Raw, "nonce").String()

	if len(message) > 0 {
		return nil, errors.New(message)
	}
	freeBalance, ok := new(big.Int).SetString(balanceStr, 10)
	if !ok {
		return nil, errors.New("wrong balance " + balanceStr)
	}
	nonce, err := strconv.ParseUint(nonceStr, 10, 64)
	if err!=nil {
		return nil, errors.New("wrong nonce " + nonceStr + ", error : " + err.Error() )
	}

	addrBalance.Free = freeBalance
	addrBalance.Freeze = big.NewInt(0)
	addrBalance.Balance = freeBalance
	addrBalance.Nonce = nonce

	return addrBalance, nil
}

// 获取地址余额
func (c *BalanceApiClient) getBlockByHeight(height uint64) (*Block, error) {
	url := "/block/getblock?height=" + strconv.FormatUint(height, 10)

	r, err := c.BalanceApiGetCall(url);

	if err != nil {
		return nil, err
	}

	return NewBlockFromRpc(r, c.Symbol)
}