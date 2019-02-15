package worker

import (
	"context"
	"encoding/json"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"gocrontab/crontab/common"
	"time"
)

/**
任务管理器(负责与etcd等交互)
*/

type JobMgr struct {
	Client  *clientv3.Client
	Kv      clientv3.KV
	Lease   clientv3.Lease
	Watcher clientv3.Watcher
}

var (
	G_jobmgr *JobMgr
)

func LoadJobMgr() (err error) {
	var (
		config  clientv3.Config
		client  *clientv3.Client
		kv      clientv3.KV
		lease   clientv3.Lease
		watcher clientv3.Watcher
	)
	config = clientv3.Config{
		Endpoints:   G_config.EtcdEndpoints,
		DialTimeout: time.Duration(G_config.EtcdDialTimeout) * time.Millisecond,
	}
	if client, err = clientv3.New(config); err != nil {
		//fmt.Println("连接失败")
		return
	}

	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)
	watcher = clientv3.NewWatcher(client)

	G_jobmgr = &JobMgr{
		Client:  client,
		Kv:      kv,
		Lease:   lease,
		Watcher: watcher,
	}

	//启动监听任务协程
	G_jobmgr.WatchJobs()

	//监听强杀协程
	G_jobmgr.WatchKill()
	return
}

//监听任务的变更
func (this *JobMgr) WatchJobs() {
	var (
		getRes              *clientv3.GetResponse
		kValue              *mvccpb.KeyValue
		job                 *common.Job
		currentRevision     int64
		startListenRevision int64
		watchChan           clientv3.WatchChan
		watchRes            clientv3.WatchResponse
		watchEvent          *clientv3.Event
		jobEvent            *common.JobEvent
		err                 error
	)
	//获取所有任务列表
	if getRes, err = this.Kv.Get(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithPrefix()); err != nil {
		return
	}
	//遍历任务 发送给调度协程
	for _, kValue = range getRes.Kvs {
		if job, err = common.UnpackJob(kValue.Value); err == nil {
			jobEvent = &common.JobEvent{
				Job:       job,
				EventType: common.JOB_EVENT_SAVE,
			}
			G_scheduler.PushJobEvent(jobEvent)
		}
	}
	//取得当前的Revision版本
	currentRevision = getRes.Header.Revision
	startListenRevision = currentRevision + 1
	//监听etcd 中任务的变化
	go func() {
		var (
			err      error
			job      *common.Job
			jobName  string
			jobEvent *common.JobEvent
		)
		//监听etcd的变化
		watchChan = this.Watcher.Watch(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithRev(startListenRevision), clientv3.WithPrefix())

		for watchRes = range watchChan {
			//遍历变化的事件
			for _, watchEvent = range watchRes.Events {
				switch watchEvent.Type {
				case mvccpb.PUT:
					//任务保存 (新建或修改) 获取保存的值
					if job, err = common.UnpackJob(watchEvent.Kv.Value); err != nil {
						continue
					}
					//构造变更事件
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_SAVE, job)
				case mvccpb.DELETE:
					//任务删除
					jobName = common.ExtractJobName(string(watchEvent.Kv.Key))
					job = &common.Job{
						Name: jobName,
					}
					//构造变更事件
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_DELETE, job)
				}
				//事件推送到调度协程
				G_scheduler.PushJobEvent(jobEvent)
			}
		}
	}()
	return
}

