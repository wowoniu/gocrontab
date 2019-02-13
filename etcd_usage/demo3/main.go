package main

/**
租约相关操作
*/

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"time"
)

func main() {
	var (
		config        clientv3.Config
		client        *clientv3.Client
		err           error
		kv            clientv3.KV
		putRes        *clientv3.PutResponse
		lease         clientv3.Lease
		leaseGrantRes *clientv3.LeaseGrantResponse
	)

	config = clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	}
	if client, err = clientv3.New(config); err != nil {
		fmt.Println("CONNECT ETCD ERROR:", err)
		return
	}
	kv = clientv3.NewKV(client)
	if putRes, err = kv.Put(context.TODO(), "/cron/jobs/job1", "hello"); err != nil {
		fmt.Println("PUT KEY ERROR:", err)
		return
	}
	fmt.Println("PUT KEY:", putRes.Header.Revision)

	//获取租约操作对象
	lease = clientv3.NewLease(client)
	//租约创建 设置5S过期
	if leaseGrantRes, err = lease.Grant(context.TODO(), 5); err != nil {
		fmt.Println("LEASE GRANT ERROR:", err)
		return
	}
	defer lease.Revoke(context.TODO(), leaseGrantRes.ID)
	fmt.Println("LEASE GRANT SUCCESS:", leaseGrantRes.ID)

	//put key 与租约绑定
	if putRes, err = kv.Put(context.TODO(), "/cron/jobs/job2", "world", clientv3.WithLease(leaseGrantRes.ID)); err != nil {
		fmt.Println("LEASES BIND ERROR:", err)
		return
	}

	go func() {
		//每2s读取一次设置过租约的值
		var (
			getRes *clientv3.GetResponse
			err    error
		)
		for {
			select {
			case <-time.Tick(2 * time.Second):
				if getRes, err = kv.Get(context.TODO(), "/cron/jobs/job2"); err != nil {
					fmt.Println("定时读取失败")
					return
				}
				if len(getRes.Kvs) == 0 {
					fmt.Println("租约过期")
					return
				} else {
					fmt.Println("定时读取:", getRes.Kvs)
				}
			}
		}
	}()

	time.Sleep(20 * time.Second)
}
