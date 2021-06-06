package huawei

import (
	"fmt"
	"net/http"

	LMCommon "github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/pkg/pushapi/httputil"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	// golog "log"
	"time"
)

// HMS Core > 华为帐号服务 > API参考 基于OAuth 2.0开放鉴权 客户端模式（Client Credentials）
// https://developer.huawei.com/consumer/cn/doc/development/HMSCore-Guides-V5/open-platform-oauth-0000001053629189-V5#ZH-CN_TOPIC_0000001053629189__section12493191334711
type AuthReq struct {
	GrantType    string `json:"grant_type,omitempty"`    // 填写为“client_credentials”，表示为客户端模式。
	ClientId     string `json:"client_id,omitempty"`     // 在接入前准备中得到的OAuth 2.0客户端ID，对于AppGallery Connect类应用，该值为应用的APP ID。
	ClientSecret string `json:"client_secret,omitempty"` // 在接入前准备中给客户端ID分配的密钥，对于AppGallery Connect类应用，该值为应用的APP SECRET。
}

type AuthRes struct {
	AccessToken string `json:"access_token"` // 应用级Access Token。
	ExpiresIn   int    `json:"expires_in"`   // Access Token的剩余有效期，单位：秒。3600 秒
	TokenType   string `json:"token_type"`   // 固定返回Bearer，标识返回Access Token的类型。
}

type WorkerJob interface {
	Run()
	Name() string
	run() error
}

type InitDataJob struct {
	manage *HuaweiManage
	name   string
}

func initDataJob(manage *HuaweiManage) *InitDataJob {
	j := new(InitDataJob)
	j.manage = manage
	j.name = manage.GetName() + " init data"
	return j
}

func (j *InitDataJob) Run() {
	_ = j.run()
}

func (j *InitDataJob) Name() string {
	return j.name
}

func (j *InitDataJob) run() error {
	startTime := time.Now()
	j.manage.logger.Debug(fmt.Sprintf("job:%s startTime:%s", j.name, startTime.String()))
	defer func() {
		j.manage.logger.Debug(fmt.Sprintf("job:%s endtime:%s cost:%s", j.name, time.Now().String(), time.Now().Sub(startTime).String()))
	}()

	j.manage.logger.Debug(fmt.Sprintf("job run  "))
	return nil
}

type RefreshHuaweiAssessTokenJob struct {
	jobId  cron.EntryID
	spec   string
	manage *HuaweiManage
	name   string
}

func initRefreshHuaweiAssessTokenJob(manage *HuaweiManage,
	spec string) (j *RefreshHuaweiAssessTokenJob, err error) {

	tmp := new(RefreshHuaweiAssessTokenJob)
	tmp.manage = manage
	tmp.name = manage.GetName() + " refresh access_token"
	tmp.spec = spec
	j = tmp
	tmp.jobId, err = manage.cronWorker.crontab.AddJob(spec, j)
	if err != nil {
		j = nil
		return
	}

	return
}

func (j *RefreshHuaweiAssessTokenJob) Run() {

	j.manage.cronWorker.crontab.Remove(j.jobId)
	defer func() {
		j.jobId, _ = j.manage.cronWorker.crontab.AddJob(j.spec, j)
	}()

	_ = j.run()
}

func (j *RefreshHuaweiAssessTokenJob) Name() string {
	return j.name
}

//job主体逻辑
func (j *RefreshHuaweiAssessTokenJob) run() (err error) {

	startTime := time.Now()
	j.manage.logger.Debug(fmt.Sprintf("job:%s startTime:%s", j.name, startTime.String()))
	j.manage.logger.Debug(fmt.Sprintf("printf job:%s starttime:%s", j.name, startTime.String()))
	defer func() {
		j.manage.logger.Debug(fmt.Sprintf("job:%s endtime:%s cost:%s", j.name, time.Now().String(), time.Now().Sub(startTime).String()))
	}()

	req := &AuthReq{
		GrantType:    "client_credentials",
		ClientId:     LMCommon.HuaweiAppId,
		ClientSecret: LMCommon.HuaweiAppSecret,
	}
	res := &AuthRes{}

	params := httputil.StructToUrlValues(req)
	code, resBody, err := httputil.PostForm(LMCommon.HuaweiAuthURL, params, res, nil)
	if err != nil {
		return fmt.Errorf("code=%d body=%s err=%v", code, resBody, err)
	}

	if code != http.StatusOK || res.AccessToken == "" {
		return fmt.Errorf("code=%d body=%s", code, resBody)
	}

	authToken := fmt.Sprintf("%s %s", res.TokenType, res.AccessToken)

	redisConn := j.manage.redisPool.Get()
	defer redisConn.Close()
	_, err = redisConn.Do("SET", "HuaweiAuthToken", authToken)
	if err != nil {
		j.manage.logger.Error("获取华为access_token失败", zap.Error(err))
		return err
	}

	j.manage.logger.Debug("获取华为access_token成功", zap.String("authToken", authToken))
	return nil
}
