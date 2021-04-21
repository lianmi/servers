# 这个目录是在mac里运行

##   编译智能合约 

```
$ cd internal/pkg/blockchain/lnmc_server/contracts/ERC20
$ solc --abi --bin ERC20Token.sol -o build --overwrite 
$ abigen --bin=./build/ERC20Token.bin --abi=./build/ERC20Token.abi --pkg=contracts --out=ERC20Token.go


```

 ## 发币 LNMC 10000亿枚
 
 ### 部署发币智能合约，并生成合约地址
```
$ cd /Users/mac/developments/lianmi/lm-cloud/servers/internal/pkg/blockchain/lnmc
$ go run deploy.go

输出:

```

 ### 将生成的发币合约地址保存到配置
 ```

 ```