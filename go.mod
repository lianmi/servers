module github.com/lianmi/servers

go 1.12

require (
	github.com/AlecAivazis/survey/v2 v2.1.1
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751
	github.com/aliyun/aliyun-oss-go-sdk v2.1.4+incompatible
	github.com/allegro/bigcache v1.2.1 // indirect
	github.com/appleboy/gin-jwt/v2 v2.6.3
	github.com/aristanetworks/goarista v0.0.0-20190912214011-b54698eaaca6 // indirect
	github.com/baiyubin/aliyun-sts-go-sdk v0.0.0-20180326062324-cfa1a18b161f // indirect
	github.com/bitly/go-simplejson v0.5.0
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/btcsuite/btcutil v0.0.0-20190425235716-9e5f4b9a998d
	github.com/casbin/casbin v1.7.0
	github.com/casbin/casbin/v2 v2.19.8
	github.com/casbin/gorm-adapter/v3 v3.0.4
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/eclipse/paho.golang v0.9.0
	github.com/edsrzf/mmap-go v1.0.0 // indirect
	github.com/ethereum/go-ethereum v1.9.21
	github.com/fastly/go-utils v0.0.0-20180712184237-d95a45783239 // indirect
	github.com/fsnotify/fsnotify v1.4.9
	github.com/fvbock/endless v0.0.0-20170109170031-447134032cb6
	github.com/gin-contrib/pprof v1.2.0
	github.com/gin-contrib/zap v0.0.0-20190528085758-3cc18cd8fce3
	github.com/gin-gonic/gin v1.4.0
	github.com/go-playground/validator/v10 v10.3.0
	github.com/go-redis/redis v6.14.2+incompatible
	github.com/go-sql-driver/mysql v1.5.0
	github.com/golang/protobuf v1.4.2
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/uuid v1.2.0 // indirect
	github.com/google/wire v0.3.0
	github.com/gookit/color v1.3.6
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-opentracing v0.0.0-20180507213350-8e809c8a8645
	github.com/hashicorp/consul/api v1.3.0
	github.com/hashicorp/go-msgpack v0.5.5 // indirect
	github.com/hashicorp/memberlist v0.1.4 // indirect
	github.com/hpcloud/tail v1.0.0
	github.com/ipfs/go-ipfs-api v0.1.0
	github.com/jehiah/go-strftime v0.0.0-20171201141054-1d33003b3869 // indirect
	github.com/jordan-wright/email v4.0.1-0.20201230161404-3656afaabac4+incompatible
	github.com/lestrrat-go/file-rotatelogs v2.4.0+incompatible
	github.com/lestrrat-go/strftime v1.0.3 // indirect
	github.com/mbobakov/grpc-consul-resolver v1.4.1
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mojocn/base64Captcha v1.3.1
	github.com/nsqio/go-nsq v1.0.8
	github.com/opentracing-contrib/go-gin v0.0.0-20190301172248-2e18f8b9c7d4
	github.com/opentracing/opentracing-go v1.1.0
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.4.0
	github.com/qiniu/api.v7/v7 v7.8.0
	github.com/rjeczalik/notify v0.9.2 // indirect
	github.com/robfig/cron v1.2.0
	github.com/rs/cors v1.7.0 // indirect
	github.com/satori/go.uuid v1.2.0
	github.com/shirou/gopsutil v2.20.5+incompatible
	github.com/shopspring/decimal v1.2.0
	github.com/smartwalle/alipay/v3 v3.1.5
	github.com/smartwalle/xid v1.0.6
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.7.0
	github.com/swaggo/gin-swagger v1.3.0
	github.com/swaggo/swag v1.5.1
	github.com/tebeka/strftime v0.1.5 // indirect
	github.com/tyler-smith/go-bip39 v1.0.1-0.20181017060643-dbb3b84ba2ef
	github.com/uber-go/atomic v1.4.0 // indirect
	github.com/uber/jaeger-client-go v2.16.0+incompatible
	github.com/uber/jaeger-lib v2.0.0+incompatible
	github.com/unrolled/secure v1.0.8
	go.uber.org/zap v1.16.0
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/mod v0.4.1 // indirect
	golang.org/x/net v0.0.0-20210220033124-5f55cee0dc0d // indirect
	golang.org/x/sys v0.0.0-20210220050731-9a76102bfb43 // indirect
	golang.org/x/tools v0.0.0-20200903185744-af4cc2cd812e // indirect
	google.golang.org/grpc v1.30.0
	google.golang.org/protobuf v1.25.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gorm.io/driver/mysql v1.0.3
	gorm.io/gorm v1.20.7
)

replace github.com/hashicorp/consul => github.com/hashicorp/consul v1.5.1

replace github.com/ugorji/go v1.1.4 => github.com/ugorji/go v0.0.0-20190204201341-e444a5086c43

// replace github.com/lianmi/servers => /Users/mac/developments/goprojects/src/lianmi/servers

// replace github.com/eclipse/paho.golang v0.9.0 => github.com/eclipse/paho.golang v0.9.1-0.20200717101128-7369e711591a
replace github.com/eclipse/paho.golang v0.9.0 => github.com/lianmi/paho.golang v0.9.2

replace github.com/casbin/gorm-adapter/v3 => github.com/casbin/gorm-adapter/v3 v3.0.2
