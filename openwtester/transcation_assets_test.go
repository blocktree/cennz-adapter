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

package openwtester

import (
	"strconv"
	"testing"
	"time"

	"github.com/blocktree/openwallet/v2/log"
	"github.com/blocktree/openwallet/v2/openw"
	"github.com/blocktree/openwallet/v2/openwallet"
)

func testGetAssetsAccountBalance(tm *openw.WalletManager, walletID, accountID string) {
	balance, err := tm.GetAssetsAccountBalance(testApp, walletID, accountID)
	if err != nil {
		log.Error("GetAssetsAccountBalance failed, unexpected error:", err)
		return
	}
	log.Info("balance:", balance)
}

func testGetAssetsAccountTokenBalance(tm *openw.WalletManager, walletID, accountID string, contract openwallet.SmartContract) {
	balance, err := tm.GetAssetsAccountTokenBalance(testApp, walletID, accountID, contract)
	if err != nil {
		log.Error("GetAssetsAccountTokenBalance failed, unexpected error:", err)
		return
	}
	log.Info("token balance:", balance.Balance)
}

func testCreateTransactionStep(tm *openw.WalletManager, walletID, accountID, to, amount, feeRate string, contract *openwallet.SmartContract) (*openwallet.RawTransaction, error) {

	//err := tm.RefreshAssetsAccountBalance(testApp, accountID)
	//if err != nil {
	//	log.Error("RefreshAssetsAccountBalance failed, unexpected error:", err)
	//	return nil, err
	//}

	rawTx, err := tm.CreateTransaction(testApp, walletID, accountID, amount, to, feeRate, "test", contract)

	if err != nil {
		log.Error("CreateTransaction failed, unexpected error:", err)
		return nil, err
	}

	return rawTx, nil
}

func testCreateSummaryTransactionStep(
	tm *openw.WalletManager,
	walletID, accountID, summaryAddress, minTransfer, retainedBalance, feeRate string,
	start, limit int,
	contract *openwallet.SmartContract,
	feeSupportAccount *openwallet.FeesSupportAccount) ([]*openwallet.RawTransactionWithError, error) {

	rawTxArray, err := tm.CreateSummaryRawTransactionWithError(testApp, walletID, accountID, summaryAddress, minTransfer,
		retainedBalance, feeRate, start, limit, contract, feeSupportAccount)

	if err != nil {
		log.Error("CreateSummaryTransaction failed, unexpected error:", err)
		return nil, err
	}

	return rawTxArray, nil
}

func testSignTransactionStep(tm *openw.WalletManager, rawTx *openwallet.RawTransaction) (*openwallet.RawTransaction, error) {

	log.Info("wait sign message : ", rawTx.Signatures[rawTx.Account.AccountID][0].Message)

	_, err := tm.SignTransaction(testApp, rawTx.Account.WalletID, rawTx.Account.AccountID, "12345678", rawTx)
	if err != nil {
		log.Error("SignTransaction failed, unexpected error:", err)
		return nil, err
	}

	log.Infof("rawTx: %+v", rawTx)
	return rawTx, nil
}

func testVerifyTransactionStep(tm *openw.WalletManager, rawTx *openwallet.RawTransaction) (*openwallet.RawTransaction, error) {

	//log.Info("rawTx.Signatures:", rawTx.Signatures)

	_, err := tm.VerifyTransaction(testApp, rawTx.Account.WalletID, rawTx.Account.AccountID, rawTx)
	if err != nil {
		log.Error("VerifyTransaction failed, unexpected error:", err)
		return nil, err
	}

	log.Infof("rawTx: %+v", rawTx)
	return rawTx, nil
}

