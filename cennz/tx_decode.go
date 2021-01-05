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
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/blocktree/cennz-adapter/cennzTransaction"
	"github.com/blocktree/openwallet/v2/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/shopspring/decimal"
	"math/big"
	"sort"
	"strconv"
	"time"

	"github.com/blocktree/openwallet/v2/openwallet"
	"github.com/prometheus/common/log"

	"github.com/blocktree/go-owcdrivers/polkadotTransaction"
)

type TransactionDecoder struct {
	openwallet.TransactionDecoderBase
	openwallet.AddressDecoderV2
	wm *WalletManager //钱包管理者
}

//NewTransactionDecoder 交易单解析器
func NewTransactionDecoder(wm *WalletManager) *TransactionDecoder {
	decoder := TransactionDecoder{}
	decoder.wm = wm
	return &decoder
}

//CreateRawTransaction 创建交易单
func (decoder *TransactionDecoder) CreateRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	if rawTx.Coin.IsContract {
		return decoder.CreateCENNZRawTransaction(wrapper, rawTx)
	}
	return openwallet.Errorf(openwallet.ErrCreateRawTransactionFailed, "[%s] Miss contract details to create transaction!", rawTx.Account.AccountID)
}

//SignRawTransaction 签名交易单
func (decoder *TransactionDecoder) SignRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	return decoder.SignCennzRawTransaction(wrapper, rawTx)
}

//VerifyRawTransaction 验证交易单，验证交易单并返回加入签名后的交易单
func (decoder *TransactionDecoder) VerifyRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	return decoder.VerifyCENNZRawTransaction(wrapper, rawTx)
}

func (decoder *TransactionDecoder) SubmitRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) (*openwallet.Transaction, error) {
	if len(rawTx.RawHex) == 0 {
		return nil, fmt.Errorf("transaction hex is empty")
	}

	if !rawTx.IsCompleted {
		return nil, fmt.Errorf("transaction is not completed validation")
	}

	from := rawTx.Signatures[rawTx.Account.AccountID][0].Address.Address
	nonce := rawTx.Signatures[rawTx.Account.AccountID][0].Nonce
	nonceUint, _ := strconv.ParseUint(nonce[2:], 16, 64)

	decoder.wm.Log.Info("nonce : ", nonceUint, " update from : ", from)

	txid, err := decoder.wm.SendRawTransaction(rawTx.RawHex)
	if err != nil {
		decoder.wm.UpdateAddressNonce(wrapper, from, 0)
		decoder.wm.Log.Error("Error Tx to send: ", rawTx.RawHex)
		return nil, err
	}

	//交易成功，地址nonce+1并记录到缓存
	newNonce, _ := math.SafeAdd(nonceUint, uint64(1)) //nonce+1
	decoder.wm.UpdateAddressNonce(wrapper, from, newNonce)

	rawTx.TxID = txid
	rawTx.IsSubmit = true

	decimals := int32(4)

	tx := openwallet.Transaction{
		From:       rawTx.TxFrom,
		To:         rawTx.TxTo,
		Amount:     rawTx.TxAmount,
		Coin:       rawTx.Coin,
		TxID:       rawTx.TxID,
		Decimal:    decimals,
		AccountID:  rawTx.Account.AccountID,
		Fees:       rawTx.Fees,
		SubmitTime: time.Now().Unix(),
	}

	tx.WxID = openwallet.GenTransactionWxID(&tx)

	return &tx, nil
}

