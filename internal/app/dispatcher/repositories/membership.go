package repositories

import (
	// "encoding/json"
	// "fmt"
	// "time"

	// "github.com/golang/protobuf/proto"
	// "github.com/gomodule/redigo/redis"
	// "github.com/jinzhu/gorm"
	// Auth "github.com/lianmi/servers/api/proto/auth"
	Service "github.com/lianmi/servers/api/proto/service"
	// User "github.com/lianmi/servers/api/proto/user"
	// "github.com/lianmi/servers/internal/app/dispatcher/grpcclients"
	// "github.com/lianmi/servers/internal/app/dispatcher/nsqMq"
	// "github.com/lianmi/servers/internal/common"
	// "github.com/lianmi/servers/internal/pkg/models"
	// "github.com/pkg/errors"
	// "go.uber.org/zap"
	// "github.com/lianmi/servers/util/dateutil"
)

//TODO
//商户查询当前名下用户总数，按月统计付费会员总数及返佣金额，是否已经返佣
func (s *MysqlLianmiRepository) GetBusinessMembership(isRebate bool) (*Service.GetBusinessMembershipResp, error) {
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
