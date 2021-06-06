package huawei

import (
	"errors"
	"fmt"
	golog "log"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/robfig/cron/v3"
)

const (
	RedisAddr = "127.0.0.1:6379"
)

type CronWorker struct {
	manager *HuaweiManage
	crontab *cron.Cron
}

func NewCronWorker(manage *HuaweiManage) *CronWorker {
	worker := new(CronWorker)
	worker.manager = manage
	var _ = golog.LstdFlags

	worker.crontab = cron.New() //cron.New(cron.WithLogger(cron.VerbosePrintfLogger(golog.New(os.Stdout, "cron: ", golog.LstdFlags))))

	return worker
}

func (w *CronWorker) Stop() {
	beginTime := time.Now()
	ctx := w.crontab.Stop()
	select {
	case <-ctx.Done():
		fmt.Printf("coin manage:%s  stop cost:%s\n", w.manager.GetName(), time.Now().Sub(beginTime).String())

	}
}

func (w *CronWorker) Prepare() error {
	jobList := []WorkerJob{
		initDataJob(w.manager),
	}

	for _, job := range jobList {
		fmt.Printf("worker %s: prepare %s\n", w.manager.GetName(), job.Name())
		err := job.run()
		if err != nil {
			fmt.Printf("worker %s: prepare %s err:%s\n", w.manager.GetName(), job.Name(), err.Error())
			return err
		}
	}
	//TODO: 判断redis里的access_token是否失效 获取一次access_token

	redisConn, err := redis.Dial("tcp", RedisAddr)
	if err != nil {
		fmt.Println(err)
		return err
	}

	defer redisConn.Close()

	huaweiAccessToken, err := redis.String(redisConn.Do("GET", "HUAWEI_ACCESS_TOEKN"))
	if err != nil {
		fmt.Println("GET HUAWEI_ACCESS_TOEKN", err)
		// return err
	}
	if huaweiAccessToken == "" {
		fmt.Println("huaweiAccessToken is empty, need refresh")
		//TODO POST

	}

	fmt.Printf("huaweiAccessToken: %s, worker %s: prepare end\n", huaweiAccessToken, w.manager.GetName())
	return nil
}

func (w *CronWorker) Start() error {
	var err error
	var expr string
	_ = err
	expr = "@every 59m" //59分钟获取一次
	if len(expr) != 0 {
		if w.manager == nil {
			fmt.Println("w.manager == nil")
			return errors.New("w.manager == nil")
		}

		_, err = initRefreshHuaweiAssessTokenJob(w.manager, expr)
		if err != nil {
			return err
		}

		//

	}

	w.crontab.Start()
	// fmt.Printf("worker %s Start end\n", w.manager.GetName())
	return nil
}
