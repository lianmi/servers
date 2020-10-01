这是连米信息科技后端微服务整体框架项目，囊括了项目的结构、分层思想、依赖注入、错误处理、单元测试、服务治理、框架选择等方面.

项目分为dispatcher、authservice、orderservice、chatservice、walletservice等微服务。
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
