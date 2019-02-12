package main

import (
	"fmt"
	"github.com/gorhill/cronexpr"
	"time"
)

type cronJob struct {
	jobName  string
	expr     *cronexpr.Expression
	nextTime time.Time
}

func main() {
	var (
		exp           *cronexpr.Expression
		err           error
		now           time.Time
		scheduleTable map[string]*cronJob
	)
	now = time.Now()
	if exp, err = cronexpr.Parse("*/5 * * * * *"); err != nil {
		fmt.Println("cron error:", err)
		return
	}
	fmt.Println(exp.Next(now))

	scheduleTable = make(map[string]*cronJob)
	exp = cronexpr.MustParse("*/5 * * * * * *")
	scheduleTable["job1"] = &cronJob{
		"job1",
		exp,
		exp.Next(now),
	}

	exp = cronexpr.MustParse("*/2 * * * * * *")
	scheduleTable["job2"] = &cronJob{
		"job2",
		exp,
		exp.Next(now),
	}

	//调度协程
	go func() {
		var (
			jobName string
			cronjob *cronJob
			now     time.Time
		)
		for {
			now = time.Now()
			for jobName, cronjob = range scheduleTable {
				//执行任务协程
				if cronjob.nextTime.Before(now) {
					go func(jobName string) {
						fmt.Println(now, "任务执行:", jobName)
					}(jobName)
					//计算下一次执行时间
					cronjob.nextTime = cronjob.expr.Next(now)
					//fmt.Println(jobName,"下一次执行时间:",cronjob.nextTime)
				}
			}
			//睡眠100毫秒
			select {
			case <-time.Tick(100 * time.Microsecond):
			}
			//time.Sleep(100*time.Microsecond)
		}
	}()

	time.Sleep(20 * time.Second)
}
