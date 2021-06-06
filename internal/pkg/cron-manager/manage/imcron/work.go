package imcron

import (
	"errors"
	"fmt"
	golog "log"
	"time"

	// "github.com/gomodule/redigo/redis"
	"github.com/robfig/cron/v3"
)

type CronWorker struct {
	manager *IMCronManage
	crontab *cron.Cron
}

func NewCronWorker(manage *IMCronManage) *CronWorker {
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

	// fmt.Printf("huaweiAccessToken: %s, worker %s: prepare end\n", huaweiAccessToken, w.manager.GetName())
	return nil
}

func (w *CronWorker) Start() error {
	var err error
	var expr string
	_ = err
	expr = "0 0 1 * * ?" //解禁  每天凌晨1点执行一次："0 0 1 * * ?"
	if len(expr) != 0 {
		if w.manager == nil {
			fmt.Println("w.manager == nil")
			return errors.New("w.manager == nil")
		}

		_, err = initDissMuteTeamUsersJob(w.manager, expr)
		if err != nil {
			return err
		}

		//

	}

	w.crontab.Start()
	// fmt.Printf("worker %s Start end\n", w.manager.GetName())
	return nil
}
