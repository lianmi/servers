package controllers

import (
	"github.com/lianmi/servers/internal/app/dispatcher/services"
	"go.uber.org/zap"
)

type LianmiApisController struct {
	logger  *zap.Logger
	service services.LianmiApisService
}

func NewLianmiApisController(logger *zap.Logger, s services.LianmiApisService) *LianmiApisController {
	return &LianmiApisController{
		logger:  logger,
		service: s,
	}
}