func (decoder *TransactionDecoder) CreateCENNZRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	var (
		accountID       = rawTx.Account.AccountID
		findAddrBalance *CennzAddrBalance
		errBalance      string
		errTokenBalance string
		feeInfo *txFeeInfo
	)

	tokenDecimals := int32(rawTx.Coin.Contract.Decimals)
	contractAddress := rawTx.Coin.Contract.Address

	//获取wallet
	addresses, err := wrapper.GetAddressList(0, -1,
		"AccountID", accountID)
	if err != nil {
		return openwallet.NewError(openwallet.ErrAddressNotFound, err.Error())
	}

	if len(addresses) == 0 {
		return openwallet.Errorf(openwallet.ErrAccountNotAddress, "[%s] have not addresses", accountID)
	}

	searchAddrs := make([]string, 0)
	for _, address := range addresses {
		searchAddrs = append(searchAddrs, address.Address)
	}

	addrBalanceArray, err := decoder.wm.ContractDecoder.GetTokenBalanceByAddress(rawTx.Coin.Contract, searchAddrs...)
	if err != nil {
		return openwallet.NewError(openwallet.ErrCallFullNodeAPIFailed, err.Error())
	}

	var amountStr, to string
	for k, v := range rawTx.To {
		to = k
		amountStr = v
		break
	}

	//地址余额从大到小排序
	sort.Slice(addrBalanceArray, func(i int, j int) bool {
		a_amount, _ := decimal.NewFromString(addrBalanceArray[i].Balance.Balance)
		b_amount, _ := decimal.NewFromString(addrBalanceArray[j].Balance.Balance)
		if a_amount.LessThan(b_amount) {
			return true
		} else {
			return false
		}
	})

	tokenBalanceNotEnough := false
	feeNotEnough := false

	for _, addrBalance := range addrBalanceArray {

		//检查余额是否超过最低转账
		addrBalance_BI := common.StringNumToBigIntWithExp(addrBalance.Balance.ConfirmBalance, tokenDecimals)

		amount := common.StringNumToBigIntWithExp(amountStr, tokenDecimals)

		if addrBalance_BI.Cmp(amount) < 0 {
			errTokenBalance = fmt.Sprintf("the token balance of all addresses is not enough")
			tokenBalanceNotEnough = true
			continue
		}
		//计算手续费
		fee, createErr := decoder.wm.GetTransactionFeeEstimated(addrBalance.Balance.Address, to, amount, contractAddress)
		if createErr != nil {
			//decoder.wm.Log.Std.Error("GetTransactionFeeEstimated from[%v] -> to[%v] failed, err=%v", addrBalance.Balance.Address, to, createErr)
			return createErr
		}

		feeBalance, err := decoder.wm.ApiClient.getBalance(addrBalance.Balance.Address, decoder.wm.GetFeeToken().Address)
		if err != nil {
			continue
		}

		if feeBalance.Free.Cmp( fee.Fee ) < 0  {
			coinBalanceDec := common.BigIntToDecimals(feeBalance.Balance, int32(decoder.wm.GetFeeToken().Decimals) )
			errBalance = fmt.Sprintf("the [%s] balance: %s is not enough to call smart contract", decoder.wm.GetFeeToken().Symbol, coinBalanceDec.String())
			feeNotEnough = true
			continue
		}

		//只要找到一个合适使用的地址余额就停止遍历
		//findAddrBalance = &AddrBalance{Address: addrBalance.Balance.Address, Balance: addrBalance_BI, FeeBalance: feeBalance.Balance}
		findAddrBalance = &CennzAddrBalance{
			Address:     addrBalance.Balance.Address,
			Balance:     addrBalance_BI,
			FeeBalance: feeBalance,
		}
		feeInfo = fee
		break
	}

	if findAddrBalance==nil {
		if tokenBalanceNotEnough {
			return openwallet.Errorf(openwallet.ErrInsufficientTokenBalanceOfAddress, errTokenBalance)
		}
		if feeNotEnough {
			return openwallet.Errorf(openwallet.ErrInsufficientFees, errBalance)
		}
	}

	//最后创建交易单
	createTxErr := decoder.createRawTransaction(
		wrapper,
		rawTx,
		findAddrBalance,
		feeInfo,
		nil)
	if createTxErr != nil {
		return createTxErr
	}

	return nil
}

