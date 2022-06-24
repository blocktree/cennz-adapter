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
	"github.com/blocktree/openwallet/v2/openwallet"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"math/big"
	"strconv"
	"strings"
	"time"
)

type txFeeInfo struct {
	GasLimit *big.Int
	GasPrice *big.Int
	Fee      *big.Int
}

type Metadata struct {
	BlockNum        uint64
	NetworkNode     string
	SpecVersion     uint32
}

type RuntimeVersion struct {
	SpecVersion         uint32
	TransactionVersion  uint32
}

type AddrBalance struct {
	Address string
	AssetId string
	Balance *big.Int
	Free    *big.Int
	Freeze  *big.Int
	Nonce   uint64
	index   int
	Actived bool
}

type CennzAddrBalance struct {
	Address      string
	Balance      *big.Int
	FeeBalance  *AddrBalance
	Index        int
}

type Block struct {
	Hash          string        `json:"block"`         // actually block signature in XRP chain
	PrevBlockHash string        `json:"previousBlock"` // actually block signature in DOT chain
	Timestamp     uint64        `json:"timestamp"`
	Height        uint64        `json:"height"`
	Transactions  []Transaction `json:"transactions"`
	Finalized      bool          `json:"finalized"`
}

type Extrinsic struct {
	Block_num uint64
	Block_timestamp uint64
	Extrinsic_hash string
	Call_module string
	Call_module_function string
	ToArr       []string //@required 格式："地址":"数量":资产id
	ToDecArr    []string //@required 格式："地址":"数量(带小数)":资产id
	From        string
	Fee         string
	Status      string
	Index       uint64
}

type Transaction struct {
	TxID        string
	Fee         uint64
	TimeStamp   uint64
	From        string
	To          string
	Amount      uint64
	BlockHeight uint64
	BlockHash   string
	Status      string
	//ToArr       []string //@required 格式："地址":"数量":资产id
	//ToDecArr    []string //@required 格式："地址":"数量(带小数)":资产id
	//FromArr     []string //@required 格式："地址":"数量(带小数)":资产id
	ToTrxDetailArr       []TrxDetail
	FromTrxDetailArr     []TrxDetail
}

type TrxDetail struct {
	Addr        string
	Amount      string
	AmountDec   string
	AssetId     string
}

func GetApiData(json *gjson.Result) (*gjson.Result, error){
	if gjson.Get(json.Raw, "code").Exists()==false {
		return nil, errors.New("api code not found")
	}

	code := gjson.Get(json.Raw, "code").Uint()
	if code!=0 {
		return nil, errors.New("wrong api code " + strconv.FormatUint(code, 10) )
	}

	if gjson.Get(json.Raw, "data").Exists()==false {
		return nil, errors.New("api data not found")
	}

	dataJSON := gjson.Get(json.Raw, "data")

	return &dataJSON, nil
}

func GetMetadata(json *gjson.Result) (*Metadata, error) {
	obj := &Metadata{}

	dataJSON, err :=  GetApiData(json)
	if err!=nil {
		return nil, err
	}

	obj.BlockNum = gjson.Get(dataJSON.Raw, "blockNum").Uint()
	obj.NetworkNode = gjson.Get(dataJSON.Raw, "networkNode").String()
	obj.SpecVersion = uint32(gjson.Get(dataJSON.Raw, "specVersion").Uint())

	return obj, nil
}

func GetRuntimeVersion(json *gjson.Result) (*RuntimeVersion, error) {
	obj := &RuntimeVersion{}

	obj.SpecVersion = uint32( gjson.Get(json.Raw, "specVersion").Uint() )
	obj.TransactionVersion = uint32( gjson.Get(json.Raw, "transactionVersion").Uint() )

	return obj, nil
}