func testSubmitTransactionStep(tm *openw.WalletManager, rawTx *openwallet.RawTransaction) (*openwallet.RawTransaction, error) {

	tx, err := tm.SubmitTransaction(testApp, rawTx.Account.WalletID, rawTx.Account.AccountID, rawTx)
	if err != nil {
		log.Error("SubmitTransaction failed, unexpected error:", err)
		return nil, err
	}

	log.Std.Info("tx: %+v", tx)
	log.Info("wxID:", tx.WxID)
	log.Info("txID:", rawTx.TxID)

	return rawTx, nil
}

func ClearAddressNonce(tm *openw.WalletManager, walletID string, accountID string) error{
	wrapper, err := tm.NewWalletWrapper(testApp, "")
	if err != nil {
		return err
	}

	list, err := tm.GetAddressList(testApp, walletID, accountID, 0, -1, false)
	if err != nil {
		log.Error("unexpected error:", err)
		return err
	}
	for i, w := range list {
		log.Info("address[", i, "] :", w.Address)

		key := "CENNZ-nonce"
		wrapper.SetAddressExtParam(w.Address, key, 0)
	}
	log.Info("address count:", len(list))

	tm.CloseDB(testApp)

	return nil
}

/*
withdraw
wallet : W7uWEkkJ4g85ixibXbe44cr5rkcxB3V8xu
account : H7mDVFKgEJQm9okqVAivVjhffGNhsdEhLU3AcuooQZKM
1 address : 5EuiJzHg1uGqPN5Rf28tdNSRzC8G6RoEJFHb44Z9rnkaze6d

charge
wallet : W48Gx98MXbU938AL7iYfbeWpRUPdTjZJkC
account : 4Q3BDMDEWbcxDoSkmN7FtGhJ833ckz7rSo2PLran1m88
1 address : 5Cg1R6qXijprZJ1z8wSqabRSAGc2z2Za9XdWD7oZaLVM1jk3
2 address : 5CjFEnKJ7KJoBwcTLecvNqMQLEHWPrFp7kA7yLeSUvUNJjKH
3 address : 5CsCkhuNtRC5Eb8p5yzBwdjpNGBhBNPXge2SRo952ep7H85u
4 address : 5Dg7qNp9sNuUNK69qy6FxJjvunSXMTjJJKXD29WKej2CciDr
5 address : 5EAEHJX44RmZ7zNqHzhHsHAMtLFnwKpKbL2dQCyaEERNH6X4
6 address : 5EWi3fJfgbFzeB5dfqBFki1dBs3MrbZrY7NRGbTakJ6BGDMX
7 address : 5EzrHFrA7ToWy7jjin76XZv8og2Hz9S4pUSGok6BBpCBRVdu
8 address : 5GmrNXhbveHrwkPEgah8R5yyjE8Pq4kD9KqkY8K4bV9zNYhH
9 address : 5H42ShpjPdEfC2JtjNpDgRdbn5YbzSqm9iKQLUMciq1vZhFz
*/
func TestTransfer(t *testing.T) {
	tm := testInitWalletManager()
	walletID := "W7uWEkkJ4g85ixibXbe44cr5rkcxB3V8xu"
	accountID := "H7mDVFKgEJQm9okqVAivVjhffGNhsdEhLU3AcuooQZKM"
	to := "5CpbevWzpUVcmikdj2fApxd3Aip6cRu2Lc3cjRKKwFbT1wVh"

	//contract := openwallet.SmartContract{
	//	ContractID:"",
	//	Address:"1",
	//	Symbol:"CENNZ",
	//	Name:"CENNZ",
	//	Token:"CENNZ",
	//	Decimals:4,
	//}
	//
	//testGetAssetsAccountTokenBalance(tm, walletID, accountID, contract)
	//
	//rawTx, err := testCreateTransactionStep(tm, walletID, accountID, to, "0.1365", "", &contract)
	//if err != nil {
	//	return
	//}

	feeContract := openwallet.SmartContract{
		ContractID:"",
		Address:"2",
		Symbol:"CPAY",
		Name:"CPAY",
		Token:"CPAY",
		Decimals:4,
	}

	testGetAssetsAccountTokenBalance(tm, walletID, accountID, feeContract)

	rawTx, err := testCreateTransactionStep(tm, walletID, accountID, to, "15", "", &feeContract)
	if err != nil {
		return
	}

	log.Std.Info("rawTx: %+v", rawTx)

	_, err = testSignTransactionStep(tm, rawTx)
	if err != nil {
		return
	}

	_, err = testVerifyTransactionStep(tm, rawTx)
	if err != nil {
		return
	}

	_, err = testSubmitTransactionStep(tm, rawTx)
	if err != nil {
		return
	}
}