func (decoder *TransactionDecoder) SignCennzRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	key, err := wrapper.HDKey()
	if err != nil {
		return nil
	}

	keySignatures := rawTx.Signatures[rawTx.Account.AccountID]

	if keySignatures != nil {
		for _, keySignature := range keySignatures {

			childKey, err := key.DerivedKeyWithPath(keySignature.Address.HDPath, keySignature.EccType)
			keyBytes, err := childKey.GetPrivateKeyBytes()
			if err != nil {
				return err
			}

			//签名交易
			///////交易单哈希签名
			signature, err := polkadotTransaction.SignTransaction(keySignature.Message, keyBytes)
			if err != nil {
				return fmt.Errorf("transaction hash sign failed, unexpected error: %v", err)
			}
			keySignature.Signature = hex.EncodeToString(signature)
		}
	}

	rawTx.Signatures[rawTx.Account.AccountID] = keySignatures

	return nil
}

func (decoder *TransactionDecoder) VerifyCENNZRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	var (
		emptyTrans = rawTx.RawHex
		signature  = ""
	)

	for accountID, keySignatures := range rawTx.Signatures {
		log.Debug("accountID Signatures:", accountID)
		for _, keySignature := range keySignatures {

			signature = keySignature.Signature

			log.Debug("Signature:", keySignature.Signature)
			log.Debug("PublicKey:", keySignature.Address.PublicKey)
		}
	}

	signedTrans, pass := cennzTransaction.VerifyAndCombineTransaction(emptyTrans, signature)

	if pass {
		log.Debug("transaction verify passed")
		rawTx.IsCompleted = true
		rawTx.RawHex = signedTrans
	} else {
		log.Debug("transaction verify failed")
		rawTx.IsCompleted = false
	}

	return nil
}

func (decoder *TransactionDecoder) GetRawTransactionFeeRate() (feeRate string, unit string, err error) {
	rate := uint64(decoder.wm.Config.FixedFee)
	return convertToAmount(rate, 4), "TX", nil
}

//CreateSummaryRawTransaction 创建汇总交易，返回原始交易单数组
func (decoder *TransactionDecoder) CreateSummaryRawTransaction(wrapper openwallet.WalletDAI, sumRawTx *openwallet.SummaryRawTransaction) ([]*openwallet.RawTransaction, error) {
	var (
		rawTxWithErrArray []*openwallet.RawTransactionWithError
		rawTxArray        = make([]*openwallet.RawTransaction, 0)
		err               error
	)
	if sumRawTx.Coin.IsContract {
		rawTxWithErrArray, err = decoder.CreateTokenSummaryRawTransaction(wrapper, sumRawTx)
		if err != nil {
			return nil, err
		}

		for _, rawTxWithErr := range rawTxWithErrArray {
			if rawTxWithErr.Error != nil {
				continue
			}
			rawTxArray = append(rawTxArray, rawTxWithErr.RawTx)
		}
		return rawTxArray, nil
	} else {
		return nil, openwallet.Errorf(openwallet.ErrSubmitRawTransactionFailed, "[%s] Miss contract details to summary!", sumRawTx.Account.AccountID)
	}
}

