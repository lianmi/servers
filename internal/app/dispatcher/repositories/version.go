package repositories

import (
	"github.com/gomodule/redigo/redis"

	// Global "github.com/lianmi/servers/api/proto/global"

	// uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
)

//获取redis里App version
func (s *MysqlLianmiRepository) GetAppVersion(oldVersion string) (string, error) {
	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	curVersion, err := redis.String(redisConn.Do("GET", "AppVersionLast"))

	if err != nil {
		s.logger.Error("ManagerSetVersionLast", zap.Error(err))
		return "", err
	}

	return curVersion, nil
}
