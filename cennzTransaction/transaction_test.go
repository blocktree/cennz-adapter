package cennzTransaction

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/blocktree/go-owcdrivers/polkadotTransaction/codec"
	"testing"
)

//0x04040101fa2b41e0bf3969544b2cfe52908c7f60499450386649cf01d1fd86a56c4fd2d46304

func Test_CENNZ_transaction(t *testing.T) {

	tx := TxStruct{
		//发送方公钥
		SenderPubkey:    "xxxx",//"xxxx",
		//接收方公钥
		RecipientPubkey: "xxxx",
		//发送金额（最小单位）
		Amount:          20000,
		//资产ID
		AssetId:         2,
		//nonce
		Nonce:           8,
		//手续费（最小单位）
		Fee:             0,
		Tip:             0,
		//当前高度
		BlockHeight:     4571393,
		//当前高度区块哈希
		BlockHash:       "0d0971c150a9741b8719b3c6c9c2e96ec5b2e3fb83641af868e6650f3e263ef0",
		//创世块哈希
		GenesisHash:     "0d0971c150a9741b8719b3c6c9c2e96ec5b2e3fb83641af868e6650f3e263ef0",
		//spec版本
		SpecVersion:     37,
		//Transaction版本
		TxVersion: 5,
	}

	// 创建空交易单和待签消息
	emptyTrans, message, err := tx.CreateEmptyTransactionAndMessage()
	if err != nil {
		t.Error("create failed : ", err)
		return
	}
	fmt.Println("空交易单 ： ", emptyTrans)
	fmt.Println("待签消息 ： ",message)

	// 签名
	prikey, _ := hex.DecodeString("xxxx")
	signature, err := SignTransaction(message, prikey)
	if err != nil {
		t.Error("sign failed")
		return
	}
	fmt.Println("签名结果 ： ", hex.EncodeToString(signature))

	// 验签与交易单合并
	signedTrans, pass := VerifyAndCombineTransaction(emptyTrans, hex.EncodeToString(signature))
	if pass {
		fmt.Println("验签成功")
		fmt.Println("签名交易单 ： ", signedTrans)
	} else {
		t.Error("验签失败")
	}
}


func Test_json(t *testing.T)  {
	ts := TxStruct{
		SenderPubkey:    "123",
		RecipientPubkey: "",
		Amount:          0,
		Nonce:           0,
		Fee:             0,
		BlockHeight:     0,
		BlockHash:       "234",
		GenesisHash:     "345",
		SpecVersion:     0,
	}

	js, _ := json.Marshal(ts)

	fmt.Println(string(js))
}

func Test_decode(t *testing.T) {
	en, _ := codec.Encode(Compact_U32, uint64(139))
	fmt.Println(en)
}

func Test_Verify(t *testing.T){

}