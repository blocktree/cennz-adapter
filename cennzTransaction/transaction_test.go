package cennzTransaction

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"
)

// 0xa0 0401 04 7dd904f18b1e42c7f2b62429245771f51749a42be06c02bd4251df69f2c141df 0350dc88f400 30000025000000050000000d0971c150a9741b8719b3c6c9c2e96ec5b2e3fb83641af868e6650f3e263ef00d0971c150a9741b8719b3c6c9c2e96ec5b2e3fb83641af868e6650f3e263ef0
//      0401 04 7dd904f18b1e42c7f2b62429245771f51749a42be06c02bd4251df69f2c141df 0750dc88f400 0010000025000000050000000d0971c150a9741b8719b3c6c9c2e96ec5b2e3fb83641af868e6650f3e263ef00d0971c150a9741b8719b3c6c9c2e96ec5b2e3fb83641af868e6650f3e263ef0

// 0x39028488dc3417d5058ec4b4503e0c12ea1a0a89be200fe98922423d4334014fa6b0ee00ec59fa45d747eeaca87ff82fb7cd41137fcc976f6da4e010e46cc337d0033f1171a8b38f2692d467ba1019b7cbd66959ae3b1e632308ea4a5a2dd2ab27818f0c003000000401047dd904f18b1e42c7f2b62429245771f51749a42be06c02bd4251df69f2c141df0327ee88f4
// 0x3d028486377c388ec1afc558ef40c5edb3b4f7bba1a697b1bb711ece23fc7cdbfe2e1f0054c9cc865e71f4da69821dee315fe9fc01c422d954128740a76123c78392ecb2ff69eeaf6c307f133c8f4728f8e7958fbb39d8a3cb603f12482f2ebfaeada501001000000401047dd904f18b1e42c7f2b62429245771f51749a42be06c02bd4251df69f2c141df0727ee88f400

// 5F6gkpLsrdFaWtUu4UH87qFJtRJb5DpLuN7UaDkspDwNef5D -> 5EuiJzHg1uGqPN5Rf28tdNSRzC8G6RoEJFHb44Z9rnkaze6d
func Test_CENNZ_transaction(t *testing.T) {

	tx := TxStruct{
		//发送方公钥
		SenderPubkey:    "86377c388ec1afc558ef40c5edb3b4f7bba1a697b1bb711ece23fc7cdbfe2e1f",//"88dc3417d5058ec4b4503e0c12ea1a0a89be200fe98922423d4334014fa6b0ee",
		//接收方公钥
		RecipientPubkey: "7dd904f18b1e42c7f2b62429245771f51749a42be06c02bd4251df69f2c141df",
		//发送金额（最小单位）
		Amount:         1303822400,
		//资产ID
		AssetId:         1,
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
	prikey, _ := hex.DecodeString("e86bcaaab0a5aa5e3f3b0885db7e932e34eddb5a620b6bcc097a4b236a5a0354")
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
	//en, _ := codec.Encode(Compact_U32, uint64(139))
	//fmt.Println(en)
}

func Test_Verify(t *testing.T){

}