/*
withdraw
wallet : W7uWEkkJ4g85ixibXbe44cr5rkcxB3V8xu
account : H7mDVFKgEJQm9okqVAivVjhffGNhsdEhLU3AcuooQZKM
1 address : 5EuiJzHg1uGqPN5Rf28tdNSRzC8G6RoEJFHb44Z9rnkaze6d

charge
wallet : W48Gx98MXbU938AL7iYfbeWpRUPdTjZJkC
account : 4Q3BDMDEWbcxDoSkmN7FtGhJ833ckz7rSo2PLran1m88
1 address : 5Cg1R6qXijprZJ1z8wSqabRSAGc2z2Za9XdWD7oZaLVM1jk3
2 address : 5CjFEnKJ7KJoBwcTLecvNqMQLEHWPrFp7kA7yLeSUvUNJjKH
3 address : 5CsCkhuNtRC5Eb8p5yzBwdjpNGBhBNPXge2SRo952ep7H85u
4 address : 5Dg7qNp9sNuUNK69qy6FxJjvunSXMTjJJKXD29WKej2CciDr
5 address : 5EAEHJX44RmZ7zNqHzhHsHAMtLFnwKpKbL2dQCyaEERNH6X4
6 address : 5EWi3fJfgbFzeB5dfqBFki1dBs3MrbZrY7NRGbTakJ6BGDMX
7 address : 5EzrHFrA7ToWy7jjin76XZv8og2Hz9S4pUSGok6BBpCBRVdu
8 address : 5GmrNXhbveHrwkPEgah8R5yyjE8Pq4kD9KqkY8K4bV9zNYhH
9 address : 5H42ShpjPdEfC2JtjNpDgRdbn5YbzSqm9iKQLUMciq1vZhFz
*/
func TestBatchTransfer(t *testing.T) {
	tm := testInitWalletManager()
	walletID := "W7uWEkkJ4g85ixibXbe44cr5rkcxB3V8xu"
	accountID := "H7mDVFKgEJQm9okqVAivVjhffGNhsdEhLU3AcuooQZKM"
	toArr := make([]string, 0)
	toArr = append(toArr, "5Cg1R6qXijprZJ1z8wSqabRSAGc2z2Za9XdWD7oZaLVM1jk3")
	toArr = append(toArr, "5CjFEnKJ7KJoBwcTLecvNqMQLEHWPrFp7kA7yLeSUvUNJjKH")
	toArr = append(toArr, "5CsCkhuNtRC5Eb8p5yzBwdjpNGBhBNPXge2SRo952ep7H85u")
	toArr = append(toArr, "5Dg7qNp9sNuUNK69qy6FxJjvunSXMTjJJKXD29WKej2CciDr")

	//contract := openwallet.SmartContract{
	//	ContractID:"",
	//	Address:"1",
	//	Symbol:"CENNZ",
	//	Name:"CENNZ",
	//	Token:"CENNZ",
	//	Decimals:4,
	//}
	contract := openwallet.SmartContract{
		ContractID:"",
		Address:"2",
		Symbol:"CPAY",
		Name:"CPAY",
		Token:"CPAY",
		Decimals:4,
	}

	for i := 0; i < len(toArr); i++{
		to := toArr[i]
		testGetAssetsAccountBalance(tm, walletID, accountID)

		testGetAssetsAccountTokenBalance(tm, walletID, accountID, contract)

		rawTx, err := testCreateTransactionStep(tm, walletID, accountID, to, "1.631"+strconv.FormatInt(int64(i), 10), "", &contract)
		if err != nil {
			return
		}

		log.Std.Info("rawTx: %+v", rawTx)

		_, err = testSignTransactionStep(tm, rawTx)
		if err != nil {
			return
		}

		_, err = testVerifyTransactionStep(tm, rawTx)
		if err != nil {
			return
		}

		_, err = testSubmitTransactionStep(tm, rawTx)
		if err != nil {
			return
		}

		time.Sleep(time.Duration(5) * time.Second)
	}
}

