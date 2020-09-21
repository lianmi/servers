module github.com/lianmi/servers

go 1.12

require (
	github.com/aliyun/aliyun-oss-go-sdk v2.1.4+incompatible
	github.com/appleboy/gin-jwt/v2 v2.6.3
	github.com/bingjian-zhu/gin-vue-admin v0.0.0-20200506131022-dcbf95f91663
	github.com/bitly/go-simplejson v0.5.0
	github.com/btcsuite/btcutil v0.0.0-20190425235716-9e5f4b9a998d
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/confluentinc/confluent-kafka-go v1.4.2 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/eclipse/paho.golang v0.9.0
	github.com/ethereum/go-ethereum v1.9.21
	github.com/gin-contrib/pprof v1.2.0
	github.com/gin-contrib/zap v0.0.0-20190528085758-3cc18cd8fce3
	github.com/gin-gonic/gin v1.4.0
	github.com/go-playground/validator/v10 v10.3.0
	github.com/golang/protobuf v1.4.2
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/wire v0.3.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-opentracing v0.0.0-20180507213350-8e809c8a8645
	github.com/hashicorp/consul/api v1.3.0
	github.com/hashicorp/go-msgpack v0.5.5 // indirect
	github.com/hashicorp/memberlist v0.1.4 // indirect
	github.com/ipfs/go-ipfs-api v0.1.0
	github.com/jinzhu/gorm v1.9.11
	github.com/mbobakov/grpc-consul-resolver v1.4.1
	github.com/miguelmota/go-ethereum-hdwallet v0.0.0-20200123000308-a60dcd172b4c
	github.com/opentracing-contrib/go-gin v0.0.0-20190301172248-2e18f8b9c7d4
	github.com/opentracing/opentracing-go v1.1.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.4.0
	github.com/robfig/cron v1.2.0
	github.com/satori/go.uuid v1.2.0
	github.com/shopspring/decimal v1.2.0
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.5.1
	github.com/uber-go/atomic v1.4.0 // indirect
	github.com/uber/jaeger-client-go v2.16.0+incompatible
	github.com/uber/jaeger-lib v2.0.0+incompatible
	github.com/unrolled/secure v1.0.8
	go.uber.org/zap v1.12.0
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/tools v0.0.0-20200903185744-af4cc2cd812e // indirect
	google.golang.org/grpc v1.30.0
	google.golang.org/protobuf v1.25.0
	gopkg.in/confluentinc/confluent-kafka-go.v1 v1.4.2
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v2 v2.3.0
)

replace github.com/hashicorp/consul => github.com/hashicorp/consul v1.5.1

replace github.com/ugorji/go v1.1.4 => github.com/ugorji/go v0.0.0-20190204201341-e444a5086c43

replace github.com/lianmi/servers => /Users/mac/developments/goprojects/src/lianmi/servers

replace github.com/eclipse/paho.golang v0.9.0 => github.com/eclipse/paho.golang v0.9.1-0.20200717101128-7369e711591a
