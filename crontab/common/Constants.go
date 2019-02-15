package common

const (
	//ETCD中 任务保存的目录(键值前缀)
	JOB_SAVE_DIR = "/cron/jobs/"

	//ETCD中 任务强杀保存的目录(键值前缀)
	JOB_KILL_DIR = "/cron/killer/"

	//ETC中 任务分布式锁的目录(键值前缀)
	JOB_LOCK_DIR = "/cron/lock/"

	//ETCD中任务的变更事件 保存事件
	JOB_EVENT_SAVE = 1

	//ETCD中任务的变更事件 删除事件
	JOB_EVENT_DELETE = 2
)
