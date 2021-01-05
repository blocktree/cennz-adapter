/*
 * Copyright 2019 The openwallet Authors
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
	"github.com/blocktree/openwallet/v2/log"
	"github.com/blocktree/openwallet/v2/openwallet"
	"testing"
)

func TestContractDecoder_GetTokenBalanceByAddress(t *testing.T) {
	addr := ""
	contract := openwallet.SmartContract{
		Address:  "1",
		Symbol:   "CENNZ",
		Name:     "CENNZ",
		Token:    "CENNZ",
		Decimals: 4,
	}

	tw.InitTokenMap()

	balances, err := tw.ContractDecoder.GetTokenBalanceByAddress(contract, addr)
	if err != nil {
		log.Errorf(err.Error())
		return
	}
	for _, b := range balances {
		log.Infof("balance[%s] = %s", b.Balance.Address, b.Balance.Balance)
		log.Infof("UnconfirmBalance[%s] = %s", b.Balance.Address, b.Balance.UnconfirmBalance)
		log.Infof("ConfirmBalance[%s] = %s", b.Balance.Address, b.Balance.ConfirmBalance)
	}
}
