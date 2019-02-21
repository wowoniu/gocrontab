package worker

import (
	"context"
	"encoding/json"
	"go.etcd.io/etcd/clientv3"
	"gocrontab/crontab/common"
	"math/rand"
	"time"
)

type Register struct {
	Client     *clientv3.Client
	Kv         clientv3.KV
	Lease      clientv3.Lease
	LeaseId    clientv3.LeaseID
	CancelCtx  context.Context
	CancelFunc context.CancelFunc
	Worker     *common.Worker
}

var G_register *Register

func LoadRegister() (err error) {
	var (
		config clientv3.Config
		client *clientv3.Client
	)
	G_register = &Register{}
	//连接ETCD
	config = clientv3.Config{
		Endpoints:   G_config.EtcdEndpoints,
		DialTimeout: time.Duration(G_config.EtcdDialTimeout) * time.Microsecond,
	}
	if client, err = clientv3.New(config); err != nil {
		return
	}
	G_register.Client = client
	G_register.Kv = clientv3.NewKV(client)
	G_register.Lease = clientv3.NewLease(client)

	G_register.Register()
	return
}

//服务注册
func (this *Register) Register() (err error) {
	var (
		workerIp              string
		leaseGrantRes         *clientv3.LeaseGrantResponse
		leaseId               clientv3.LeaseID
		leaseKeepAliveResChan <-chan *clientv3.LeaseKeepAliveResponse
		cancelCtx             context.Context
		cancelFunc            context.CancelFunc
		registerKey           string
		workerData            string
		workerFlag            string
	)
	//获取本机IP
	if workerIp, err = common.GetLocalIp(); err != nil {
		return
	}
	//创建租约
	if leaseGrantRes, err = this.Lease.Grant(context.TODO(), 5); err != nil {
		return
	}
	leaseId = leaseGrantRes.ID
	this.LeaseId = leaseId
	//自动续租
	cancelCtx, cancelFunc = context.WithCancel(context.TODO())
	if leaseKeepAliveResChan, err = this.Lease.KeepAlive(cancelCtx, leaseId); err != nil {
		return
	}
	this.CancelCtx = cancelCtx
	this.CancelFunc = cancelFunc
	//开启协程 监听自动续租响应
	go func() {
		var (
			leaseKeepAliveRes *clientv3.LeaseKeepAliveResponse
		)
		for {
			for leaseKeepAliveRes = range leaseKeepAliveResChan {
				if leaseKeepAliveRes == nil {
					//续约终止
					return
				}
			}
		}
	}()

	//服务标记注册
	workerFlag = this.createWorkerFlag()
	registerKey = common.JOB_REGISTER_DIR + workerIp + ":" + workerFlag

	this.Worker = &common.Worker{
		Name:      G_config.WorkerName,
		Ip:        workerIp,
		Flag:      workerFlag,
		StartTime: time.Now().Unix(),
	}
	if G_config.WorkerGroup == "" {
		this.Worker.Group = common.DEFAULT_WORKER_GROUP_NAME
	} else {
		this.Worker.Group = G_config.WorkerGroup
	}

	//序列化
	if workerData, err = this.buildWorkerData(this.Worker); err != nil {
		return
	}
	if _, err = this.Kv.Put(context.TODO(), registerKey, workerData, clientv3.WithLease(leaseId)); err != nil {
		//注册失败
		cancelFunc()
		this.Lease.Revoke(context.TODO(), leaseId)
		return
	}
	return

}

func (this *Register) buildWorkerData(worker *common.Worker) (workerData string, err error) {
	var (
		workerJsonByte []byte
	)
	//序列化
	if workerJsonByte, err = json.Marshal(worker); err != nil {
		return
	}
	workerData = string(workerJsonByte)
	return
}

//生成32位长度的节点随机标识
func (this *Register) createWorkerFlag() string {
	str := "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 32; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}
