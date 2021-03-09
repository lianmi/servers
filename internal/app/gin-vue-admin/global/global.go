package global

import (
	"go.uber.org/zap"

	"github.com/go-redis/redis"
	"github.com/lianmi/servers/internal/app/gin-vue-admin/config"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var (
	GVA_DB     *gorm.DB
	GVA_REDIS  *redis.Client
	GVA_CONFIG config.Server
	GVA_VP     *viper.Viper
	//GVA_LOG    *oplogging.Logger
	GVA_LOG *zap.Logger

	LIANMI_DB *gorm.DB //连米mysql
)
