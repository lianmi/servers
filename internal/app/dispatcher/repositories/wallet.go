package repositories

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/lianmi/servers/internal/pkg/models"
)

func (s *MysqlLianmiRepository) GetAlipayInfoByTradeNo(outTradeNo string) (string, float64, bool, error) {
	redisConn := s.redisPool.Get()
	defer redisConn.Close()

	preAlipayKey := fmt.Sprintf("PreAlipay:%s", outTradeNo)

	//获取username
	username, err := redis.String(redisConn.Do("HGET", preAlipayKey, "Username"))

	//获取充值金额
	totalAmount, err := redis.Float64(redisConn.Do("HGET", preAlipayKey, "TotalAmount"))

	//获取充值状态
	IsPayed, err := redis.Bool(redisConn.Do("HGET", preAlipayKey, "IsPayed"))

	return username, totalAmount, IsPayed, err
}

func (s *MysqlLianmiRepository) GetChargeProductID() (string, error) {
	var systemCharge models.SystemCharge
	if err := s.db.Model(&models.SystemCharge{}).First(&systemCharge).Error; err != nil {
		return "", err
	}

	return systemCharge.ChargeProductID, nil
}
