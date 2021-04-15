package controllers

import (
	"github.com/lianmi/servers/internal/app/dispatcher/nsqMq"
	"github.com/lianmi/servers/internal/app/dispatcher/services"
	"go.uber.org/zap"
)

type LianmiApisController struct {
	logger    *zap.Logger
	service   services.LianmiApisService
	nsqClient *nsqMq.NsqClient       //nsqMq
	cacheMap  map[string]interface{} // 内存缓存信息map , 用语缓存一些常用不变的值
}

func NewLianmiApisController(logger *zap.Logger, s services.LianmiApisService, nsqClient *nsqMq.NsqClient) *LianmiApisController {
	return &LianmiApisController{
		logger:    logger,
		service:   s,
		nsqClient: nsqClient,
		cacheMap:  make(map[string]interface{}),
	}
}
