package imcron

import (
	"github.com/gomodule/redigo/redis"
	"github.com/lianmi/servers/internal/app/dispatcher/services"
	"go.uber.org/zap"
)

type IMCronManage struct {
	service services.LianmiApisService

	logger     *zap.Logger
	redisPool  *redis.Pool
	name       string
	url        string //huawei get access_token url
	cronWorker *CronWorker
}

func NewIMCronManage(
	service services.LianmiApisService,
	logger *zap.Logger,
	redisPool *redis.Pool,
	name string,
) (*IMCronManage, error) {
	var err error
	manage := &IMCronManage{}
	manage.service = service
	manage.logger = logger
	manage.redisPool = redisPool
	manage.name = name
	manage.cronWorker = NewCronWorker(manage)

	err = manage.cronWorker.Prepare()
	if err != nil {
		return nil, err
	}

	_ = err
	return manage, nil

}

func (manage *IMCronManage) GetName() string {
	return manage.name
}

func (s *IMCronManage) Start() error {

	return s.cronWorker.Start()

}

func (s *IMCronManage) Stop() {
	s.cronWorker.Stop()
}
