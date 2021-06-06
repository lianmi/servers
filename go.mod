module github.com/lianmi/servers

go 1.14

require (
	github.com/360EntSecGroup-Skylar/excelize/v2 v2.3.2
	github.com/AlecAivazis/survey/v2 v2.1.1
	github.com/aliyun/aliyun-oss-go-sdk v2.1.4+incompatible
	github.com/appleboy/gin-jwt/v2 v2.6.3
	github.com/baiyubin/aliyun-sts-go-sdk v0.0.0-20180326062324-cfa1a18b161f // indirect
	github.com/bitly/go-simplejson v0.5.0
	github.com/bmizerany/assert v0.0.0-20160611221934-b7ed37b82869 // indirect
	github.com/codahale/hdrhistogram v0.0.0-20161010025455-3a0bb77429bd // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/eclipse/paho.golang v0.9.0
	github.com/ethereum/go-ethereum v1.9.21
	github.com/gin-contrib/pprof v1.2.0
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/gin-contrib/zap v0.0.0-20190528085758-3cc18cd8fce3
	github.com/gin-gonic/gin v1.4.0
	github.com/go-playground/validator/v10 v10.3.0
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/mock v1.4.4 // indirect
	github.com/golang/protobuf v1.4.2
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/google/wire v0.3.0
	github.com/gopherjs/gopherjs v0.0.0-20190812055157-5d271430af9f // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/grpc-ecosystem/grpc-opentracing v0.0.0-20180507213350-8e809c8a8645
	github.com/hashicorp/consul/api v1.3.0
	github.com/hashicorp/go-msgpack v0.5.5 // indirect
	github.com/hashicorp/go-multierror v1.1.0 // indirect
	github.com/hashicorp/memberlist v0.1.4 // indirect
	github.com/hpcloud/tail v1.0.0
	github.com/iGoogle-ink/gopay v1.5.38
	github.com/ipfs/go-ipfs-api v0.1.0
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/kr/pretty v0.2.0 // indirect
	github.com/kr/pty v1.1.8 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lestrrat-go/file-rotatelogs v2.4.0+incompatible
	github.com/lestrrat-go/strftime v1.0.4 // indirect
	github.com/libp2p/go-libp2p-core v0.8.0 // indirect
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/mbobakov/grpc-consul-resolver v1.4.1
	github.com/miekg/dns v1.1.31 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/multiformats/go-multiaddr v0.3.1 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/nsqio/go-nsq v1.0.8
	github.com/opentracing-contrib/go-gin v0.0.0-20190301172248-2e18f8b9c7d4
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.7.1
	github.com/robfig/cron/v3 v3.0.1
	github.com/satori/go.uuid v1.2.0
	github.com/shopspring/decimal v1.2.0
	github.com/smartystreets/assertions v1.0.1 // indirect
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.7.0
	github.com/uber-go/atomic v1.4.0 // indirect
	github.com/uber/jaeger-client-go v2.16.0+incompatible
	github.com/uber/jaeger-lib v2.0.0+incompatible
	github.com/ugorji/go v1.1.13 // indirect
	github.com/unrolled/secure v1.0.8
	github.com/whyrusleeping/tar-utils v0.0.0-20201201191210-20a61371de5b // indirect
	go.opencensus.io v0.22.5 // indirect
	go.uber.org/zap v1.16.0
	golang.org/x/crypto v0.0.0-20210415154028-4f45737414dc
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/mod v0.4.1 // indirect
	golang.org/x/tools v0.1.0 // indirect
	google.golang.org/grpc v1.31.1
	google.golang.org/protobuf v1.25.0
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
	gorm.io/driver/mysql v1.0.3
	gorm.io/gorm v1.20.7

)

replace github.com/hashicorp/consul => github.com/hashicorp/consul v1.5.1

replace github.com/ugorji/go v1.1.4 => github.com/ugorji/go v0.0.0-20190204201341-e444a5086c43

// replace github.com/eclipse/paho.golang v0.9.0 => github.com/eclipse/paho.golang v0.9.1-0.20200717101128-7369e711591a
replace github.com/eclipse/paho.golang v0.9.0 => github.com/lianmi/paho.golang v0.9.2

replace github.com/casbin/gorm-adapter/v3 => github.com/casbin/gorm-adapter/v3 v3.0.2
