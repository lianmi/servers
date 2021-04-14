# 服务端专用, 用于初始化并部署发币智能合约

智能合约地址保存在redis的key： ERC20DeployContractAddress
```
$ redis-cli
127.0.0.1:6379> get  ERC20DeployContractAddress
"0x1D2bDDA8954b401fEB52008C63878e698b6B8444"
``` 