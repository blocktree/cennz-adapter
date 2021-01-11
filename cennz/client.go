package cennz

import "errors"

const APIClientHttpMode = "http"

type ApiClient struct {
	Client    *Client
	RpcClient *RpcClient
	APIChoose string
}

func NewApiClient(wm *WalletManager) error {
	api := ApiClient{}

	if len(wm.Config.APIChoose) == 0 {
		wm.Config.APIChoose = APIClientHttpMode //默认采用rpc连接
	}
	api.APIChoose = wm.Config.APIChoose
	if api.APIChoose == APIClientHttpMode {
		api.Client = NewClient(wm.Config.NodeAPI, false, wm.Symbol() )
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
	}

	return block, err
}

func (c *ApiClient) sendTransaction(rawTx string) (string, error) {
	var (
		txid string
		err  error
	)
	if c.APIChoose == APIClientHttpMode {
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
	if c.APIChoose == APIClientHttpMode {
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
	}

	return mostHeightBlock, err
}

func (c *ApiClient) getGenesisBlockHash() (string, error) {
	var (
		result string
		err    error
	)
	if c.APIChoose == APIClientHttpMode {
		result, err = c.RpcClient.GetGenesisHash()
	}

	return result, err
}