func NewBlock(json *gjson.Result, symbol string) *Block {
	obj := &Block{}
	// 解析
	obj.Hash = gjson.Get(json.Raw, "hash").String()
	obj.PrevBlockHash = gjson.Get(json.Raw, "parent_hash").String()
	obj.Height = gjson.Get(json.Raw, "block_num").Uint()
	obj.Timestamp = gjson.Get(json.Raw, "block_timestamp").Uint()
	obj.Finalized = gjson.Get(json.Raw, "finalized").Bool()
	obj.Transactions = GetTransactionInBlock(json, symbol)

	if obj.Hash == "" {
		time.Sleep(5 * time.Second)
	}
	return obj
}

func NewBlockFromRpc(json *gjson.Result, symbol string) (*Block, error) {
	obj := &Block{}
	// 解析
	obj.Hash = gjson.Get(json.Raw, "hash").String()
	obj.PrevBlockHash = gjson.Get(json.Raw, "parentHash").String()
	obj.Height = gjson.Get(json.Raw, "number").Uint()
	obj.Finalized = gjson.Get(json.Raw, "finalized").Bool()

	transactions, blockTime, err := GetTransactionAndBlockTimeInBlock(json, symbol)
	if err!=nil {
		return nil, err
	}
	obj.Timestamp = blockTime
	obj.Transactions = transactions

	if obj.Hash == "" {
		time.Sleep(5 * time.Second)
	}
	return obj, nil
}

//BlockHeader 区块链头
func (b *Block) BlockHeader() *openwallet.BlockHeader {

	obj := openwallet.BlockHeader{}
	//解析json
	obj.Hash = b.Hash
	obj.Previousblockhash = b.PrevBlockHash
	obj.Height = b.Height

	return &obj
}

