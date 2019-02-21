package common

const (
	//ETCD中 任务保存的目录(键值前缀)
	JOB_SAVE_DIR = "/cron/jobs/"

	//ETCD中 任务强杀保存的目录(键值前缀)
	JOB_KILL_DIR = "/cron/killer/"

	//ETC中 任务分布式锁的目录(键值前缀)
	JOB_LOCK_DIR = "/cron/lock/"

	//ETCD 中服务发现与注册的目录(键值前缀)
	JOB_REGISTER_DIR = "/cron/register/"

	//ETCD中任务的变更事件 保存事件
	JOB_EVENT_SAVE = 1

	//ETCD中任务的变更事件 删除事件
	JOB_EVENT_DELETE = 2

	//任务强杀事件
	JOB_EVENT_KILL = 3

	//任务类型:shell命令任务
	JOB_TYPE_SHELL = 1

	//任务类型:远程WEB触发式任务
	JOB_TYPE_CURL = 2

	//默认worker分组名称
	DEFAULT_WORKER_GROUP_NAME = "默认分组"

	//任务执行不限分组
	UNLIMIT_WORKER_GROUP_NAME = "不限分组"
)
