package imcron

import (
	"fmt"

	"github.com/gomodule/redigo/redis"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	// golog "log"
	"time"
)

type WorkerJob interface {
	Run()
	Name() string
	run() error
}

type InitDataJob struct {
	manage *IMCronManage
	name   string
}

func initDataJob(manage *IMCronManage) *InitDataJob {
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
	// out := fmt.Sprintf("job:%s startTime:%s\n", j.name, startTime.String())

	// j.manage.nc.service.a
	defer func() {
		// fmt.Printf("job:%s endtime:%s cost:%s\n", j.name, time.Now().String(), time.Now().Sub(startTime).String())
	}()

	// fmt.Printf("job run \n ")
	_ = startTime
	return nil
}

type DissMuteTeamUsersJob struct {
	jobId  cron.EntryID
	spec   string
	manage *IMCronManage
	name   string
}

func initDissMuteTeamUsersJob(manage *IMCronManage,
	spec string) (j *DissMuteTeamUsersJob, err error) {

	tmp := new(DissMuteTeamUsersJob)
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

func (j *DissMuteTeamUsersJob) Run() {

	j.manage.cronWorker.crontab.Remove(j.jobId)

	defer func() {
		j.jobId, _ = j.manage.cronWorker.crontab.AddJob(j.spec, j)
	}()

	_ = j.run()
}

func (j *DissMuteTeamUsersJob) Name() string {
	return j.name
}

//job主体逻辑
func (j *DissMuteTeamUsersJob) run() (err error) {

	startTime := time.Now()
	out := fmt.Sprintf("job:%s startTime:%s", j.name, startTime.String())
	j.manage.logger.Debug(out)

	// fmt.Printf("printf job:%s starttime:%s\n", j.name, startTime.String())

	defer func() {
		out := fmt.Sprintf("job:%s endtime:%s cost:%s", j.name, time.Now().String(), time.Now().Sub(startTime).String())
		j.manage.logger.Debug(out)
	}()

	// fmt.Printf("===> DissMuteTeamUsersJob run\n")

	j.manage.logger.Info("DissMuteTeamUsers start...")

	redisConn := j.manage.redisPool.Get()
	defer redisConn.Close()

	//ZRANGE Teams，取出所有群组id
	teamIDs, _ := redis.Strings(redisConn.Do("ZRANGE", "Teams", 0, -1))
	for _, teamID := range teamIDs {
		//for..range DissMuteUsers:{群组id}， 如果分数小于当前毫秒，则解禁
		dissMuteUsers, _ := redis.Strings(redisConn.Do("ZRANGEBYSCORE", fmt.Sprintf("DissMuteUsers:%s", teamID), "-inf", time.Now().UnixNano()/1e6))
		for _, dissMuteUser := range dissMuteUsers {
			j.manage.logger.Debug(fmt.Sprintf("dissMuteUser:%s", dissMuteUser))

			key := fmt.Sprintf("TeamUser:%s:%s", teamID, dissMuteUser)

			//写入MySQL
			if err := j.manage.service.SetMuteTeamUser(teamID, dissMuteUser, false, 0); err != nil {
				j.manage.logger.Error("SetMuteTeamUser Error", zap.Error(err))
				continue
			}

			//刷新redis
			if _, err := redisConn.Do("HSET", key, "IsMute", 0); err != nil {
				j.manage.logger.Error("错误：HSET teamUser", zap.Error(err))
				continue
			}
			j.manage.logger.Info("DissMuteTeamUsers ok", zap.String("dissMuteUser", dissMuteUser))

		}
		//一次性删除禁言的集合成员
		if _, err := redisConn.Do("ZREMRANGEBYSCORE", fmt.Sprintf("DissMuteUsers:%s", teamID), "-inf", time.Now().UnixNano()/1e6); err != nil {
			j.manage.logger.Error("ZREMRANGEBYSCORE Error", zap.Error(err))
			continue
		}

	}

	j.manage.logger.Info("DissMuteTeamUsers done.")

	return nil
}
