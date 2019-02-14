package common

import "encoding/json"

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

//job JSON反序列化
func UnpackJob(data interface{}) (job *Job, err error) {
	if err = json.Unmarshal(data.([]byte), job); err != nil {
		return
	}
	return
}