func GetTransactionAndBlockTimeInBlock(json *gjson.Result, symbol string) ([]Transaction, uint64, error) {
	transactions := make([]Transaction, 0)

	blockHash := gjson.Get(json.Raw, "hash").String()
	blockHeight := gjson.Get(json.Raw, "number").Uint()

	blockTime := uint64(time.Now().Unix())

	extrinsicMap := make(map[uint64]Extrinsic)	// key = extrinsicIndex, value = extrinsic
	transactionMap := make(map[uint64]Transaction)    // key = extrinsicIndex, value = Transaction

	for extrinsicIndex, extrinsicJSON := range gjson.Get(json.Raw, "extrinsics").Array() {
		section := gjson.Get(extrinsicJSON.Raw, "section").String()
		method := gjson.Get(extrinsicJSON.Raw, "method").String()
		isSigned := gjson.Get(extrinsicJSON.Raw, "isSigned").Bool()
		txid := gjson.Get(extrinsicJSON.Raw, "hash").String()
		args := gjson.Get(extrinsicJSON.Raw, "args").Array()

		//log.Debug("section : ", section, "method : ", method, ", txid : ", txid, ", isSigned : ", isSigned, ", args : ", args)

		//获取这个区块的时间
		if section == "timestamp" && method=="set" {
			args := gjson.Get(extrinsicJSON.Raw, "args")
			if len(args.Raw) >0 {
				blockTime = args.Array()[0].Uint()
			}
		}

		if !isSigned {
			continue
		}

		isSimpleTransfer := section=="genericAsset" && method=="transfer"
		if isSimpleTransfer {
			if len( args ) != 3{
				log.Error("wrong extrinsic args length : ", txid)
				continue
			}

			assetId := args[0].String()
			to := args[1].String()
			from := gjson.Get(extrinsicJSON.Raw, "signer").String()
			amount := args[2].String()

			if to=="" || amount=="" || assetId=="" || from==""{
				log.Error("wrong txid : ", txid)
				continue
			}

			toStr := to + ":" + amount + ":" + assetId

			toArr := make([]string, 0)
			toArr = append(toArr, toStr)
			fee := gjson.Get(extrinsicJSON.Raw, "partialFee").String()

			extrinsic := Extrinsic{
				Extrinsic_hash:       txid,
				Call_module:          section,
				Call_module_function: method,
				ToArr:                toArr,
				ToDecArr:             nil,
				From:                 from,
				Fee:                  fee,
				Status:               "0",
				Index:                uint64(extrinsicIndex),
			}

			extrinsicMap[ uint64(extrinsicIndex) ] = extrinsic
		}
	}

	for _, eventJSON := range gjson.Get(json.Raw, "events").Array() {
		phase := gjson.Get(eventJSON.Raw, "phase")
		if !phase.Exists() {
			continue
		}
		extrinsicIndex := gjson.Get(phase.Raw, "applyExtrinsic").Uint()

		eventIndex := gjson.Get(eventJSON.Raw, "index").String()
		eventMethod := gjson.Get(eventJSON.Raw, "method").String()

		if eventIndex=="0x0000" && eventMethod=="ExtrinsicSuccess" {	//指明，当前transaction的status可以改为1
			transaction, ok := transactionMap[extrinsicIndex]
			if ok {
				transaction.Status = "1"
				transactions = append(transactions, transaction)
			}
		}

		if eventIndex=="0x0401" && eventMethod=="Transferred" {
			extrinsic, ok := extrinsicMap[extrinsicIndex]
			if ok {
				if gjson.Get(eventJSON.Raw, "data").Exists()==false {
					continue
				}

				data := gjson.Get(eventJSON.Raw, "data").Array()

				if len(data) != 4{
					log.Error("wrong event args length : ", extrinsic.Extrinsic_hash)
					continue
				}

				assetId := data[0].String()
				from := data[1].String()
				to := data[2].String()
				amount := data[3].String()

				eventTo := to + ":" + amount + ":" + assetId

				//普通转账，一个txid，只有一笔资金转账
				if eventTo != extrinsic.ToArr[0]{
					log.Error("failed txid : ", extrinsic.Extrinsic_hash)
					continue
				}

				if from == "" {
					log.Error("from not found txid : ", extrinsic.Extrinsic_hash)
					continue
				}

				feeInt, feeErr := strconv.ParseInt(extrinsic.Fee, 10, 64)
				amountInt, err := strconv.ParseInt(amount, 10, 64)
				if err == nil  && feeErr == nil{
					amountUint := uint64(amountInt)
					fee := uint64(feeInt)

					toTrxDetailArr := make([]TrxDetail, 0)
					toTrxDetail := TrxDetail{
						Addr:      to,
						Amount:    amount,
						AmountDec: "",
						AssetId:   assetId,
					}
					toTrxDetailArr = append(toTrxDetailArr, toTrxDetail)

					fromTrxDetailArr := make([]TrxDetail, 0)
					fromTrxDetail := TrxDetail{
						Addr:      from,
						Amount:    amount,
						AmountDec: "",
						AssetId:   assetId,
					}
					fromTrxDetailArr = append(fromTrxDetailArr, fromTrxDetail)

					if feeInt>0 {
						feeTrxDetail := TrxDetail{
							Addr:      from,
							Amount:    extrinsic.Fee,
							AmountDec: "",
							AssetId:  feeToken.Address,
						}
						fromTrxDetailArr = append(fromTrxDetailArr, feeTrxDetail)
					}

					transaction := Transaction{
						TxID:             extrinsic.Extrinsic_hash,
						TimeStamp:        blockTime,
						From:             from,
						To:               to,
						Amount:           amountUint,
						BlockHeight:      blockHeight,
						BlockHash:        blockHash,
						Status:           "0",
						ToTrxDetailArr:   toTrxDetailArr,
						FromTrxDetailArr: fromTrxDetailArr,
						Fee :             fee,
					}

					transactionMap[extrinsicIndex] = transaction
				}
			}
		}
	}

	return transactions, blockTime, nil
}