func (decoder *TransactionDecoder) CreateTokenSummaryRawTransaction(wrapper openwallet.WalletDAI, sumRawTx *openwallet.SummaryRawTransaction) ([]*openwallet.RawTransactionWithError, error) {

	var (
		rawTxArray         = make([]*openwallet.RawTransactionWithError, 0)
		accountID          = sumRawTx.Account.AccountID
		minTransfer        *big.Int
		retainedBalance    *big.Int
		feesSupportAccount *openwallet.AssetsAccount
		tmpNonce           uint64
	)

	// 如果有提供手续费账户，检查账户是否存在
	if feesAcount := sumRawTx.FeesSupportAccount; feesAcount != nil {
		account, supportErr := wrapper.GetAssetsAccountInfo(feesAcount.AccountID)
		if supportErr != nil {
			return nil, openwallet.Errorf(openwallet.ErrAccountNotFound, "can not find fees support account")
		}

		feesSupportAccount = account

		//获取手续费支持账户的地址nonce
		feesAddresses, feesSupportErr := wrapper.GetAddressList(0, 1,
			"AccountID", feesSupportAccount.AccountID)
		if feesSupportErr != nil {
			return nil, openwallet.NewError(openwallet.ErrAddressNotFound, "fees support account have not addresses")
		}

		if len(feesAddresses) == 0 {
			return nil, openwallet.Errorf(openwallet.ErrAccountNotAddress, "fees support account have not addresses")
		}

		nonce, nonceErr := decoder.wm.GetAddressNonce(wrapper, feesAddresses[0].Address)
		if nonceErr!=nil {
			return nil, openwallet.Errorf(openwallet.ErrCreateRawTransactionFailed, nonceErr.Error())
		}
		tmpNonce = nonce
	}
	tokenDecimals := int32(sumRawTx.Coin.Contract.Decimals)
	contractAddress := sumRawTx.Coin.Contract.Address

	minTransfer = common.StringNumToBigIntWithExp(sumRawTx.MinTransfer, tokenDecimals)
	retainedBalance = common.StringNumToBigIntWithExp(sumRawTx.RetainedBalance, tokenDecimals)

	if minTransfer.Cmp(retainedBalance) < 0 {
		return nil, openwallet.Errorf(openwallet.ErrCreateRawTransactionFailed, "mini transfer amount must be greater than address retained balance")
	}

	//获取wallet
	addresses, err := wrapper.GetAddressList(sumRawTx.AddressStartIndex, sumRawTx.AddressLimit,
		"AccountID", sumRawTx.Account.AccountID)
	if err != nil {
		return nil, err
	}

	if len(addresses) == 0 {
		return nil, openwallet.Errorf(openwallet.ErrAccountNotAddress, "[%s] have not addresses", accountID)
	}

	searchAddrs := make([]string, 0)
	for _, address := range addresses {
		searchAddrs = append(searchAddrs, address.Address)
	}

	//查询Token余额
	addrBalanceArray, err := decoder.wm.ContractDecoder.GetTokenBalanceByAddress(sumRawTx.Coin.Contract, searchAddrs...)
	if err != nil {
		return nil, err
	}

	for _, addrBalance := range addrBalanceArray {

		//检查余额是否超过最低转账
		addrBalance_BI := common.StringNumToBigIntWithExp(addrBalance.Balance.Balance, tokenDecimals)

		if addrBalance_BI.Cmp(minTransfer) < 0 || addrBalance_BI.Cmp(big.NewInt(0)) <= 0 {
			continue
		}
		//计算汇总数量 = 余额 - 保留余额
		sumAmount_BI := new(big.Int)
		sumAmount_BI.Sub(addrBalance_BI, retainedBalance)

		////计算手续费
		feeInfo, createErr := decoder.wm.GetTransactionFeeEstimated(addrBalance.Balance.Address, contractAddress, sumAmount_BI, sumRawTx.Coin.Contract.Address)
		if createErr != nil {
			//decoder.wm.Log.Std.Error("GetTransactionFeeEstimated from[%v] -> to[%v] failed, err=%v", addrBalance.Balance.Address, sumRawTx.SummaryAddress, createErr)
			return nil, createErr
		}

		sumAmount := common.BigIntToDecimals(sumAmount_BI, tokenDecimals)
		fees := common.BigIntToDecimals(feeInfo.Fee, int32(decoder.wm.GetFeeToken().Decimals) )

		feeBalance, err := decoder.wm.ApiClient.getBalance(addrBalance.Balance.Address, decoder.wm.GetFeeToken().Address) //decoder.wm.GetAddrBalance(addrBalance.Balance.Address, "pending")
		if err != nil {
			continue
		}

		//判断主币余额是否够手续费
		if feeBalance.Balance.Cmp(feeInfo.Fee) < 0 {

			//有手续费账户支持
			if feesSupportAccount != nil {

				//通过手续费账户创建交易单
				supportAddress := addrBalance.Balance.Address
				supportAmount := decimal.Zero
				feesSupportScale, _ := decimal.NewFromString(sumRawTx.FeesSupportAccount.FeesSupportScale)
				fixSupportAmount, _ := decimal.NewFromString(sumRawTx.FeesSupportAccount.FixSupportAmount)

				//优先采用固定支持数量
				if fixSupportAmount.GreaterThan(decimal.Zero) {
					supportAmount = fixSupportAmount
				} else {
					//没有固定支持数量，有手续费倍率，计算支持数量
					if feesSupportScale.GreaterThan(decimal.Zero) {
						supportAmount = feesSupportScale.Mul(fees)
					} else {
						//默认支持数量为手续费
						supportAmount = fees

						//补充到，地址有足够的手续费就行了
						supportAmountBigInt := big.NewInt(0).Sub(feeInfo.Fee, feeBalance.Free )
						supportAmount = common.BigIntToDecimals(supportAmountBigInt, int32(decoder.wm.GetFeeToken().Decimals) )
					}
				}

				decoder.wm.Log.Debugf("create transaction for fees support account")
				decoder.wm.Log.Debugf("fees account: %s", feesSupportAccount.AccountID)
				decoder.wm.Log.Debugf("mini support amount: %s", fees.String())
				decoder.wm.Log.Debugf("allow support amount: %s", supportAmount.String())
				decoder.wm.Log.Debugf("support address: %s", supportAddress)

				supportCoin := openwallet.Coin{
					Symbol:     sumRawTx.Coin.Symbol,
					IsContract: true,
					Contract: feeToken,
				}

				//创建一笔交易单
				rawTx := &openwallet.RawTransaction{
					Coin:    supportCoin,
					Account: feesSupportAccount,
					To: map[string]string{
						addrBalance.Balance.Address: supportAmount.String(),
					},
					Required: 1,
				}

				createTxErr := decoder.CreateSimpleRawTransaction(wrapper, rawTx, &tmpNonce)
				rawTxWithErr := &openwallet.RawTransactionWithError{
					RawTx: rawTx,
					Error: openwallet.ConvertError(createTxErr),
				}

				//创建成功，添加到队列
				rawTxArray = append(rawTxArray, rawTxWithErr)

				//需要手续费支持的地址会有很多个，nonce要连续递增以保证交易广播生效
				tmpNonce++

				//汇总下一个
				continue
			}
		}

		decoder.wm.Log.Debugf("balance: %v", addrBalance.Balance.Balance)
		decoder.wm.Log.Debugf("%s fees: %v", sumRawTx.Coin.Symbol, fees)
		decoder.wm.Log.Debugf("sumAmount: %v", sumAmount)

		//创建一笔交易单
		rawTx := &openwallet.RawTransaction{
			Coin:    sumRawTx.Coin,
			Account: sumRawTx.Account,
			To: map[string]string{
				sumRawTx.SummaryAddress: sumAmount.StringFixed(int32(tokenDecimals)),
			},
			Required: 1,
		}

		findAddrBalance := CennzAddrBalance{
			Address: addrBalance.Balance.Address,
			Balance: addrBalance_BI,
			FeeBalance: feeBalance,
		}

		createTxErr := decoder.createRawTransaction(
			wrapper,
			rawTx,
			&findAddrBalance,
			feeInfo,
			nil)
		rawTxWithErr := &openwallet.RawTransactionWithError{
			RawTx: rawTx,
			Error: createTxErr,
		}

		//创建成功，添加到队列
		rawTxArray = append(rawTxArray, rawTxWithErr)

	}

	return rawTxArray, nil
}

