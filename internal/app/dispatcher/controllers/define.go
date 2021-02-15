package controllers

import (
	"github.com/lianmi/servers/internal/app/dispatcher/nsqMq"
	"github.com/lianmi/servers/internal/app/dispatcher/services"
	"go.uber.org/zap"
)

type LianmiApisController struct {
	logger    *zap.Logger
	service   services.LianmiApisService
	nsqClient *nsqMq.NsqClient //nsqMq
}

func NewLianmiApisController(logger *zap.Logger, s services.LianmiApisService, nsqClient *nsqMq.NsqClient) *LianmiApisController {
	return &LianmiApisController{
		logger:    logger,
		service:   s,
		nsqClient: nsqClient,
	}
}
