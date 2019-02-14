package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"time"
)

/**
租约自动续期
*/
func main() {
	var (
		config            clientv3.Config
		client            *clientv3.Client
		err               error
		kv                clientv3.KV
		putRes            *clientv3.PutResponse
		lease             clientv3.Lease
		leaseGrantRes     *clientv3.LeaseGrantResponse
		leaseKeepRes      *clientv3.LeaseKeepAliveResponse
		leaseKeepAliveRes <-chan *clientv3.LeaseKeepAliveResponse
		ctx               context.Context
		cancelFunc        context.CancelFunc
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

	fmt.Println("LEASE GRANT SUCCESS:", leaseGrantRes.ID)

	//put key 与租约绑定
	if putRes, err = kv.Put(context.TODO(), "/cron/jobs/job2", "world", clientv3.WithLease(leaseGrantRes.ID)); err != nil {
		fmt.Println("LEASES BIND ERROR:", err)
		return
	}
	//租约自动续期
	ctx, cancelFunc = context.WithCancel(context.TODO())
	if leaseKeepAliveRes, err = lease.KeepAlive(ctx, leaseGrantRes.ID); err != nil {
		fmt.Println("LEASE KEEP ALIVE ERROR:", err)
		return
	}
	defer lease.Revoke(context.TODO(), leaseGrantRes.ID)
	defer cancelFunc()

	//接收续约的状态
	//leaseKeepRes=leaseKeepRes
	//leaseKeepAliveRes=leaseKeepAliveRes

	go func() {
		var (
			now time.Time
		)
		for {
			now = time.Now()
			select {
			case leaseKeepRes = <-leaseKeepAliveRes:
				if leaseKeepRes != nil {
					fmt.Println(now, "续约成功:", leaseKeepRes.ID)
				} else {
					fmt.Println(now, "续约取消")
					return
				}
			}
		}
	}()
	//每2s读取一次设置过租约的值
	go func() {
		var (
			getRes *clientv3.GetResponse
			err    error
			now    time.Time
		)
		for {
			now = time.Now()
			select {
			case <-time.Tick(2 * time.Second):
				if getRes, err = kv.Get(context.TODO(), "/cron/jobs/job2"); err != nil {
					fmt.Println(now, "定时读取失败")
					return
				}
				if len(getRes.Kvs) == 0 {
					fmt.Println(now, "定时读取[租约过期]")
					return
				} else {
					fmt.Println(now, "定时读取:", getRes.Kvs)
				}
			}
		}
	}()

	//10S后取消自动续约
	time.AfterFunc(10*time.Second, func() {
		cancelFunc()
	})
	time.Sleep(30 * time.Second)
}
