package huawei

import (
	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
)

type HuaweiManage struct {
	logger     *zap.Logger
	redisPool  *redis.Pool
	name       string
	url        string //huawei get access_token url
	cronWorker *CronWorker
}

func NewHuaweiManage(logger *zap.Logger, redisPool *redis.Pool, name, huaweiUrl string) (*HuaweiManage, error) {
	var err error
	manage := &HuaweiManage{}
	manage.logger = logger
	manage.redisPool = redisPool
	manage.name = name
	manage.url = huaweiUrl
	// 初始化就获取一次，保存到redis

	manage.cronWorker = NewCronWorker(manage)
	err = manage.cronWorker.Prepare()
	if err != nil {
		// log.DetailError(err)
		return nil, err
	}

	_ = err
	return manage, nil

}

func (manage *HuaweiManage) GetName() string {
	return manage.name
}

func (s *HuaweiManage) Start() error {

	return s.cronWorker.Start()

}

func (s *HuaweiManage) Stop() {
	s.cronWorker.Stop()
}