func (decoder *TransactionDecoder) CreateSimpleRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction, tmpNonce *uint64) error {

	var (
		accountID       = rawTx.Account.AccountID
		findAddrBalance *CennzAddrBalance
		feeInfo          *txFeeInfo
		decimals        = int32( rawTx.Coin.Contract.Decimals )
	)

	//获取wallet
	addresses, err := wrapper.GetAddressList(0, -1,
		"AccountID", accountID)
	if err != nil {
		return openwallet.NewError(openwallet.ErrAddressNotFound, err.Error())
	}

	if len(addresses) == 0 {
		return openwallet.Errorf(openwallet.ErrAccountNotAddress, "[%s] have not addresses", accountID)
	}

	searchAddrs := make([]string, 0)
	for _, address := range addresses {
		searchAddrs = append(searchAddrs, address.Address)
	}

	addrBalanceArray, err := decoder.wm.ContractDecoder.GetTokenBalanceByAddress(rawTx.Coin.Contract, searchAddrs...) //decoder.wm.Blockscanner.GetBalanceByAddress(searchAddrs...)
	if err != nil {
		return openwallet.NewError(openwallet.ErrCallFullNodeAPIFailed, err.Error())
	}

	var amountStr, to string
	for k, v := range rawTx.To {
		to = k
		amountStr = v
		break
	}

	amount := common.StringNumToBigIntWithExp(amountStr, decimals)

	//地址余额从大到小排序
	sort.Slice(addrBalanceArray, func(i int, j int) bool {
		a_amount, _ := decimal.NewFromString(addrBalanceArray[i].Balance.Balance)
		b_amount, _ := decimal.NewFromString(addrBalanceArray[j].Balance.Balance)
		if a_amount.LessThan(b_amount) {
			return true
		} else {
			return false
		}
	})

	for _, addrBalance := range addrBalanceArray {

		//检查余额是否超过最低转账
		addrBalance_BI := common.StringNumToBigIntWithExp(addrBalance.Balance.Balance, decimals)

		////计算手续费
		feeInfo, err = decoder.wm.GetTransactionFeeEstimated(addrBalance.Balance.Address, to, amount, rawTx.Coin.Contract.Address)
		if err != nil {
			//decoder.wm.Log.Std.Error("GetTransactionFeeEstimated from[%v] -> to[%v] failed, err=%v", addrBalance.Address, to, err)
			continue
		}
		//
		//if rawTx.FeeRate != "" {
		//	feeInfo.GasPrice = common.StringNumToBigIntWithExp(rawTx.FeeRate, decoder.wm.Decimal())
		//	feeInfo.CalcFee()
		//}

		//总消耗数量 = 转账数量 + 手续费
		totalAmount := new(big.Int)
		totalAmount.Add(amount, feeInfo.Fee)

		if addrBalance_BI.Cmp(totalAmount) < 0 {
			continue
		}

		feeBalance, err := decoder.wm.ApiClient.getBalance(addrBalance.Balance.Address, decoder.wm.GetFeeToken().Address)
		if err != nil {
			//decoder.wm.Log.Std.Error("GetTransactionFeeEstimated from[%v] -> to[%v] failed, err=%v", addrBalance.Address, to, err)
			continue
		}

		//只要找到一个合适使用的地址余额就停止遍历
		//findAddrBalance = &AddrBalance{Address: addrBalance.Balance.Address, Balance: addrBalance_BI, FeeBalance: feeBalance.Balance}
		findAddrBalance = &CennzAddrBalance{
			Address:    addrBalance.Balance.Address,
			Balance:    addrBalance_BI,
			FeeBalance: feeBalance,
		}
		break
	}

	if findAddrBalance == nil {
		return openwallet.Errorf(openwallet.ErrInsufficientBalanceOfAccount, "the balance: %s is not enough", amountStr)
	}

	//最后创建交易单
	createTxErr := decoder.createRawTransaction(
		wrapper,
		rawTx,
		findAddrBalance,
		feeInfo,
		tmpNonce)
	if createTxErr != nil {
		return createTxErr
	}

	return nil
}

