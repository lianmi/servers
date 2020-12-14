package nsqMq

import (
	"fmt"
	// "github.com/aliyun/aliyun-oss-go-sdk/oss"
	simpleJson "github.com/bitly/go-simplejson"
	"github.com/gomodule/redigo/redis"
	LMCommon "github.com/lianmi/servers/internal/common"
	"github.com/lianmi/servers/internal/pkg/sts"
	// "github.com/lianmi/servers/internal/pkg/models"
	"github.com/robfig/cron"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (nc *NsqClient) RunCron() {
	nc.logger.Info("RunCron start...")
	c := cron.New()

	//解禁  每天凌晨1点执行一次："0 0 1 * * ?"
	c.AddFunc("0 0 1 * * ?", func() {
		nc.logger.Info("DissMuteTeamUsers start...")

		redisConn := nc.redisPool.Get()
		defer redisConn.Close()

		//ZRANGE Teams，取出所有群组id
		teamIDs, _ := redis.Strings(redisConn.Do("ZRANGE", "Teams", 0, -1))
		for _, teamID := range teamIDs {
			//for..range DissMuteUsers:{群组id}， 如果分数小于当前毫秒，则解禁
			dissMuteUsers, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("DissMuteUsers:%s", teamID), "-inf", time.Now().UnixNano()/1e6))
			for _, dissMuteUser := range dissMuteUsers {
				nc.logger.Debug(fmt.Sprintf("dissMuteUser:%s", dissMuteUser))

				key := fmt.Sprintf("TeamUser:%s:%s", teamID, dissMuteUser)

				//写入MySQL
				if err := nc.service.SetMuteTeamUser(teamID, dissMuteUser, false, 0); err != nil {
					nc.logger.Error("SetMuteTeamUser Error", zap.Error(err))
					continue
				}

				//刷新redis
				if _, err := redisConn.Do("HSET", key, "IsMute", 0); err != nil {
					nc.logger.Error("错误：HMSET teamUser", zap.Error(err))
					continue
				}
				nc.logger.Info("DissMuteTeamUsers ok", zap.String("dissMuteUser", dissMuteUser))

			}
			//一次性删除禁言的集合成员
			if _, err := redisConn.Do("ZREMRANGEBYSCORE", fmt.Sprintf("DissMuteUsers:%s", teamID), "-inf", time.Now().UnixNano()/1e6); err != nil {
				nc.logger.Error("ZREMRANGEBYSCORE Error", zap.Error(err))
				continue
			}

		}

		nc.logger.Info("DissMuteTeamUsers done.")

		//查询出用户表所有用户账号
		usernames, err := nc.service.QueryAllUsernames()
		if err != nil {
			return
		}

		var businessUsers []string

		for _, username := range usernames {
			userlikeKey := fmt.Sprintf("UserLike:%s", username)

			if businessUsers, err = redis.Strings(redisConn.Do("SMEMBERS", userlikeKey)); err != nil {
				nc.logger.Error("SMEMBERS Error", zap.Error(err))
				continue
			}

			for _, businessUser := range businessUsers {
				//将记录插入到UserLike表
				nc.logger.Debug("SMEMBERS 将记录插入到UserLike表", zap.String("username", username), zap.String("businessUser", businessUser))
				if err := nc.service.AddUserLike(username, businessUser); err != nil {
					continue
				}

			}

		}

		nc.logger.Info("AddUserLike done.")

	})

	//生成oss完全控制权限的token, 仅仅是服务端使用，有效期是1小时，每隔自动刷新一次，保存在redis里
	c.AddFunc("@hourly", func() {
		nc.RefreshOssSTSToken()

	})

	//启动任务
	c.Start()

	// 关闭任务
	defer c.Stop()

	var (
		sigchan chan os.Signal
		run     bool = true
	)

	sigchan = make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	for run == true {
		select {
		case sig := <-sigchan:
			nc.logger.Info("Caught signal terminating")
			_ = sig
			run = false
		}
	}

	nc.logger.Info("RunCron end")

}

