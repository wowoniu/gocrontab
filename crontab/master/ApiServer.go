package master

import (
	"encoding/json"
	"gocrontab/crontab/common"
	"net"
	"net/http"
	"strconv"
	"time"
)

type ApiServer struct {
	HttpServer *http.Server
}

var G_apiServer *ApiServer

func InitApiServer() (err error) {
	var (
		mux        *http.ServeMux //路由器
		listener   net.Listener
		httpServer *http.Server
	)
	//设置路由
	mux = http.NewServeMux()
	//API
	mux.HandleFunc("/job/save", handleJobSave)
	mux.HandleFunc("/job/delete", handleJobDelete)
	mux.HandleFunc("/job/list", handleJobList)
	mux.HandleFunc("/job/kill", handleJobKill)
	mux.HandleFunc("/job/log", handleJobLog)
	//静态文件路由
	mux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir(G_config.WebRoot))))
	//监听设置
	if listener, err = net.Listen("tcp", ":"+strconv.Itoa(G_config.ApiPort)); err != nil {
		return
	}
	//创建http服务器
	httpServer = &http.Server{
		ReadTimeout:  time.Duration(G_config.ReadTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(G_config.WriteTimeout) * time.Millisecond,
		Handler:      mux,
	}

	G_apiServer = &ApiServer{
		HttpServer: httpServer,
	}
	//启动协程 开启http服务
	go httpServer.Serve(listener)

	return
}

//任务保存接口
func handleJobSave(w http.ResponseWriter, r *http.Request) {
	//POST job={"name":"job1","command":"echo 1","cron_expr":"* * * * *","desc":"任务描述"}
	var (
		err    error
		jobStr string
		job    common.Job
		oldJob *common.Job
	)
	if err = r.ParseForm(); err != nil {
		output(w, 1000, "无效的POST请求:"+err.Error(), nil)
		return
	}
	jobStr = r.PostForm.Get("job")
	//josn反序列化
	if err = json.Unmarshal([]byte(jobStr), &job); err != nil {
		output(w, 1001, "无效的job配置:"+err.Error(), nil)
		return
	}
	//保存任务到etcd
	if oldJob, err = G_jobmgr.SaveJob(&job); err != nil {
		//保存失败 TODO
		output(w, 1003, "任务保存失败:"+err.Error(), nil)
		return
	}
	//保存成功

	//oldJob = oldJob
	output(w, 0, "success", oldJob)
	return
}

//任务删除接口
func handleJobDelete(w http.ResponseWriter, r *http.Request) {
	//POST name=job1
	var (
		err     error
		jobName string
		oldJob  *common.Job
	)
	if err = r.ParseForm(); err != nil {
		output(w, 1000, "无效的POST请求:"+err.Error(), nil)
		return
	}
	jobName = r.PostForm.Get("name")

	//删除任务
	if oldJob, err = G_jobmgr.DeleteJob(jobName); err != nil {
		output(w, 1101, "删除失败:"+err.Error(), nil)
		return
	}
	output(w, 0, "success", oldJob)
}

//获取所有任务列表
func handleJobList(w http.ResponseWriter, r *http.Request) {
	var (
		jobList []*common.Job
		err     error
	)
	if jobList, err = G_jobmgr.ListJobs(); err != nil {
		output(w, 12001, "列表获取失败:"+err.Error(), nil)
		return
	}

	output(w, 0, "success", jobList)
}

//强杀任务
func handleJobKill(w http.ResponseWriter, r *http.Request) {
	var (
		jobName string
		err     error
	)
	r.ParseForm()
	jobName = r.PostForm.Get("name")
	if err = G_jobmgr.KillJob(jobName); err != nil {
		output(w, 13001, "设置强杀失败:"+err.Error(), nil)
		return
	}
	output(w, 0, "success", nil)
}

//获取任务日志
func handleJobLog(w http.ResponseWriter, r *http.Request) {
	var (
		jobName string
		err     error
		logList []*common.JobLog
	)
	r.ParseForm()
	jobName = r.PostForm.Get("name")
	if logList, err = G_joblog.GetLogList(jobName); err != nil {
		output(w, 14001, "获取日志失败:"+err.Error(), nil)
		return
	}
	output(w, 0, "success", logList)
}

func output(w http.ResponseWriter, errno int, msg string, data interface{}) {
	var (
		res []byte
		err error
	)
	if res, err = buildResponse(errno, msg, data); err != nil {
		//log todo
		res, _ = buildResponse(9999, "系统错误", nil)
	}
	w.Write(res)
}

func buildResponse(errno int, msg string, data interface{}) (responseJson []byte, err error) {
	var (
		res common.ApiResponse
	)
	res.Errno = errno
	res.Msg = msg
	res.Data = data
	//序列化为JSON
	if responseJson, err = json.Marshal(res); err != nil {
		return
	}
	return
}
