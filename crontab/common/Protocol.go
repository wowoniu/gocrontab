package common

import (
	"context"
	"encoding/json"
	"github.com/gorhill/cronexpr"
	"strings"
	"time"
)

//计划任务
type Job struct {
	Name     string `json:"name"`
	Command  string `json:"command"`
	CronExpr string `json:"cron_expr"`
	Desc     string `json:"desc"`
}

//接口响应
type ApiResponse struct {
	Errno int         `json:"errno"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"`
}

//任务变化调度事件
type JobEvent struct {
	Job       *Job
	EventType int
}

//任务的调度计划
type JobSchedulePlan struct {
	Job      *Job
	Expr     *cronexpr.Expression
	NextTime time.Time
}

//任务执行状态
type JobExecuteInfo struct {
	Job        *Job
	PlanTime   time.Time //计划执行时间
	RealTime   time.Time //实际执行时间
	CancelCtx  context.Context
	CancelFunc context.CancelFunc
}

//任务执行结果
type JobExecuteResult struct {
	JobExecuteInfo *JobExecuteInfo
	OutPut         []byte
	Err            error
	StartTime      time.Time
	EndTime        time.Time
}

type JobLog struct {
	JobName      string `bson:"job_name"`
	Command      string `bson:"command"`
	Err          string `bson:"err"`
	Output       string `bson:"output"`
	PlanTime     int64  `bson:"plan_time"`
	ScheduleTime int64  `bson:"schedule_time"`
	StartTime    int64  `bson:"start_time"`
	EndTime      int64  `bson:"end_time"`
}

//job JSON反序列化
func UnpackJob(data []byte) (job *Job, err error) {
	job = &Job{}
	if err = json.Unmarshal(data, job); err != nil {
		return
	}
	return
}

//从etcd的key中提取任务名
func ExtractJobName(jobKey string) string {
	return strings.TrimLeft(jobKey, JOB_SAVE_DIR)
}

func ExtractKillJobName(jobKey string) string {
	return strings.TrimLeft(jobKey, JOB_KILL_DIR)
}

//构造调度协程的事件对象
func BuildJobEvent(eventType int, job *Job) *JobEvent {
	return &JobEvent{
		EventType: eventType,
		Job:       job,
	}
}
