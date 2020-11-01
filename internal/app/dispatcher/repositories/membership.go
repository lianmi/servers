package repositories

import (
	// "encoding/json"
	"fmt"
	// "time"
	// "net/http"

	// "github.com/golang/protobuf/proto"
	"github.com/gomodule/redigo/redis"
	// "github.com/jinzhu/gorm"
	// Auth "github.com/lianmi/servers/api/proto/auth"
	Auth "github.com/lianmi/servers/api/proto/auth"
	// User "github.com/lianmi/servers/api/proto/user"
	// "github.com/lianmi/servers/internal/app/dispatcher/grpcclients"
	// "github.com/lianmi/servers/internal/app/dispatcher/nsqMq"
	// LMCommon "github.com/lianmi/servers/internal/common"
	// "github.com/lianmi/servers/internal/pkg/models"
	// "github.com/pkg/errors"
	"go.uber.org/zap"
	// "github.com/lianmi/servers/util/dateutil"
)

//TODO
//商户查询当前名下用户总数，按月统计付费会员总数及返佣金额，是否已经返佣
func (s *MysqlLianmiRepository) GetBusinessMembership(isRebate bool) (*Auth.GetBusinessMembershipResp, error) {
	var err error

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	return nil, err

}

//TODO
func (s *MysqlLianmiRepository) PayForMembership(payForUsername string) error {

	//支付完成后，需要向商户，支付者，会员获得者推送系统通知
	//构建数据完成，向dispatcher发送
	return nil
}

//预生成一个购买会员的订单， 返回OrderID及预转账裸交易数据
func (s *MysqlLianmiRepository) PreOrderForPayMembership(username, deviceID string) error {

	var err error

	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	//取出当前设备的os， clientType， logonAt
	curDeviceHashKey := fmt.Sprintf("devices:%s:%s", username, deviceID)
	isMaster, _ := redis.Bool(redisConn.Do("HGET", curDeviceHashKey, "ismaster"))
	curOs, _ := redis.String(redisConn.Do("HGET", curDeviceHashKey, "os"))
	curClientType, _ := redis.Int(redisConn.Do("HGET", curDeviceHashKey, "clientType"))
	curLogonAt, _ := redis.Uint64(redisConn.Do("HGET", curDeviceHashKey, "logonAt"))

	s.logger.Debug("HandlePreTransfer",
		zap.Bool("isMaster", isMaster),
		zap.String("username", username),
		zap.String("deviceID", deviceID),
		zap.String("curOs", curOs),
		zap.Int("curClientType", curClientType),
		zap.Uint64("curLogonAt", curLogonAt))

	_ = err
	return nil
}
