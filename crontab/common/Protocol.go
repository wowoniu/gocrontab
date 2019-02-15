package common

import (
	"encoding/json"
	"strings"
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

//构造调度协程的事件对象
func BuildJobEvent(eventType int, job *Job) *JobEvent {
	return &JobEvent{
		EventType: eventType,
		Job:       job,
	}
}
