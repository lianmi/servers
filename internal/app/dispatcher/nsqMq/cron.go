package nsqMq

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
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
