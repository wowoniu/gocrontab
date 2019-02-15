package main

import (
	"flag"
	"fmt"
	"gocrontab/crontab/master"
	"runtime"
	"time"
)

/**
master
*/

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
	if err = master.LoadConfig(configFile); err != nil {
		checkErr(err)
		return
	}
	//加载任务管理器
	if err = master.LoadJobMgr(); err != nil {
		checkErr(err)
		return
	}
	//加载日志管理器
	if err = master.LoadJobLog(); err != nil {
		checkErr(err)
		return
	}
	//API SERVER初始化
	if err = master.InitApiServer(); err != nil {
		checkErr(err)
		return
	}

	for {
		time.Sleep(100 * time.Millisecond)
	}

}

func initEnv() {
	//设置线程数
	runtime.GOMAXPROCS(runtime.NumCPU())
	//加载命令行参数
	flag.StringVar(&configFile, "config", "./master.json", "MASTER服务的配置JSON文件")
	flag.Parse()
}

func checkErr(err error) {
	fmt.Println(err)
}
