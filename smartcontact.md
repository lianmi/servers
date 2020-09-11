# 钱包与智能合约

##  区块链

### 创建创世区块
```
$ geth init --datadir node0 genesis.json
$ geth account new --password passwd --datadir node0
```

### 以太坊单节点版
```
$ cd /Users/mac/developments/lianmi/lm-cloud/ethereum-single-node
$ docker-compose up -d

$ docker ps

#进入命令行
$ docker exec -it simplenode sh

```


### 外部进入geth控制台
```
$ geth attach http://localhost:8545
```