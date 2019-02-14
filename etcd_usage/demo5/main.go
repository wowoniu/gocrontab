package main

import (
	"context"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"time"
)

/**
分布式事务
*/
func main() {
	var (
		config            clientv3.Config
		client            *clientv3.Client
		err               error
		kv                clientv3.KV
		lease             clientv3.Lease
		leaseGrantRes     *clientv3.LeaseGrantResponse
		leaseID           clientv3.LeaseID
		leaseKeepRes      *clientv3.LeaseKeepAliveResponse
		leaseKeepAliveRes <-chan *clientv3.LeaseKeepAliveResponse
		ctx               context.Context
		cancelFunc        context.CancelFunc
		txn               clientv3.Txn
		txnRes            *clientv3.TxnResponse
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
	//获取租约操作对象
	lease = clientv3.NewLease(client)
	//租约创建 设置5S过期
	if leaseGrantRes, err = lease.Grant(context.TODO(), 5); err != nil {
		fmt.Println("LEASE GRANT ERROR:", err)
		return
	}
	leaseID = leaseGrantRes.ID
	//租约自动续期
	ctx, cancelFunc = context.WithCancel(context.TODO())
	defer lease.Revoke(context.TODO(), leaseGrantRes.ID)
	defer cancelFunc()
	if leaseKeepAliveRes, err = lease.KeepAlive(ctx, leaseGrantRes.ID); err != nil {
		fmt.Println("LEASE KEEP ALIVE ERROR:", err)
		return
	}
	//创建事务  PUT一个有租约的KEY
	txn = kv.Txn(context.TODO())

	//事务定义
	txn.If(clientv3.Compare(clientv3.CreateRevision("/cron/jobs/job1"), "=", 0)).
		Then(clientv3.OpPut("/cron/jobs/job1", "hello", clientv3.WithLease(leaseID))).
		Else(clientv3.OpGet("/cron/jobs/job1"))

	//事务提交
	if txnRes, err = txn.Commit(); err != nil {
		//事务提交失败
		fmt.Println("事务提交失败:", err)
		return
	}
	//判断是否抢到了锁
	if !txnRes.Succeeded {
		fmt.Println("锁被占用:", txnRes.Responses[0].GetResponseRange().Kvs[0].Value)
		return
	}

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

	time.Sleep(30 * time.Second)

}
