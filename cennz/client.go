package cennz

import (
	"errors"
)

const APIClientHttpMode = "http"
const APIClientAllRpcMode = "allRpc"

type ApiClient struct {
	Client    *Client
	RpcClient *RpcClient
	BalanceApiClient *BalanceApiClient
	APIChoose string
}

func NewApiClient(wm *WalletManager) error {
	api := ApiClient{}

	if len(wm.Config.APIChoose) == 0 {
		wm.Config.APIChoose = APIClientHttpMode //默认采用rpc连接
	}
	if len(wm.Config.APIChoose) == 1 {
		wm.Config.APIChoose = APIClientAllRpcMode //默认采用rpc连接
	}
	api.APIChoose = wm.Config.APIChoose
	if api.APIChoose == APIClientHttpMode {
		api.Client = NewClient(wm.Config.NodeAPI, false, wm.Symbol() )
		api.BalanceApiClient = NewBalanceClient(wm.Config.BalanceAPI, false, wm.Symbol())
		api.RpcClient = NewRpcClient(wm.Config.RpcAPI, false, wm.Symbol() )
	}
	if api.APIChoose == APIClientAllRpcMode {
		api.BalanceApiClient = NewBalanceClient(wm.Config.BalanceAPI, false, wm.Symbol())
		api.RpcClient = NewRpcClient(wm.Config.RpcAPI, false, wm.Symbol() )
	}

	wm.ApiClient = &api

	return nil
}

// 获取当前最高区块
func (c *ApiClient) getBlockHeight() (uint64, error) {
	var (
		currentHeight uint64
		err           error
	)
	if c.APIChoose == APIClientHttpMode {
		currentHeight, err = c.Client.getBlockHeight()
	}else if c.APIChoose == APIClientAllRpcMode {
		currentHeight, err = c.RpcClient.getBlockHeight()
	}

	return currentHeight, err
}

// 获取地址余额
func (c *ApiClient) getBalance(address string, assetId string) (*AddrBalance, error) {
	var (
		balance *AddrBalance
		err     error
	)

	if c.APIChoose == APIClientHttpMode {
		balance, err = c.Client.getBalance(address, assetId)
		if err != nil {
			return nil, err
		}

		finalizedHeadBlockHash, err := c.RpcClient.GetFinalizedHead()
		if err != nil {
			return nil, err
		}

		balance, err = c.BalanceApiClient.getApiBalance(balance, finalizedHeadBlockHash)
	}else if c.APIChoose == APIClientAllRpcMode {
		finalizedHeadBlockHash, err := c.RpcClient.GetFinalizedHead()
		if err != nil {
			return nil, err
		}

		balance, err = c.BalanceApiClient.getApiBalanceWithNonce(address, finalizedHeadBlockHash, assetId)
	}

	return balance, err
}

func (c *ApiClient) getBlockByHeight(height uint64) (*Block, error) {
	var (
		block *Block
		err   error
	)
	if c.APIChoose == APIClientHttpMode {
		block, err = c.Client.getBlockByHeight(height)
		if err!=nil {
			return nil, err
		}

		hashInRpc, err := c.RpcClient.GetBlockHash(height)
		if err != nil {
			return nil, err
		}

		if hashInRpc!=block.Hash {
			return nil, errors.New("wrong block, rpc :" + hashInRpc + ", http : " + block.Hash )
		}
	}else if c.APIChoose == APIClientAllRpcMode {
		block, err = c.BalanceApiClient.getBlockByHeight(height)
		if err!=nil {
			return nil, err
		}
	}

	return block, err
}

func (c *ApiClient) sendTransaction(rawTx string) (string, error) {
	var (
		txid string
		err  error
	)
	if c.APIChoose == APIClientHttpMode || c.APIChoose == APIClientAllRpcMode {
		txid, err = c.RpcClient.sendTransaction(rawTx)
	}

	return txid, err
}

func (c *ApiClient) getMetadata() (*Metadata, error) {
	var (
		metadata    *Metadata
		err         error
	)
	if c.APIChoose == APIClientHttpMode {
		metadata, err = c.Client.getMetaData()
	}

	return metadata, err
}

func (c *ApiClient) getRuntimeVersion() (*RuntimeVersion, error){
	var (
		result    *RuntimeVersion
		err         error
	)
	if c.APIChoose == APIClientHttpMode || c.APIChoose == APIClientAllRpcMode {
		result, err = c.RpcClient.GetRuntimeVersion()
	}

	return result, err
}

//获取当前最新高度
func (c *ApiClient) getMostHeightBlock() (*Block, error) {
	var (
		mostHeightBlock *Block
		err             error
	)
	if c.APIChoose == APIClientHttpMode {
		mostHeightBlock, err = c.Client.getMostHeightBlock()
	}else if c.APIChoose == APIClientAllRpcMode {
		mostHeightBlock, err = c.RpcClient.getMostHeightBlock()
	}

	return mostHeightBlock, err
}

func (c *ApiClient) getGenesisBlockHash() (string, error) {
	var (
		result string
		err    error
	)
	if c.APIChoose == APIClientHttpMode || c.APIChoose == APIClientAllRpcMode {
		result, err = c.RpcClient.GetGenesisHash()
	}

	return result, err
}
