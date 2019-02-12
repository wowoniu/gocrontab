package main

import (
	"context"
	"fmt"
	"os/exec"
)

type result struct {
	err    error
	output []byte
}

//通过上下文环境来控制命令的执行取消
func main() {
	var (
		ctx context.Context
		//cancelFunc context.CancelFunc
		cmd        *exec.Cmd
		err        error
		resultChan chan *result
		res        *result
	)

	resultChan = make(chan *result, 1)
	//创建一个上下文环境 和一个取消上下文执行的方法
	ctx, _ = context.WithCancel(context.TODO())

	//开启协程执行命令
	go func() {
		var (
			output []byte
		)
		cmd = exec.CommandContext(ctx, "E:\\cygwin\\bin\\bash.exe", "-c", "/usr/bin/sleep 5;/usr/bin/ls -l")
		if output, err = cmd.CombinedOutput(); err != nil {
			//fmt.Println("CMD ERROR:",err,string(output))
			resultChan <- &result{
				err,
				output,
			}
			return
		}
		resultChan <- &result{
			nil,
			output,
		}
	}()

	//1s后 取消命令行的执行
	//time.Sleep(1*time.Second)

	//cancelFunc()

	res = <-resultChan

	fmt.Println(res.err, string(res.output))

}
