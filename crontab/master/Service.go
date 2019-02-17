package master

import (
	"context"
	"encoding/json"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"gocrontab/crontab/common"
	"time"
)

type Service struct {
	Client *clientv3.Client
	Kv     clientv3.KV
}

var G_service *Service

func LoadService() (err error) {
	var (
		config clientv3.Config
		client *clientv3.Client
	)
	G_service = &Service{}
	//连接ETCD
	config = clientv3.Config{
		Endpoints:   G_config.EtcdEndpoints,
		DialTimeout: time.Duration(G_config.EtcdDialTimeout) * time.Microsecond,
	}
	if client, err = clientv3.New(config); err != nil {
		return
	}
	G_service.Client = client
	G_service.Kv = clientv3.NewKV(client)

	return
}

//获取健康的节点
func (this *Service) GetWorkerList() (workerList []*common.Worker, count int64, err error) {
	var (
		getRes   *clientv3.GetResponse
		keyValue *mvccpb.KeyValue
		worker   *common.Worker
	)
	workerList = make([]*common.Worker, 0)
	if getRes, err = this.Kv.Get(context.TODO(), common.JOB_REGISTER_DIR, clientv3.WithPrefix()); err != nil {
		return
	}

	count = getRes.Count

	for _, keyValue = range getRes.Kvs {
		//反序列化
		worker = &common.Worker{}
		if err = json.Unmarshal(keyValue.Value, worker); err != nil {
			//节点数据异常
			err = nil
			continue
		}
		workerList = append(workerList, worker)
	}

	return
}
