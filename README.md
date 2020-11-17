这是连米信息科技后端微服务整体框架项目，囊括了项目的结构、分层思想、依赖注入、错误处理、单元测试、服务治理、框架选择等方面.

项目分为dispatcher、orderservice、chatservice、walletservice等微服务。
另外内部联盟链是单节点的以太坊。
ipfs星际文件系统


## 准备
参考：Golang微服务实践 https://github.com/sdgmf/go-project-sample
安装docker,go,[jsonnet](https://jsonnet.org/)

### 部署jaeger
```
$ docker pull jaegertracing/all-in-one:latest
$ docker run -d --name jaeger -e COLLECTOR_ZIPKIN_HTTP_PORT=9411 -p 5775:5775/udp -p 6831:6831/udp -p 6832:6832/udp -p 5778:5778 -p 16686:16686 -p 14268:14268 -p 9411:9411 jaegertracing/all-in-one:latest
```
### 运行consul
在一个新的终端窗口
```
$ consul agen -dev
```

## 运行mysql:
docker方式的mysql 5.7
```
$ cd ../database
$ sh runmysql.sh
```

## 快速开始
### 下载项目
```bash
    git clone https://github.com/lianmi/servers.git
    cd servers
    git submodule init
    git submodule update
    make build
```
如果出现： cannot load github.com/hashicorp/consul/api: ambiguous import 
修改go.mod
```
replace github.com/hashicorp/consul => github.com/hashicorp/consul v1.5.1
```


* **访问接口**：post http://localhost:28080/register
* **consul**: http://localhost:8500/
* **grafana**: http://localhost:3000/ 
* **jaeger**: http://localhost:16686/search
* **Prometheus**: http://localhost:9090/graph
* **AlertManager**: http://localhost:9093



### Grafana Dashboard,可以自动生成!


### Prometheus Alert 监控告警,自动生成！

### 调用链跟踪


## 使用eclipse paho.golang master版本
```
go mod edit -replace=github.com/eclipse/paho.golang@v0.9.0=github.com/eclipse/paho.golang@master
```

## proto文件编译
```
make proto

```

## 定时任务类
https://github.com/robfig/cron
例子在： /Users/mac/developments/lianmi/lm-cloud/cron_demo
使用说明：
https://www.jianshu.com/p/fd3dda663953

 
## 本项目采用Gorm, 但需要创建表
internal/pkg/database/database.go :

```
db.AutoMigrate(&models.User{}) //用户表
```

GORM 中文文档
```
http://gorm.book.jasperxu.com/
```

## Gin JWT中间件
```
https://github.com/appleboy/gin-jwt
```

例子
```
https://github.com/Bingjian-Zhu/gin-vue
```

### 后台
```
https://github.com/Bingjian-Zhu/gin-vue-admin
```

### JWT+Redis实现用户登录验证
```

https://blog.csdn.net/mirage003/article/details/87865582
```

### ipfs client
```

```

##  mysql新建用户并授权
```
USE mysql;
set global validate_password.policy=0;
set global validate_password.length=1;

CREATE USER lianmidba IDENTIFIED BY '12345678';

CREATE USER lianmidba IDENTIFIED BY 'lianmicloud!@#$1234';

set password for 'lianmidba'@'localhost'=password('lianmicloud!@#$1234');

//让lianmidba拥有lianmicloud数据库的所有权限
GRANT ALL PRIVILEGES ON lianmicloud.* TO 'lianmidba'@'%';
FLUSH PRIVILEGES;

use lianmicloud;
SELECT * FROM USER WHERE USER='lianmidba' ;
SHOW GRANTS FOR lianmidba;

```

## 平台 linux  

### 1 编译 
```
$ make linux
```

### 2 停止（首次运行不需要）
```
$ make stop
```

### 3 构造镜像并运行
```
$ make docker-compose
```

### 4 检查是否正常运行
```
$ docker ps
$ netstat -tunlp|grep 28080
$ ps -el|grep dispatcher
$ ps -el|grep authervice
```
如果运行成功，都会出现正常结果

## Restful http 接口测试

### 安装httpie
Download and install [httpie](https://github.com/jkbrzt/httpie) CLI HTTP client.

### 注册


```sh
http -v --json POST localhost:28080/register username=lsj001 password=C33367701511B4F6020EC61DED352059 gender=1 mobile=13702290109 user_type=1 contact_person=李示佳
```

### 登录

```sh
http -v --json POST localhost:28080/login username=lsj001 password=C33367701511B4F6020EC61DED352059 
```

输出：
```
POST /login HTTP/1.1
Accept: application/json, */*;q=0.5
Accept-Encoding: gzip, deflate
Connection: keep-alive
Content-Length: 70
Content-Type: application/json
Host: localhost:28080
User-Agent: HTTPie/2.2.0

{
    "password": "C33367701511B4F6020EC61DED352059",
    "username": "lsj001"
}

HTTP/1.1 200 OK
Content-Length: 302
Content-Type: application/json; charset=utf-8
Date: Fri, 14 Aug 2020 02:39:20 GMT
Location: https://:28080/login

{
    "code": 200,
    "data": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1OTczNzYzNjAsIm9yaWdfaWF0IjoxNTk3MzcyNzYwLCJ1c2VyTmFtZSI6ImxzajAwMSIsInVzZXJSb2xlcyI6Ilt7XCJpZFwiOjIsXCJ1c2VyX2lkXCI6MixcInVzZXJfbmFtZVwiOlwibHNqMDAxXCIsXCJ2YWx1ZVwiOlwiXCJ9XSJ9.Kz8JNpAggbfGeCG1Ky2H6r4Qxe8shdqxXj46GC94JNU",
    "msg": "ok"
}

```

### 查询
```
http -v --json GET localhost:28080/v1/user/2 "Authorization:Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1OTczNzYzNjAsIm9yaWdfaWF0IjoxNTk3MzcyNzYwLCJ1c2VyTmFtZSI6ImxzajAwMSIsInVzZXJSb2xlcyI6Ilt7XCJpZFwiOjIsXCJ1c2VyX2lkXCI6MixcInVzZXJfbmFtZVwiOlwibHNqMDAxXCIsXCJ2YWx1ZVwiOlwiXCJ9XSJ9.Kz8JNpAggbfGeCG1Ky2H6r4Qxe8shdqxXj46GC94JNU"  "Content-Type: application/json"
```

## mosquitto鉴权
```
mosquitto-1.6.10/test/random/auth_plugin.c
```

# 服务器的日志目录
部署在腾讯云服务器后，日志输出需要在 /root/developments/lianmi/lm-cloud/servers/deployments 建立.env文件
内容是：
```
LOG_DIR=/root/developments/lianmi/work/logs
``` 

export env_file=./aa/deployments/.env
mkdir -p "${env_file%/*}" && echo "LOG_DIR=/root/developments/lianmi/work/logs" > $env_file

export env_file=/root/developments/lianmi/lm-cloud/servers/deployments/.env
mkdir -p "${env_file%/*}" && echo "LOG_DIR=/root/developments/lianmi/work/logs" > $env_file


# 服务器的MySQL命令行 
~/.zshrc
```
#mysql
alias mysql="docker-compose exec db mysql -ulianmidba -p12345678 lianmicloud"

```

因此可以这样：
```
$  cd /root/developments/lianmi/work/basic
$  mysql
```

# 服务器的rdis命令行 
~/.zshrc
```
#redis
alias redis-cli="docker exec -it redis redis-cli"

```

因此可以这样：
```
$  cd /root/developments/lianmi/work/basic
$  redis-cli
```
就能进入redis-cli

# 文档 
```
https://github.com/lianmi/docs
```
gitbook插件
https://www.jianshu.com/p/427b8bb066e6

# ca

```
https://github.com/do-know/Crypt-LE
https://hub.docker.com/r/zerossl/client/
```
docker
```
docker pull zerossl/client

alias le.pl='docker run -it -v /root/developments/lianmi/work/keys_and_certs:/data -v /home/my_user/public_html/.well-known/acme-challenge:/webroot -u $(id -u) --rm zerossl/client'

证书存放目录:
/root/developments/lianmi/work/keys_and_certs
```

生成key及签名:
```
le.pl --key account.key --csr domain.csr --csr-key domain.key --crt domain.crt --domains "mqtt.lianmi.cloud,api.lianmi.cloud" --generate-missing --path /webroot --unlink
```

生成证书:

```
cd /root/developments/lianmi/work/keys_and_certs
openssl req -new -x509 -sha256 -key domain.key -out domain.crt -days 3650
```

# 区块链 blockchain
开发目录：
## 1. 发币合约
```
/Users/mac/developments/lianmi/blockchain/ethereum-simplenode-erc20/erc20_demo
```
## 2. 多签合约
```
/Users/mac/developments/lianmi/blockchain/ethereum-simplenode-erc20/erc20_multisig
```

# Mac 部署
## ethereum目录

```
/Users/mac/developments/lianmi/blockchain/ethereum-poa

```

## 参考mac.md



#  腾讯云服务器部署 
##   geth 参数详解 
https://www.cnblogs.com/wanghui-garcia/p/10256520.html

## 双节点
### 1. 初始化
```
$ cd /store/blockchain/lianmichain
$ 
```

### 2. 创建新账号
node1 & node2
```
$ geth --datadir node1/ account new
输入两次 LianmiSky8900388
输出：

Public address of the key:   0x614075e0853a1d34f85120758A7dfA9316Ab9109
Path of the secret key file: node1/keystore/UTC--2020-10-09T10-53-24.039679171Z--614075e0853a1d34f85120758a7dfa9316ab9109

$ geth --datadir node2/ account new

输入两次 LianmiSky8900388
输出：

Public address of the key:   0x30D199128ad70cBD0ccC1b0cc69296C523aFF894
Path of the secret key file: node2/keystore/UTC--2020-10-09T10-54-24.533221434Z--30d199128ad70cbd0ccc1b0cc69296c523aff894

```

## 部署及运行
### 生成创世文件
```
$ puppeth
``` 

##  部署节点1，2创世
```
rm  -rf  ./node1/geth
geth --datadir node1/ init lianmichain.json

rm  -rf  ./node2/geth
geth --datadir node2/ init lianmichain.json
```

### 新开一个终端，  bootnode 服务 
```
$ bootnode -genkey boot.key
$ nohup bootnode -nodekey boot.key -verbosity 9 -addr :30310  >/dev/null 2>&1 &
输出 ：
enode://3331ac1ea468c46fb336ed96c0c2be4066fa3592459baf882dab193157f9148f15dfdc335d3ed7522a225a0431ba3b473a90602cb3dfb80d47c8157981db4cc1@127.0.0.1:0?discport=30310
Note: you're using cmd/bootnode, a developer tool.
We recommend using a regular node as bootstrap node for production deployments.
INFO [10-09|19:01:05.821] New local node record                    seq=1 id=441c8e31af296c8d ip=<nil> udp=0 tcp=0

```


###  运行 节点1
```
$ nohup geth --datadir node1/ --syncmode 'full' --mine --port 30311  --bootnodes 'enode://3331ac1ea468c46fb336ed96c0c2be4066fa3592459baf882dab193157f9148f15dfdc335d3ed7522a225a0431ba3b473a90602cb3dfb80d47c8157981db4cc1@127.0.0.1:30310' --networkid 1 --gasprice '1' -unlock '614075e0853a1d34f85120758a7dfa9316ab9109' --password node1/password.txt --ws --wsaddr 0.0.0.0 --wsport 8546 --wsorigins '*' --wsapi personal,admin,eth,net,web3,miner,txpool,debug --allow-insecure-unlock  >/dev/null 2>&1 &
 ```
 
###  运行 节点2
```
$ nohup geth --datadir node2/ --syncmode 'full' --mine --port 30312 --bootnodes 'enode://3331ac1ea468c46fb336ed96c0c2be4066fa3592459baf882dab193157f9148f15dfdc335d3ed7522a225a0431ba3b473a90602cb3dfb80d47c8157981db4cc1@127.0.0.1:30310' --networkid 1 --gasprice '1' --unlock '30d199128ad70cbd0ccc1b0cc69296c523aff894' --password node2/password.txt --ws --wsaddr 0.0.0.0 --wsport 8547 --wsorigins '*' --wsapi personal,admin,eth,net,web3,miner,txpool,debug --allow-insecure-unlock  >/dev/null 2>&1 &
 ```


 # console 控制台 
 ```
$ geth attach ipc:/store/blockchain/lianmichain/node1/geth.ipc
 ```
 ## 授权挖矿
 ```
 > clique.getSnapshot()
{
  hash: "0x125f7363bb26e8d8674c3be1520a30a3137a9a1e81bf1e0b85bc8583b03022e8",
  number: 1565,
  recents: {
    1564: "0xa7563c330c5285721632189fc6644fde324dae54",
    1565: "0xf1de15bb2cf24038d1b986515d5fe55e4eb3052d"
  },
  signers: {
    0xa7563c330c5285721632189fc6644fde324dae54: {},
    0xf1de15bb2cf24038d1b986515d5fe55e4eb3052d: {}
  },
  tally: {},
  votes: []
}

 //授权node1  (？？)
 > clique.propose("0xf1de15bb2cf24038d1b986515d5fe55e4eb3052d",true)

 //授权node2 (？？)
 > clique.propose("0xa7563c330c5285721632189fC6644FdE324dae54",true)
 ```

 ## 向HD钱包的子地址转账
 ```
> account1 = web3.eth.coinbase
> web3.eth.getBalance(account1)
> web3.fromWei(web3.eth.getBalance(account1), 'ether')

> leaf0 = '0xe14D151e0511b61357DDe1B35a74E9c043c34C47'
>  web3.fromWei(web3.eth.getBalance(leaf0), 'ether')
> eth.sendTransaction({from:account1,to:leaf0,value:web3.toWei(10000000000,"ether")})

> web3.fromWei(web3.eth.getBalance(leaf0), 'ether')
10000000000

> leaf1 = '0x4acea697f366C47757df8470e610a2d9B559DbBE'
> eth.sendTransaction({from:account1,to:leaf1,value:web3.toWei(10000000000,"ether")})
> web3.fromWei(web3.eth.getBalance(leaf1), 'ether')
10000000000

> leaf2 = '0x9DEb6E226b84b21b354cCa634e4867C6F7A0f77c'
> eth.sendTransaction({from:account1,to:leaf2,value:web3.toWei(10000000000,"ether")})
> web3.fromWei(web3.eth.getBalance(leaf2), 'ether')
10000000000

 ```

 ## 查询链上的交易哈希数据
 ```

 > eth.getTransactionReceipt("0xcc96f63d1013244962b7eaa511b232148a01e79da90f4e94450e7c7bfe6e4ec0")

 ```


 ## 发币 LNMC 10000亿枚
 
 ### 部署发币智能合约，并生成合约地址
```

```

 ### 将生成的发币合约地址保存到配置
 ```
 ```

 ## TODO

 我觉得 交易完成后 这个订单 的 attach 服务端进行一个hash 计算 保存好最后的 一次确定交易 hash , 到时候查证的时候 , 可以通过这个hash 知道 内容有没有篡改
