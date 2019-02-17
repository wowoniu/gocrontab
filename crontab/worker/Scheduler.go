package worker

import (
	"github.com/gorhill/cronexpr"
	"gocrontab/crontab/common"
	"time"
)

type Scheduler struct {
	JobEventChan         chan *common.JobEvent              //任务事件队列
	JobPlanTable         map[string]*common.JobSchedulePlan //任务执行计划列表
	JobExecuteingTable   map[string]*common.JobExecuteInfo  //正在执行的任务列表
	JobExecuteResultChan chan *common.JobExecuteResult
}

var G_scheduler *Scheduler

func LoadScheduler() (err error) {
	G_scheduler = &Scheduler{
		JobEventChan:         make(chan *common.JobEvent, 1),
		JobPlanTable:         make(map[string]*common.JobSchedulePlan),
		JobExecuteingTable:   make(map[string]*common.JobExecuteInfo),
		JobExecuteResultChan: make(chan *common.JobExecuteResult, 1000),
	}
	//开启任务状态监听协程
	go G_scheduler.scheduleLoop()
	return
}

//循环调度逻辑
func (this *Scheduler) scheduleLoop() {
	var (
		jobEvent          *common.JobEvent
		afterScheduleTime time.Duration
		afterTimer        *time.Timer
		jobExecuteRes     *common.JobExecuteResult
	)
	//第一次任务调度 计算出最近一次任务离现在的时间
	afterScheduleTime = this.trySchedule()
	//创建延时定时器
	afterTimer = time.NewTimer(afterScheduleTime)
	for {
		select {
		//任务事件管道监听
		case jobEvent = <-this.JobEventChan:
			//fmt.Println("任务事件管道消息:",jobEvent)
			this.HandleJobEvent(jobEvent)
		case <-afterTimer.C:
			//定时器 最近一次任务执行时间到达
		case jobExecuteRes = <-this.JobExecuteResultChan:
			this.HandleJobResult(jobExecuteRes)
		}
		//开始调度
		afterScheduleTime = this.trySchedule()
		//定时器重置
		afterTimer.Reset(afterScheduleTime)
	}
}

//尝试执行任务调度 返回任务队列中最近一次需要执行的任务至现在的时间
func (this *Scheduler) trySchedule() (afterTime time.Duration) {
	var (
		jobSchedulePlan *common.JobSchedulePlan
		now             time.Time
		nearestTime     *time.Time
	)
	now = time.Now()
	if len(this.JobPlanTable) == 0 {
		afterTime = 1 * time.Second
		return
	}
	//遍历当前的任务队列
	for _, jobSchedulePlan = range this.JobPlanTable {
		if jobSchedulePlan.NextTime.Before(now) {
			//已过期 立即执行
			this.DoSchedule(jobSchedulePlan)
			//计算下一次执行时间
			jobSchedulePlan.NextTime = jobSchedulePlan.Expr.Next(now)
		}
		//获取最近一次要执行任务的时间
		if nearestTime == nil || jobSchedulePlan.NextTime.Before(*nearestTime) {
			nearestTime = &jobSchedulePlan.NextTime
		}
	}
	afterTime = (*nearestTime).Sub(now)
	return
}

//推送任务事件
func (this *Scheduler) PushJobEvent(jobEvent *common.JobEvent) {
	//fmt.Println("任务变化:", jobEvent.Job.Name, jobEvent.EventType)
	this.JobEventChan <- jobEvent
}

//处理任务变更事件
func (this *Scheduler) HandleJobEvent(jobEvent *common.JobEvent) {
	var (
		err             error
		jobExisted      bool
		jobSchedulePlan *common.JobSchedulePlan
		jobExecuteInfo  *common.JobExecuteInfo
		jobExecuteing   bool
	)
	switch jobEvent.EventType {
	//任务删除
	case common.JOB_EVENT_DELETE:
		if _, jobExisted = this.JobPlanTable[jobEvent.Job.Name]; jobExisted {
			//删除本地的任务
			delete(this.JobPlanTable, jobEvent.Job.Name)
		}
	//任务变更
	case common.JOB_EVENT_SAVE:
		if jobSchedulePlan, err = this.BuildSchedulPlan(jobEvent.Job); err != nil {
			return
		}
		//加入计划队列
		this.JobPlanTable[jobEvent.Job.Name] = jobSchedulePlan
	//任务强杀
	case common.JOB_EVENT_KILL:
		if jobExecuteInfo, jobExecuteing = this.JobExecuteingTable[jobEvent.Job.Name]; jobExecuteing {
			//fmt.Println("任务强杀(执行中):",jobEvent.Job.Name)
			jobExecuteInfo.CancelFunc()
		}
	}
	//fmt.Println(this.JobPlanTable)
}

