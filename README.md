# cennz-adapter

本项目适配了openwallet.AssetsAdapter接口，给应用提供了底层的区块链协议支持。

## 如何测试

openwtester包下的测试用例已经集成了openwallet钱包体系，创建conf文件，新建CENNZ.ini文件，编辑如下内容：

```ini
# node api url public
nodeAPI = "http://xxx.xxx.xxx.xxx:xxxxx"
rpcAPI = "http:///xxx.xxx.xxx.xxx:xxxxx"

# fixed Fee in smallest unit
fixedFee = 15000
# Cache data file directory, default = "", current directory: ./data
dataDir = ""

APIChoose = "http"
decimal = 4
```

## 项目资料

在线钱包

https://cennznet.io/


浏览器

https://uncoverexplorer.com/


接口文档

https://github.com/polkadot-js/apps

https://github.com/polkadot-js/common

https://github.com/paritytech/substrate-api-sidecar


公共API

https://github.com/paritytech/substrate-api-sidecar/

精度 : 4, 确认数 100
目前链上手续费0.011，推荐收取商户0.05(mxc:0.1)
汇总时，需要保留0.01作为余额