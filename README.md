这是连米信息科技后端微服务整体框架项目，囊括了项目的结构、分层思想、依赖注入、错误处理、单元测试、服务治理、框架选择等方面.

项目分为dispatcher、authservice、singlechatservice、groupchatservice、e-commerceservice、区块链认证的下单小程序、lmc公链(ChinkLink)、FileCoin区块链，收银称重一体系统等微服务。


## 准备

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

### 平台 linux && mac
```
	do \
	    GOOS=linux GOARCH="amd64" go build -o dist/$$app-linux-amd64 ./cmd/$$app/; \
		GOOS=darwin GOARCH="amd64" go build -o dist/$$app-darwin-amd64 ./cmd/$$app/; \
	done
```


### 运行, 以dispatcher为例
```
$ ./dist/dispatcher-darwin-amd64 -f ./configs/dispatcher.yml
```

* **访问接口**： http://localhost:28080/auth/v1/register
* **consul**: http://localhost:8500/
* **grafana**: http://localhost:3000/ 
* **jaeger**: http://localhost:16686/search
* **Prometheus**: http://localhost:9090/graph
* **AlertManager**: http://localhost:9093



### Grafana Dashboard,可以自动生成!


### Prometheus Alert 监控告警,自动生成！

### 调用链跟踪

## 开发文档及相关协议约定
```
cd doc
gitbook 
```


## 使用eclipse paho.golang master版本
```
go mod edit -replace=github.com/eclipse/paho.golang@v0.9.0=github.com/eclipse/paho.golang@master
```

## proto文件编译
```
make proto

```

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

## minio文件服务器
Android客户端 -> Minio -> ipfs

方法： 将Minio作为服务端运行，当Android发送完成后，通过mqtt发送到dispatcher，再由miniouploader进行ipfs上链操作
```
/Users/mac/developments/lianmi/ipfs/minio-client/file-uploader.go
```

### ipfs client
```

```