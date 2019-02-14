package main

import (
	"flag"
	"fmt"
	"gocrontab/crontab/worker"
	"runtime"
	"time"
)

var (
	configFile string
)

func main() {
	var (
		err error
	)
	//初始化环境设置
	initEnv()
	//加载配置
	if err = worker.LoadConfig(configFile); err != nil {
		checkErr(err)
		return
	}
	//加载任务管理器
	if err = worker.LoadJobMgr(); err != nil {
		checkErr(err)
		return
	}

	//监听任务
	go worker.G_jobmgr.WatchJobs()

	for {
		time.Sleep(1 * time.Second)
	}

}

func initEnv() {
	//设置线程数
	runtime.GOMAXPROCS(runtime.NumCPU())
	//加载命令行参数
	flag.StringVar(&configFile, "config", "./worker.json", "worker的配置JSON文件")
	flag.Parse()
}

func checkErr(err error) {
	fmt.Println(err)
}
