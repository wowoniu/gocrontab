package worker

import (
	"context"
	"gocrontab/crontab/common"
	"os/exec"
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
	var (
		cmd           *exec.Cmd
		output        []byte
		err           error
		startTime     time.Time
		endTime       time.Time
		jobExecuteRes *common.JobExecuteResult
		jobLock       *JobLock
	)
	//获取锁对象
	jobLock = CreateJobLock(jobExecuteInfo.Job.Name, G_jobmgr.Kv, G_jobmgr.Lease)
	//执行加锁
	if err = jobLock.TryLock(); err != nil {
		//加锁失败
		startTime = time.Now()
		jobExecuteRes = &common.JobExecuteResult{
			JobExecuteInfo: jobExecuteInfo,
			Err:            err,
			StartTime:      startTime,
			EndTime:        startTime,
		}
	} else {
		//加锁成功
		startTime = time.Now()
		cmd = exec.CommandContext(context.TODO(), G_config.ExecuteShell, "-c", jobExecuteInfo.Job.Command)
		//命令执行 并获取输出
		output, err = cmd.CombinedOutput()
		endTime = time.Now()
		//命令结果
		jobExecuteRes = &common.JobExecuteResult{
			JobExecuteInfo: jobExecuteInfo,
			OutPut:         output,
			Err:            err,
			StartTime:      startTime,
			EndTime:        endTime,
		}
	}
	//执行结果写入调度器管道
	G_scheduler.PushJobResult(jobExecuteRes)
	return
}