func (this *Scheduler) BuildSchedulPlan(job *common.Job) (schedulePlan *common.JobSchedulePlan, err error) {
	var (
		expr *cronexpr.Expression
		now  time.Time
	)
	if expr, err = cronexpr.Parse(job.CronExpr); err != nil {
		//表达式错误
		return
	}
	now = time.Now()
	schedulePlan = &common.JobSchedulePlan{
		Job:      job,
		Expr:     expr,
		NextTime: expr.Next(now),
	}
	return
}

//构造执行任务的信息数据体
func (this *Scheduler) BuildJobExecuteInfo(jobSchedulePlan *common.JobSchedulePlan) (jobExecuteInfo *common.JobExecuteInfo) {
	jobExecuteInfo = &common.JobExecuteInfo{
		Job:      jobSchedulePlan.Job,
		PlanTime: jobSchedulePlan.NextTime,
		RealTime: time.Now(),
	}
	return
}

//任务执行
func (this *Scheduler) DoSchedule(jobSchedulePlan *common.JobSchedulePlan) {
	var (
		jobExcuteing   bool
		jobExecuteInfo *common.JobExecuteInfo
	)
	if jobExecuteInfo, jobExcuteing = this.JobExecuteingTable[jobSchedulePlan.Job.Name]; jobExcuteing {
		//正在执行 跳过
		//fmt.Println(time.Now(),"正在执行中:", jobSchedulePlan.Job.Name, "跳过")
		return
	}
	//加入正在执行的队列中去
	jobExecuteInfo = this.BuildJobExecuteInfo(jobSchedulePlan)
	this.JobExecuteingTable[jobSchedulePlan.Job.Name] = jobExecuteInfo

	//任务执行器 执行任务
	//fmt.Println(time.Now(),"执行任务:",jobSchedulePlan.Job.Name,jobSchedulePlan.Job.Command)
	G_jobexecutor.ExecJob(jobExecuteInfo)
}

//推送任务结果
func (this *Scheduler) PushJobResult(jobExecuteRes *common.JobExecuteResult) {
	this.JobExecuteResultChan <- jobExecuteRes
}

//处理任务结果
func (this *Scheduler) HandleJobResult(jobExecuteRes *common.JobExecuteResult) {
	//从正在执行队列中移除任务
	delete(this.JobExecuteingTable, jobExecuteRes.JobExecuteInfo.Job.Name)
	//日志存储
	G_log.PushLog(this.BuildJobLog(jobExecuteRes))
	//fmt.Println(time.Now(), jobExecuteRes.JobExecuteInfo.Job.Name, "任务执行完毕:", string(jobExecuteRes.OutPut), jobExecuteRes.Err,"耗时:",jobExecuteRes.EndTime.Sub(jobExecuteRes.StartTime))
}

func (this *Scheduler) BuildJobLog(jobExecuteRes *common.JobExecuteResult) *common.JobLog {
	res := &common.JobLog{
		JobName:      jobExecuteRes.JobExecuteInfo.Job.Name,
		Command:      jobExecuteRes.JobExecuteInfo.Job.Command,
		Output:       string(jobExecuteRes.OutPut),
		PlanTime:     jobExecuteRes.JobExecuteInfo.PlanTime.Unix(),
		ScheduleTime: jobExecuteRes.JobExecuteInfo.RealTime.Unix(),
		StartTime:    jobExecuteRes.StartTime.Unix(),
		EndTime:      jobExecuteRes.EndTime.Unix(),
	}
	if jobExecuteRes.Err != nil {
		res.Err = jobExecuteRes.Err.Error()
	} else {
		res.Err = ""
	}
	return res
}
