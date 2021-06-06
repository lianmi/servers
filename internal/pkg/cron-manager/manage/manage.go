package manage

import (
	"errors"
	"fmt"

	"github.com/gomodule/redigo/redis"
	"github.com/lianmi/servers/internal/app/dispatcher/services"
	"github.com/lianmi/servers/internal/pkg/cron-manager/manage/huawei"
	"github.com/lianmi/servers/internal/pkg/cron-manager/manage/imcron"
	"go.uber.org/zap"
)

const (
	HuaweiUrl = "https://oauth-login.cloud.huawei.com/oauth2/v3/token"
)

var g_manage *Manage = nil

type Manage struct {
	huaweiManage *huawei.HuaweiManage
	imcronManage *imcron.IMCronManage
}

func newManage(
	service services.LianmiApisService,
	logger *zap.Logger,
	redisPool *redis.Pool,
) (man *Manage, err error) {
	huaweiManage, err := huawei.NewHuaweiManage(logger,redisPool, "huawei", HuaweiUrl)
	if err != nil {
		return
	}
	imcronManage, err := imcron.NewIMCronManage(service, logger, redisPool, "imcron")
	man = &Manage{
		huaweiManage: huaweiManage,
		imcronManage: imcronManage,
	}

	return
}

func Run(
	service services.LianmiApisService,
	logger *zap.Logger,
	redisPool *redis.Pool,
) (err error) {

	if g_manage != nil {
		err = fmt.Errorf("manage has init")
		return
	}

	man, err := newManage(service, logger, redisPool)
	if err != nil {
		return
	}

	err = man.Start()
	if err == nil {
		g_manage = man
	}
	return err
}

func (m *Manage) Start() error {

	if m.huaweiManage != nil {

		_ = m.huaweiManage.Start()
	} else {
		// log.DetailError("m.huaweiManage is nil")
		return errors.New("m.huaweiManage is nil")
	}

	if m.imcronManage != nil {

		_ = m.imcronManage.Start()
	} else {
		// log.DetailError("m.imcronManage is nil")
		return errors.New("m.imcronManage is nil")
	}

	return nil

}

func (m *Manage) Stop() error {

	if m.huaweiManage != nil {
		m.huaweiManage.Stop()
	}
	if m.imcronManage != nil {
		m.imcronManage.Stop()
	}

	return nil
}

func GetHuaweiManage() *huawei.HuaweiManage {
	if g_manage == nil {
		panic("p_manage null")
	}
	return g_manage.huaweiManage
}

func GetIMCronManage() *imcron.IMCronManage {
	if g_manage == nil {
		panic("p_manage null")
	}
	return g_manage.imcronManage
}

func DestroyManage() error {
	if g_manage == nil {
		return fmt.Errorf("p_manage null")
	}
	g_manage.Stop()
	g_manage = nil
	return nil

}
