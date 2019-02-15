package common

const (
	//ETCD中 任务保存的目录(键值前缀)
	JOB_SAVE_DIR = "/cron/jobs/"

	//ETCD中 任务强杀保存的目录(键值前缀)
	JOB_KILL_DIR = "/cron/killer/"

	//ETCD中任务的变更事件 保存事件
	JOB_EVENT_SAVE = 1

	//ETCD中任务的变更事件 删除事件
	JOB_EVENT_DELETE = 2
)
