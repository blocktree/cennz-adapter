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
	"math/big"
	"path/filepath"

	"github.com/blocktree/openwallet/v2/common"
	"github.com/blocktree/openwallet/v2/hdkeystore"
	"github.com/blocktree/openwallet/v2/log"
	"github.com/blocktree/openwallet/v2/openwallet"
)

type WalletManager struct {
	openwallet.AssetsAdapterBase

	Storage   *hdkeystore.HDKeystore //秘钥存取
	ApiClient *ApiClient

	Config          *WalletConfig                 //钱包管理配置
	WalletsInSum    map[string]*openwallet.Wallet //参与汇总的钱包
	Blockscanner    *CENNZBlockScanner              //区块扫描器
	Decoder         openwallet.AddressDecoderV2   //地址编码器
	TxDecoder       openwallet.TransactionDecoder //交易单编码器
	Log             *log.OWLogger                 //日志工具
	ContractDecoder *ContractDecoder              //智能合约解析器
}

func NewWalletManager() *WalletManager {
	wm := WalletManager{}
	wm.Config = NewConfig(Symbol, MasterKey, AddrPrefix)
	storage := hdkeystore.NewHDKeystore(wm.Config.keyDir, hdkeystore.StandardScryptN, hdkeystore.StandardScryptP)
	wm.Storage = storage
	//参与汇总的钱包
	wm.WalletsInSum = make(map[string]*openwallet.Wallet)
	//区块扫描器
	wm.Blockscanner = NewCENNZBlockScanner(&wm)
	wm.Decoder = NewAddressDecoderV2(&wm)
	wm.TxDecoder = NewTransactionDecoder(&wm)
	wm.Log = log.NewOWLogger(wm.Symbol())
	wm.ContractDecoder = NewContractDecoder(&wm)

	//	wm.RPCClient = NewRpcClient("http://localhost:20336/")
	return &wm
}

//GetWalletInfo 获取钱包列表
func (wm *WalletManager) GetWalletInfo(walletID string) (*openwallet.Wallet, error) {

	wallets, err := wm.GetWallets()
	if err != nil {
		return nil, err
	}

	//获取钱包余额
	for _, w := range wallets {
		if w.WalletID == walletID {
			return w, nil
		}

	}

	return nil, errors.New("The wallet that your given name is not exist!")
}

//GetWallets 获取钱包列表
func (wm *WalletManager) GetWallets() ([]*openwallet.Wallet, error) {

	wallets, err := openwallet.GetWalletsByKeyDir(wm.Config.keyDir)
	if err != nil {
		return nil, err
	}

	for _, w := range wallets {
		w.DBFile = filepath.Join(wm.Config.dbPath, w.FileName()+".db")
	}

	return wallets, nil
}

func (wm *WalletManager) InitTokenMap() (error){
	result := make(map[string]openwallet.SmartContract)
	
	token := openwallet.SmartContract{}
	
	token = openwallet.SmartContract{
		ContractID:"",
		Address:"1",
		Symbol:"CENNZ",
		Name:"CENNZ",
		Token:"CENNZ",
		Decimals:4,
	}
	result[token.Address] = token

	token = openwallet.SmartContract{
		ContractID:"",
		Address:"2",
		Symbol:"CPAY",
		Name:"CPAY",
		Token:"CPAY",
		Decimals:4,
	}
	result[token.Address] = token

	feeToken = token
	tokenMap = result

	return nil
}

func (wm *WalletManager) GetTokenInMap(assetId string) (openwallet.SmartContract, bool){
	if len(tokenMap)==0 {
		wm.InitTokenMap()
	}
	result, found := tokenMap[assetId]

	return result, found
}

func (wm *WalletManager) GetFeeToken() (openwallet.SmartContract){
	if len(tokenMap)==0 {
		wm.InitTokenMap()
	}
	return feeToken
}

//SendRawTransaction 广播交易
func (wm *WalletManager) SendRawTransaction(txHex string) (string, error) {

	return wm.sendRawTransactionByNode(txHex)
}

func (wm *WalletManager) sendRawTransactionByNode(txHex string) (string, error) {
	txid, err := wm.ApiClient.sendTransaction(txHex)

	if err != nil {
		return "", err
	}
	return txid, nil
}

func (wm *WalletManager) GetTransactionFeeEstimated(from string, to string, value *big.Int, assetId string) (*txFeeInfo, error) {

	var (
		gasLimit *big.Int
		gasPrice *big.Int
		//err      error
	)

	feeInfo := &txFeeInfo{
		GasLimit: gasLimit,
		GasPrice: gasPrice,
		//		Fee:      fee,
	}

	feeInfo.Fee = big.NewInt( wm.Config.FixedFee )
	return feeInfo, nil
}

// GetAddressNonce
func (wm *WalletManager) GetAddressNonce(wrapper openwallet.WalletDAI, address string) (uint64, error) {
	var (
		key           = wm.Symbol() + "-nonce"
		nonce         uint64
		nonce_db      interface{}
		nonce_onchain uint64
	)

	apiAddrBalance, err := wm.ApiClient.getBalance(address, "")
	if err!=nil {
		return 0, errors.New(address+" get address nonce error : " + err.Error() )
	}

	//获取db记录的nonce并确认nonce值
	nonce_db, _ = wrapper.GetAddressExtParam(address, key)

	//判断nonce_db是否为空,为空则说明当前nonce是0
	if nonce_db == nil {
		nonce = 0
	} else {
		nonce = common.NewString(nonce_db).UInt64()
	}

	nonce_onchain = apiAddrBalance.Nonce

	wm.Log.Info(address, " get nonce : ", nonce, ", nonce_onchain : ", nonce_onchain)

	//如果本地nonce_db > 链上nonce,采用本地nonce,否则采用链上nonce
	if nonce > nonce_onchain {
		//wm.Log.Debugf("%s nonce_db=%v > nonce_chain=%v,Use nonce_db...", address, nonce_db, nonce_onchain)
	} else {
		nonce = nonce_onchain
		//wm.Log.Debugf("%s nonce_db=%v <= nonce_chain=%v,Use nonce_chain...", address, nonce_db, nonce_onchain)
	}

	////临时
	//if nonce > 430 && nonce < 450 {
	//	nonce = nonce_onchain
	//}

	return nonce, nil
}

// UpdateAddressNonce
func (wm *WalletManager) UpdateAddressNonce(wrapper openwallet.WalletDAI, address string, nonce uint64) {
	key := wm.Symbol() + "-nonce"
	wm.Log.Info(address, " set nonce ", nonce)
	err := wrapper.SetAddressExtParam(address, key, nonce)
	if err != nil {
		wm.Log.Errorf("WalletDAI SetAddressExtParam failed, err: %v", err)
	}
}
