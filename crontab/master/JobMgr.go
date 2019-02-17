package master

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
	Client *clientv3.Client
	Kv     clientv3.KV
	Lease  clientv3.Lease
}

var (
	G_jobmgr JobMgr
)

func LoadJobMgr() (err error) {
	var (
		config clientv3.Config
		client *clientv3.Client
		kv     clientv3.KV
		lease  clientv3.Lease
	)
	config = clientv3.Config{
		Endpoints:   G_config.EtcdEndpoints,
		DialTimeout: time.Second,
	}
	if client, err = clientv3.New(config); err != nil {
		//fmt.Println("连接失败")
		return
	}

	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)

	//if _,err=kv.Put(context.TODO(),"/test/a","ABC");err!=nil{
	//	fmt.Println(err)
	//	return err
	//}else{
	//	_,err=kv.Get(context.TODO(),"/test/a")
	//	fmt.Println(err)
	//	return
	//}
	G_jobmgr = JobMgr{
		Client: client,
		Kv:     kv,
		Lease:  lease,
	}

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

func (this *JobMgr) ListJobs() (jobList []*common.Job, count int64, err error) {
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
	count = getRes.Count
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
