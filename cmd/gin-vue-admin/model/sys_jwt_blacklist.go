package model

import (
	"github.com/lianmi/servers/cmd/gin-vue-admin/global"
)

type JwtBlacklist struct {
	global.GVA_MODEL
	Jwt string `gorm:"type:text;comment:jwt"`
}
