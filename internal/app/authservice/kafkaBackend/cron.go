package kafkaBackend

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/lianmi/servers/internal/pkg/models"
	"github.com/robfig/cron"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (kc *KafkaClient) RunCron() {
	kc.logger.Info("RunCron start...")
	c := cron.New()

	//解禁  每天凌晨1点执行一次："0 0 1 * * ?"
	c.AddFunc("0 0 1 * * ?", func() {
		kc.logger.Info("DissMuteTeamUsers start...")

		redisConn := kc.redisPool.Get()
		defer redisConn.Close()

		//ZRANGE Teams，取出所有群组id
		teamIDs, _ := redis.Strings(redisConn.Do("ZRANGE", "Teams", 0, -1))
		for _, teamID := range teamIDs {
			//for..range DissMuteUsers:{群组id}， 如果分数小于当前毫秒，则解禁
			dissMuteUsers, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("DissMuteUsers:%s", teamID), "-inf", time.Now().UnixNano()/1e6))
			for _, dissMuteUser := range dissMuteUsers {
				kc.logger.Debug(fmt.Sprintf("dissMuteUser:%s", dissMuteUser))

				key := fmt.Sprintf("TeamUser:%s:%s", teamID, dissMuteUser)
				teamUser := new(models.TeamUser)
				if result, err := redis.Values(redisConn.Do("HGETALL", key)); err == nil {
					if err := redis.ScanStruct(result, teamUser); err != nil {
						kc.logger.Error("错误：ScanStruct", zap.Error(err))
						continue
					}
				}

				//解除禁言
				teamUser.IsMute = false

				//写入MySQL
				if err := kc.SaveTeamUser(teamUser); err != nil {
					kc.logger.Error("Save teamUser Error", zap.Error(err))
					continue
				}

				//刷新redis
				if _, err := redisConn.Do("HMSET", redis.Args{}.Add(fmt.Sprintf("TeamUser:%s:%s", teamID, dissMuteUser)).AddFlat(teamUser)...); err != nil {
					kc.logger.Error("错误：HMSET teamUser", zap.Error(err))
					continue
				}
				kc.logger.Info("DissMuteTeamUsers ok", zap.String("dissMuteUser", dissMuteUser))

			}
			//一次性删除禁言的集合成员
			if _, err := redisConn.Do("ZREMRANGEBYSCORE", fmt.Sprintf("DissMuteUsers:%s", teamID), "-inf", time.Now().UnixNano()/1e6); err != nil {
				kc.logger.Error("ZREMRANGEBYSCORE Error", zap.Error(err))
				continue
			}

		}
		kc.logger.Info("DissMuteTeamUsers done.")
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
			kc.logger.Info("Caught signal terminating")
			_ = sig
			run = false

		}
	}

	// for run == true {

	// }
	kc.logger.Info("RunCron end")

}