func (decoder *TransactionDecoder) createRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction, addrBalance *CennzAddrBalance, feeInfo *txFeeInfo, tmpNonce *uint64) *openwallet.Error {
	var (
		accountTotalSent = decimal.Zero
		txFrom           = make([]string, 0)
		txTo             = make([]string, 0)
		amountStr        string
		destination      string
	)

	tokenDecimals := int32(rawTx.Coin.Contract.Decimals)
	feeDecimals := int32(decoder.wm.GetFeeToken().Decimals)

	for k, v := range rawTx.To {
		destination = k
		amountStr = v
		break
	}

	//计算账户的实际转账amount
	accountTotalSentAddresses, findErr := wrapper.GetAddressList(0, -1, "AccountID", rawTx.Account.AccountID, "Address", destination)
	if findErr != nil || len(accountTotalSentAddresses) == 0 {
		amountDec, _ := decimal.NewFromString(amountStr)
		accountTotalSent = accountTotalSent.Add(amountDec)
	}

	txFrom = []string{fmt.Sprintf("%s:%s", addrBalance.Address, amountStr)}
	txTo = []string{fmt.Sprintf("%s:%s", destination, amountStr)}

	totalFeeDecimal := common.BigIntToDecimals(feeInfo.Fee, feeDecimals )

	feesDec, _ := decimal.NewFromString(rawTx.Fees)
	accountTotalSent = accountTotalSent.Add(feesDec)
	accountTotalSent = decimal.Zero.Sub(accountTotalSent)

	rawTx.FeeRate = strconv.FormatUint(feeInfo.Fee.Uint64(), 10)
	rawTx.Fees = totalFeeDecimal.String()
	//rawTx.ExtParam = string(extparastr)
	rawTx.TxAmount = accountTotalSent.String()
	rawTx.TxFrom = txFrom
	rawTx.TxTo = txTo

	addr, err := wrapper.GetAddress(addrBalance.Address)
	if err != nil {
		return openwallet.NewError(openwallet.ErrAccountNotAddress, err.Error())
	}

	var nonce uint64
	if tmpNonce == nil {
		//使用外部传入的扩展字段填充nonce
		useExtNonce := false
		if rawTx.GetExtParam().Exists() {
			if rawTx.GetExtParam().Get("nonce").Exists() {
				nonce = rawTx.GetExtParam().Get("nonce").Uint()
				useExtNonce = true
			}
		}
		if useExtNonce==false {
			txNonce, nonceErr := decoder.wm.GetAddressNonce(wrapper, addr.Address)
			if nonceErr!=nil {
				return openwallet.Errorf(openwallet.ErrCreateRawTransactionFailed, nonceErr.Error())
			}
			nonce = txNonce
		}
	} else {
		nonce = *tmpNonce
	}

	//构建合约交易
	amount := common.StringNumToBigIntWithExp(amountStr, tokenDecimals)
	if addrBalance.Balance.Cmp(amount) < 0 {
		return openwallet.Errorf(openwallet.ErrInsufficientTokenBalanceOfAddress, "the token balance: %s is not enough", amountStr)
		//return openwallet.Errorf("the token balance: %s is not enough", amountStr)
	}

	if addrBalance.FeeBalance.Free.Cmp( feeInfo.Fee ) < 0 {
		coinBalance := common.BigIntToDecimals(addrBalance.Balance, decoder.wm.Decimal())
		return openwallet.Errorf(openwallet.ErrInsufficientFees, "the [%s] balance: %s is not enough to call smart contract", rawTx.Coin.Symbol, coinBalance)
		//return openwallet.Errorf("the [%s] balance: %s is not enough to call smart contract", rawTx.Coin.Symbol, coinBalance)
	}

	nonceJSON := map[string]interface{}{}
	if len(rawTx.ExtParam) > 0 {
		err = json.Unmarshal([]byte(rawTx.ExtParam), &nonceJSON)
		if err != nil {
			return openwallet.NewError(openwallet.ErrCreateRawTransactionFailed, err.Error())
		}
	}
	nonceJSON[addrBalance.Address] = nonce

	rawTx.SetExtParam("nonce", nonceJSON)

	mostHeightBlock, err := decoder.wm.ApiClient.getMostHeightBlock()
	if err != nil {
		return openwallet.NewError(openwallet.ErrCreateRawTransactionFailed, err.Error())
	}

	toPub, err := decoder.wm.Decoder.AddressDecode(destination)
	if err != nil {
		return openwallet.NewError(openwallet.ErrCreateRawTransactionFailed, err.Error())
	}

	emptyTrans, hash, err := decoder.CreateEmptyRawTransactionAndMessage(addr.PublicKey, hex.EncodeToString(toPub), amount.Uint64(), nonce, feeInfo.Fee.Uint64(), mostHeightBlock, rawTx.Coin.Contract.Address)

	if err != nil {
		return openwallet.NewError(openwallet.ErrCreateRawTransactionFailed, err.Error())
	}
	rawTx.RawHex = emptyTrans

	if rawTx.Signatures == nil {
		rawTx.Signatures = make(map[string][]*openwallet.KeySignature)
	}

	keySigs := make([]*openwallet.KeySignature, 0)

	signature := openwallet.KeySignature{
		EccType: decoder.wm.Config.CurveType,
		Nonce:   "0x" + strconv.FormatUint(nonce, 16),
		Address: addr,
		Message: hash,
	}

	keySigs = append(keySigs, &signature)

	rawTx.Signatures[rawTx.Account.AccountID] = keySigs

	rawTx.FeeRate = strconv.FormatUint(feeInfo.Fee.Uint64(), 10)

	rawTx.IsBuilt = true

	return nil
}