func GetTransactionInBlock(json *gjson.Result, symbol string) []Transaction {
	transactions := make([]Transaction, 0)

	blockHash := gjson.Get(json.Raw, "hash").String()
	blockHeight := gjson.Get(json.Raw, "block_num").Uint()

	blockTime := uint64(time.Now().Unix())

	extrinsicMap := make(map[string]Extrinsic)

	for _, extrinsicJSON := range gjson.Get(json.Raw, "extrinsics").Array() {
		call_module := gjson.Get(extrinsicJSON.Raw, "call_module").String()
		call_module_function := gjson.Get(extrinsicJSON.Raw, "call_module_function").String()
		success := gjson.Get(extrinsicJSON.Raw, "success").Bool()
		//finalized := gjson.Get(extrinsicJSON.Raw, "finalized").Bool()
		paramsStr := gjson.Get(extrinsicJSON.Raw, "params").String()

		txid := gjson.Get(extrinsicJSON.Raw, "extrinsic_hash").String()

		//log.Debug("call_module : ", call_module, "call_module_function : ", call_module_function, ", txid : ", txid, ", finalized : ", finalized, ", success : ", success, ", paramsStr : ", paramsStr)

		//if !success || !finalized{
		if !success{
			continue
		}

		isSimpleTransfer := call_module=="genericAsset" && call_module_function=="transfer"
		if isSimpleTransfer {
			assetId := ""
			to := ""
			amount := ""

			params := gjson.Parse(paramsStr).Array()
			for _, param := range params {

				if gjson.Get(param.Raw, "name").Exists() == false {
					continue
				}
				paramName := gjson.Get(param.Raw, "name").String()

				if paramName == "asset_id" {
					if gjson.Get(param.Raw, "value").Exists() == false {
						continue
					}
					assetId = gjson.Get(param.Raw, "value").String()
				}
				if paramName == "to" {
					if gjson.Get(param.Raw, "type").Exists() == false {
						continue
					}
					paramType := gjson.Get(param.Raw, "type").String()
					if paramType != "AccountId" {
						continue
					}

					if gjson.Get(param.Raw, "value").Exists() == false {
						continue
					}
					to = gjson.Get(param.Raw, "value").String()
				}
				if paramName == "amount" {
					if gjson.Get(param.Raw, "type").Exists() == false {
						continue
					}
					paramType := gjson.Get(param.Raw, "type").String()
					if paramType != "Compact<Balance>" {
						continue
					}

					if gjson.Get(param.Raw, "value").Exists() == false {
						continue
					}
					amount = gjson.Get(param.Raw, "value").String()
				}
			}

			if to=="" && amount=="" && assetId=="" {
				log.Error("wrong txid : ", txid)
				continue
			}

			toStr := to + ":" + amount + ":" + assetId

			toArr := make([]string, 0)
			toArr = append(toArr, toStr)
			fee := gjson.Get(extrinsicJSON.Raw, "fee").String()

			extrinsic := Extrinsic{
				Extrinsic_hash:       txid,
				Call_module:          call_module,
				Call_module_function: call_module_function,
				ToArr:                toArr,
				ToDecArr:             nil,
				From:                 "",
				Fee:                  fee,
				Status:               "0",
			}

			extrinsicMap[txid] = extrinsic
		}
	}

	for _, eventJSON := range gjson.Get(json.Raw, "events").Array() {
		module_id := gjson.Get(eventJSON.Raw, "module_id").String()
		event_id := gjson.Get(eventJSON.Raw, "event_id").String()
		//finalized := gjson.Get(eventJSON.Raw, "finalized").Bool()

		isSimpleTransfer := module_id=="genericAsset" && event_id=="Transferred"
		//if isSimpleTransfer && finalized {
		if isSimpleTransfer {
			extrinsic_hash := gjson.Get(eventJSON.Raw, "extrinsic_hash").String()
			extrinsic, ok := extrinsicMap[extrinsic_hash]
			if ok {
				assetId := ""
				from := ""
				to := ""
				amount := ""

				if gjson.Get(eventJSON.Raw, "params").Exists()==false {
					continue
				}
				paramsStr := gjson.Get(eventJSON.Raw, "params").String()
				params := gjson.Parse(paramsStr).Array()
				for paramIndex, param := range params {
					if gjson.Get(param.Raw, "type").Exists() == false {
						continue
					}
					eventType := gjson.Get(param.Raw, "type").String()

					if eventType == "AssetId" {
						if gjson.Get(param.Raw, "value").Exists() == false {
							continue
						}
						assetId = gjson.Get(param.Raw, "value").String()
					}
					if eventType == "AccountId" && paramIndex==1{
						if gjson.Get(param.Raw, "value").Exists() == false {
							continue
						}
						from = gjson.Get(param.Raw, "value").String()
					}
					if eventType == "AccountId" && paramIndex==2{
						if gjson.Get(param.Raw, "value").Exists() == false {
							continue
						}
						to = gjson.Get(param.Raw, "value").String()
					}
					if eventType == "Balance" {
						if gjson.Get(param.Raw, "value").Exists() == false {
							continue
						}
						amount = gjson.Get(param.Raw, "value").String()
					}
				}

				eventTo := to + ":" + amount + ":" + assetId

				//普通转账，一个txid，只有一笔资金转账
				if eventTo != extrinsic.ToArr[0]{
					log.Error("failed txid : ", extrinsic.Extrinsic_hash)
					continue
				}

				if from == "" {
					log.Error("from not found txid : ", extrinsic.Extrinsic_hash)
					continue
				}

				feeInt, feeErr := strconv.ParseInt(extrinsic.Fee, 10, 64)
				amountInt, err := strconv.ParseInt(amount, 10, 64)
				if err == nil  && feeErr == nil{
					amountUint := uint64(amountInt)
					fee := uint64(feeInt)

					toTrxDetailArr := make([]TrxDetail, 0)
					toTrxDetail := TrxDetail{
						Addr:      to,
						Amount:    amount,
						AmountDec: "",
						AssetId:   assetId,
					}
					toTrxDetailArr = append(toTrxDetailArr, toTrxDetail)

					fromTrxDetailArr := make([]TrxDetail, 0)
					fromTrxDetail := TrxDetail{
						Addr:      from,
						Amount:    amount,
						AmountDec: "",
						AssetId:   assetId,
					}
					fromTrxDetailArr = append(fromTrxDetailArr, fromTrxDetail)

					if feeInt>0 {
						feeTrxDetail := TrxDetail{
							Addr:      from,
							Amount:    extrinsic.Fee,
							AmountDec: "",
							AssetId:  feeToken.Address,
						}
						fromTrxDetailArr = append(fromTrxDetailArr, feeTrxDetail)
					}

					transaction := Transaction{
						TxID:             extrinsic_hash,
						TimeStamp:        blockTime,
						From:             from,
						To:               to,
						Amount:           amountUint,
						BlockHeight:      blockHeight,
						BlockHash:        blockHash,
						Status:           "1",
						ToTrxDetailArr:   toTrxDetailArr,
						FromTrxDetailArr: fromTrxDetailArr,
						Fee :             fee,
					}

					transactions = append(transactions, transaction)
				}
			}
		}
	}

	return transactions
}

// 从最小单位的 amount 转为带小数点的表示
func convertToAmount(amount uint64, amountDecimal uint64) string {
	amountStr := fmt.Sprintf("%d", amount)
	d, _ := decimal.NewFromString(amountStr)
	ten := math.BigPow(10, int64(amountDecimal) )
	w, _ := decimal.NewFromString(ten.String())

	d = d.Div(w)
	return d.String()
}

// amount 字符串转为最小单位的表示
func convertFromAmount(amountStr string, amountDecimal uint64) uint64 {
	d, _ := decimal.NewFromString(amountStr)
	ten := math.BigPow(10, int64(amountDecimal) )
	w, _ := decimal.NewFromString(ten.String())
	d = d.Mul(w)
	r, _ := strconv.ParseInt(d.String(), 10, 64)
	return uint64(r)
}

func RemoveOxToAddress(addr string) string {
	if strings.Index(addr, "0x") == 0 {
		return addr[2:]
	}
	return addr
}