func (nc *NsqClient) RefreshOssSTSToken() {
	var err error
	var client *sts.AliyunStsClient
	var url string

	client = sts.NewStsClient(LMCommon.AccessID, LMCommon.AccessKey, LMCommon.RoleAcs)
	//生成阿里云oss临时sts, Policy是对lianmi-ipfs这个bucket下的 avatars, generalavatars, msg, products, stores, teamicons, 目录有可读写权限

	// Policy是对lianmi-ipfs这个bucket下的user目录有可读写权限
	acsAvatars := fmt.Sprintf("acs:oss:*:*:lianmi-ipfs/avatars/*")
	acsGeneralavatars := fmt.Sprintf("acs:oss:*:*:lianmi-ipfs/generalavatars/*")
	acsMsg := fmt.Sprintf("acs:oss:*:*:lianmi-ipfs/msg/*")
	acsProducts := fmt.Sprintf("acs:oss:*:*:lianmi-ipfs/products/*")
	acsStores := fmt.Sprintf("acs:oss:*:*:lianmi-ipfs/stores/*")
	acsOrders := fmt.Sprintf("acs:oss:*:*:lianmi-ipfs/orders/*")
	acsTeamIcons := fmt.Sprintf("acs:oss:*:*:lianmi-ipfs/teamicons/*")
	acsUsers := fmt.Sprintf("acs:oss:*:*:lianmi-ipfs/users/*")

	// Policy是对lianmi-ipfs这个bucket下的user目录有可读写权限
	policy := sts.Policy{
		Version: "1",
		Statement: []sts.StatementBase{sts.StatementBase{
			Effect:   "Allow",
			Action:   []string{"oss:GetObject", "oss:ListObjects", "oss:PutObject", "oss:AbortMultipartUpload"},
			Resource: []string{acsAvatars, acsGeneralavatars, acsMsg, acsProducts, acsStores, acsOrders, acsTeamIcons, acsUsers},
		}},
	}

	//1小时过期
	url, err = client.GenerateSignatureUrl("lianmiserver", fmt.Sprintf("%d", LMCommon.EXPIRESECONDS), policy.ToJson())
	if err != nil {
		nc.logger.Error("GenerateSignatureUrl Error", zap.Error(err))
		return
	}

	data, err := client.GetStsResponse(url)
	if err != nil {
		nc.logger.Error("阿里云oss GetStsResponse Error", zap.Error(err))
		return
	}

	// log.Println("result:", string(data))
	sjson, err := simpleJson.NewJson(data)
	if err != nil {
		nc.logger.Warn("simplejson.NewJson Error", zap.Error(err))
		return
	}
	accessKeyID := sjson.Get("Credentials").Get("AccessKeyId").MustString()
	accessSecretKey := sjson.Get("Credentials").Get("AccessKeySecret").MustString()
	securityToken := sjson.Get("Credentials").Get("SecurityToken").MustString()

	nc.logger.Debug("收到阿里云OSS服务端的回包",
		zap.String("RequestId", sjson.Get("RequestId").MustString()),
		zap.String("AccessKeyId", accessKeyID),
		zap.String("AccessKeySecret", accessSecretKey),
		zap.String("SecurityToken", securityToken),
		zap.String("Expiration", sjson.Get("Credentials").Get("Expiration").MustString()),
	)

	if accessKeyID == "" || accessSecretKey == "" || securityToken == "" {
		nc.logger.Warn("获取STS错误")
		return

	}

	//保存到redis里
	redisConn := nc.redisPool.Get()
	defer redisConn.Close()
	redisConn.Do("SET", "OSSAccessKeyId", accessKeyID)
	redisConn.Do("EXPIRE", "OSSAccessKeyId", LMCommon.EXPIRESECONDS) //设置失效时间为1小时

	redisConn.Do("SET", "OSSAccessKeySecret", accessSecretKey)
	redisConn.Do("EXPIRE", "OSSAccessKeySecret", LMCommon.EXPIRESECONDS) //设置失效时间为1小时

	redisConn.Do("SET", "OSSSecurityToken", securityToken)
	redisConn.Do("EXPIRE", "OSSSecurityToken", LMCommon.EXPIRESECONDS) //设置失效时间为1小时
}