//CreateSummaryRawTransactionWithError 创建汇总交易，返回能原始交易单数组（包含带错误的原始交易单）
func (decoder *TransactionDecoder) CreateSummaryRawTransactionWithError(wrapper openwallet.WalletDAI, sumRawTx *openwallet.SummaryRawTransaction) ([]*openwallet.RawTransactionWithError, error) {
	if sumRawTx.Coin.IsContract {
		return decoder.CreateTokenSummaryRawTransaction(wrapper, sumRawTx)
	}
	return nil, openwallet.Errorf(openwallet.ErrCreateRawTransactionFailed, "[%s] Miss contract details to create transaction!", sumRawTx.Account.AccountID)
}

func (decoder *TransactionDecoder) CreateEmptyRawTransactionAndMessage(fromPub string, toPub string, amount uint64, nonce uint64, fee uint64, mostHeightBlock *Block, assetIdStr string) (string, string, error) {

	runtimeVersion, err := decoder.wm.ApiClient.getRuntimeVersion()
	if err!=nil {
		return "", "", err
	}
	genesisHash, err := decoder.wm.ApiClient.getGenesisBlockHash()
	if err!=nil {
		return "", "", err
	}
	specVersion := runtimeVersion.SpecVersion
	txVersion := runtimeVersion.TransactionVersion

	assetId, err := strconv.ParseUint(assetIdStr, 10, 64)
	if err!=nil {
		return "", "", errors.New("wrong assetId "+assetIdStr)
	}

	tx := cennzTransaction.TxStruct{
		//发送方公钥
		SenderPubkey: fromPub,
		//接收方公钥
		RecipientPubkey: toPub,
		//发送金额（最小单位）
		Amount: amount,
		//资产id
		AssetId: assetId,
		//nonce
		Nonce: nonce,
		//手续费（最小单位）
		Fee: 0,
		//tip
		Tip: 0,
		//当前高度
		BlockHeight: mostHeightBlock.Height,
		//当前高度区块哈希
		BlockHash: RemoveOxToAddress(genesisHash),
		//创世块哈希
		GenesisHash: RemoveOxToAddress(genesisHash),
		//spec版本
		SpecVersion: specVersion,
		//TransactionVersion
		TxVersion : txVersion,
	}

	return tx.CreateEmptyTransactionAndMessage()
}
