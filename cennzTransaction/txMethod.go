package cennzTransaction

import (
	"encoding/hex"
	"errors"
	"github.com/blocktree/go-owcdrivers/polkadotTransaction/codec"
)

type MethodTransfer struct {
	DestPubkey []byte
	Amount     []byte
	AssetId    []byte
}

func NewMethodTransfer(pubkey string, amount, assetId uint64) (*MethodTransfer, error) {
	pubBytes, err := hex.DecodeString(pubkey)
	if  err != nil || len(pubBytes) != 32 {
		return nil, errors.New("invalid dest public key")
	}

	if amount == 0 {
		return nil, errors.New("zero amount")
	}
	amountStr, err := codec.Encode("Compact<u32>", amount)
	if err != nil {
		return nil, errors.New("invalid amount")
	}
	amountBytes, _ := hex.DecodeString(amountStr)

	if assetId == 0 {
		return nil, errors.New("zero assetId")
	}
	assetIdStr, err := codec.EncodeOld("Compact<u32>", assetId)
	if err != nil {
		return nil, errors.New("invalid assetId")
	}
	assetIdBytes, _ := hex.DecodeString(assetIdStr)

	return &MethodTransfer{
		DestPubkey: pubBytes,
		Amount:     amountBytes,
		AssetId:    assetIdBytes,
	}, nil
}

func (mt MethodTransfer) ToBytes(transferCode string) ([]byte, error) {

	if mt.DestPubkey == nil || len(mt.DestPubkey) != 32 || mt.Amount == nil || len(mt.Amount) == 0 {
		return nil, errors.New("invalid method")
	}

	ret, _ := hex.DecodeString(transferCode)
	if AccounntIDFollow {
		ret = append(ret, 0xff)
	}

	ret = append(ret, mt.AssetId...)
	ret = append(ret, mt.DestPubkey...)
	ret = append(ret, mt.Amount...)

	return ret, nil
}