/**
	feeSupport
	wallet : W7uWEkkJ4g85ixibXbe44cr5rkcxB3V8xu
	account : 48cZBtzDFbsTYK2PZ2HrTh799ZxAZW5U6VsbAyv2SWFm
	address : 5CpbevWzpUVcmikdj2fApxd3Aip6cRu2Lc3cjRKKwFbT1wVh

	charge
	wallet : W48Gx98MXbU938AL7iYfbeWpRUPdTjZJkC
	account : 4Q3BDMDEWbcxDoSkmN7FtGhJ833ckz7rSo2PLran1m88

	withdraw
	wallet : W7uWEkkJ4g85ixibXbe44cr5rkcxB3V8xu
	account : H7mDVFKgEJQm9okqVAivVjhffGNhsdEhLU3AcuooQZKM
	1 address : 5EuiJzHg1uGqPN5Rf28tdNSRzC8G6RoEJFHb44Z9rnkaze6d
 */
func TestSummary(t *testing.T) {
	tm := testInitWalletManager()

	walletID := "W48Gx98MXbU938AL7iYfbeWpRUPdTjZJkC"
	accountID := "4Q3BDMDEWbcxDoSkmN7FtGhJ833ckz7rSo2PLran1m88"
	summaryAddress := "5EuiJzHg1uGqPN5Rf28tdNSRzC8G6RoEJFHb44Z9rnkaze6d"

	ClearAddressNonce(tm, walletID, accountID)

	feesSupport := openwallet.FeesSupportAccount{
		AccountID: "48cZBtzDFbsTYK2PZ2HrTh799ZxAZW5U6VsbAyv2SWFm",
		//FixSupportAmount: "0.01",
		FeesSupportScale: "",
	}

	//contract := openwallet.SmartContract{
	//	Address:  "1",
	//	Symbol:   "CENNZ",
	//	Name:     "CENNZ",
	//	Token:    "CENNZ",
	//	Decimals: 4,
	//}
	contract := openwallet.SmartContract{
		Address:  "2",
		Symbol:   "CPAY",
		Name:     "CPAY",
		Token:    "CPAY",
		Decimals: 4,
	}

	testGetAssetsAccountBalance(tm, walletID, accountID)

	testGetAssetsAccountTokenBalance(tm, walletID, accountID, contract)

	rawTxArray, err := testCreateSummaryTransactionStep(tm, walletID, accountID,
		summaryAddress, "1.51", "1.5", "",
		0, 100, &contract, &feesSupport)
	if err != nil {
		log.Errorf("CreateSummaryTransaction failed, unexpected error: %v", err)
		return
	}

	//执行汇总交易
	for _, rawTxWithErr := range rawTxArray {

		if rawTxWithErr.Error != nil {
			log.Error(rawTxWithErr.Error.Error())
			continue
		}

		_, err = testSignTransactionStep(tm, rawTxWithErr.RawTx)
		if err != nil {
			return
		}

		_, err = testVerifyTransactionStep(tm, rawTxWithErr.RawTx)
		if err != nil {
			return
		}

		_, err = testSubmitTransactionStep(tm, rawTxWithErr.RawTx)
		if err != nil {
			return
		}
	}

}
