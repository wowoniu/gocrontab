package worker

import (
	"context"
	"go.etcd.io/etcd/clientv3"
	"gocrontab/crontab/common"
)

type JobLock struct {
	Kv         clientv3.KV
	Lease      clientv3.Lease
	JobName    string
	LeaseId    clientv3.LeaseID
	CancelFunc context.CancelFunc
	IsLocked   bool
}

func CreateJobLock(jobName string, kv clientv3.KV, lease clientv3.Lease) *JobLock {
	return &JobLock{
		JobName:  jobName,
		Kv:       kv,
		Lease:    lease,
		IsLocked: false,
	}
}

//尝试任务加锁
func (this *JobLock) TryLock() (err error) {
	var (
		leaseGrantRes     *clientv3.LeaseGrantResponse
		leaseId           clientv3.LeaseID
		leaseKeepAliveRes <-chan *clientv3.LeaseKeepAliveResponse
		cancelCtx         context.Context
		cancelFunc        context.CancelFunc
		txn               clientv3.Txn
		lockKey           string
		txnRes            *clientv3.TxnResponse
	)
	//创建一个5S的租约
	if leaseGrantRes, err = this.Lease.Grant(context.TODO(), 5); err != nil {
		return
	}
	leaseId = leaseGrantRes.ID
	cancelCtx, cancelFunc = context.WithCancel(context.TODO())
	this.LeaseId = leaseId
	this.CancelFunc = cancelFunc
	//自动续租
	if leaseKeepAliveRes, err = this.Lease.KeepAlive(cancelCtx, leaseId); err != nil {
		this.checkError()
		return
	}
	//开启协程 读取需要channel数据
	go func() {
		var (
			leaseKeepRes *clientv3.LeaseKeepAliveResponse
		)
		for {
			select {
			case leaseKeepRes = <-leaseKeepAliveRes:
				if leaseKeepRes == nil {
					//租约以关闭或到期
					return
				}
			}
		}
	}()

	//使用事务抢锁
	txn = this.Kv.Txn(context.TODO())
	//锁的路径
	lockKey = common.JOB_LOCK_DIR + this.JobName
	//事务定义
	txn.If(clientv3.Compare(clientv3.CreateRevision(lockKey), "=", 0)).
		Then(clientv3.OpPut(lockKey, "", clientv3.WithLease(leaseId))).
		Else(clientv3.OpGet(lockKey))
	//事务提交
	if txnRes, err = txn.Commit(); err != nil {
		this.checkError()
		return
	}

	if txnRes.Succeeded {
		//抢锁成功
		this.IsLocked = true
		return
	} else {
		//抢锁失败
		this.IsLocked = false
		err = common.ERR_LOCK_ALREADY_REQUIRED
		this.checkError()
	}

	return
}

//任务解锁
func (this *JobLock) Unlock() {
	if this.IsLocked {
		this.CancelFunc()
		this.Lease.Revoke(context.TODO(), this.LeaseId)
	}
}

//出现错误释放资源
func (this *JobLock) checkError() {
	if this.CancelFunc != nil {
		this.CancelFunc()
	}
	if this.LeaseId > 0 {
		this.Lease.Revoke(context.TODO(), this.LeaseId)
	}
}
