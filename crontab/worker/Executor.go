package worker

import (
	"context"
	"gocrontab/crontab/common"
	"time"
)

/**
任务执行器
*/

type JobExecutor struct {
}

var G_jobexecutor *JobExecutor

func LoadJobExecutor() {
	G_jobexecutor = &JobExecutor{}
}

//执行shell命令
func (this *JobExecutor) ExecJob(jobExecuteInfo *common.JobExecuteInfo) {
	go func() {
		var (
			output        []byte
			err           error
			startTime     time.Time
			endTime       time.Time
			jobExecuteRes *common.JobExecuteResult
			jobLock       *JobLock
			cancelCtx     context.Context
			cancelFunc    context.CancelFunc
		)
		//获取锁对象
		jobLock = CreateJobLock(jobExecuteInfo.Job.Name, G_jobmgr.Kv, G_jobmgr.Lease)
		defer jobLock.Unlock()
		//执行加锁
		err = jobLock.TryLock()
		startTime = time.Now()
		jobExecuteRes = &common.JobExecuteResult{
			JobExecuteInfo: jobExecuteInfo,
			StartTime:      startTime,
			OutPut:         make([]byte, 0),
		}
		if err != nil {
			//上锁失败
			//fmt.Println("抢锁失败")
			err = common.ERR_LOCK_ALREADY_REQUIRED
			jobExecuteRes.EndTime = time.Now()
		} else {
			//上锁成功
			cancelCtx, cancelFunc = context.WithCancel(context.TODO())
			jobExecuteInfo.CancelFunc = cancelFunc
			jobExecuteInfo.CancelCtx = cancelCtx
			//执行
			switch jobExecuteInfo.Job.Type {
			case common.JOB_TYPE_SHELL:
				output, err = G_shellExecutor.Exec(cancelCtx, jobExecuteInfo.Job.Command)
			case common.JOB_TYPE_CURL:
				output, err = G_webExecutor.Exec(cancelCtx, jobExecuteInfo.Job.Command)
			}
			endTime = time.Now()
			//命令结果
			jobExecuteRes.OutPut = output
			jobExecuteRes.EndTime = endTime
		}
		//执行结果写入调度器管道
		jobExecuteRes.Err = err
		G_scheduler.PushJobResult(jobExecuteRes)
		return
	}()
}
