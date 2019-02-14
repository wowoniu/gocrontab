package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"time"
)

func main() {

	var (
		config      clientv3.Config
		client      *clientv3.Client
		err         error
		kv          clientv3.KV
		putResponse *clientv3.PutResponse
	)
	config = clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	}
	//建立一个客户端连接
	if client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
		return
	}
	//获取KV操作对象
	kv = clientv3.NewKV(client)

	if putResponse, err = kv.Put(context.TODO(), "/cron/jobs/job1", "hello"); err != nil {
		fmt.Println("PUT KV ERROR:", err)
		return
	}
	fmt.Println(putResponse.Header.Revision)

}