//监听任务的强杀
func (this *JobMgr) WatchKill() {
	var (
		//getRes              *clientv3.GetResponse
		//currentRevision     int64
		//startListenRevision int64
		watchChan  clientv3.WatchChan
		watchRes   clientv3.WatchResponse
		watchEvent *clientv3.Event
		//err error
	)
	//获取所有需要强杀的任务列表
	//if getRes, err = this.Kv.Get(context.TODO(), common.JOB_KILL_DIR, clientv3.WithPrefix()); err != nil {
	//	return
	//}
	//取得当前的Revision版本
	//currentRevision = getRes.Header.Revision
	//startListenRevision = currentRevision + 1
	//监听etcd 中kill目录的变化
	go func() {
		var (
			job      *common.Job
			jobName  string
			jobEvent *common.JobEvent
		)
		//监听killer目录的变化
		//watchChan = this.Watcher.Watch(context.TODO(), common.JOB_KILL_DIR, clientv3.WithRev(startListenRevision), clientv3.WithPrefix())
		watchChan = this.Watcher.Watch(context.TODO(), common.JOB_KILL_DIR, clientv3.WithPrefix())

		for watchRes = range watchChan {
			//遍历变化的事件
			for _, watchEvent = range watchRes.Events {
				jobName = common.ExtractKillJobName(string(watchEvent.Kv.Key))
				switch watchEvent.Type {
				case mvccpb.PUT:
					//强杀
					job = &common.Job{
						Name: jobName,
					}
					//构造强杀事件
					jobEvent = common.BuildJobEvent(common.JOB_EVENT_KILL, job)
					//事件推送到调度协程
					G_scheduler.PushJobEvent(jobEvent)
				case mvccpb.DELETE:
				}
			}
		}
	}()
	return
}

func (this *JobMgr) SaveJob(job *common.Job) (oldJob *common.Job, err error) {
	var (
		jobKey    string
		jobValue  []byte
		putRes    *clientv3.PutResponse
		oldJobObj common.Job
	)
	jobKey = common.JOB_SAVE_DIR + job.Name
	if jobValue, err = json.Marshal(job); err != nil {
		return
	}
	//保存到etcd
	if putRes, err = this.Kv.Put(context.TODO(), jobKey, string(jobValue), clientv3.WithPrevKV()); err != nil {
		return
	}
	//如果是更新 有旧值
	if putRes.PrevKv != nil {
		if err = json.Unmarshal(putRes.PrevKv.Value, &oldJobObj); err != nil {
			//旧值获取失败 不影响新值的保存
			err = nil
			return
		}
		oldJob = &oldJobObj
	}
	return
}

func (this *JobMgr) DeleteJob(jobName string) (oldJob *common.Job, err error) {
	var (
		delRes    *clientv3.DeleteResponse
		jobKey    string
		oldJobObj common.Job
	)
	jobKey = common.JOB_SAVE_DIR + jobName
	if delRes, err = this.Kv.Delete(context.TODO(), jobKey, clientv3.WithPrevKV()); err != nil {
		//删除失败
		return
	}
	if len(delRes.PrevKvs) != 0 {
		if err = json.Unmarshal(delRes.PrevKvs[0].Value, &oldJobObj); err != nil {
			//反序列化失败 不影响删除业务
			err = nil
			return
		}
		oldJob = &oldJobObj
	}
	return
}

func (this *JobMgr) ListJobs() (jobList []*common.Job, err error) {
	var (
		dirKey  string
		getRes  *clientv3.GetResponse
		job     *common.Job
		jobJson *mvccpb.KeyValue
	)
	dirKey = common.JOB_SAVE_DIR
	if getRes, err = this.Kv.Get(context.TODO(), dirKey, clientv3.WithPrefix()); err != nil {
		return
	}

	jobList = make([]*common.Job, 0)
	for _, jobJson = range getRes.Kvs {
		job = &common.Job{}
		if err = json.Unmarshal(jobJson.Value, job); err != nil {
			continue
		}
		jobList = append(jobList, job)
	}
	return
}

func (this *JobMgr) KillJob(jobName string) (err error) {
	var (
		killerKey     string
		leaseGrantRes *clientv3.LeaseGrantResponse
	)
	killerKey = common.JOB_KILL_DIR + jobName
	//创建一个1秒过期的租约
	if leaseGrantRes, err = this.Lease.Grant(context.TODO(), 1); err != nil {
		//租约创建失败
		return
	}
	//PUT
	_, err = this.Kv.Put(context.TODO(), killerKey, "", clientv3.WithLease(leaseGrantRes.ID))

	return
}
