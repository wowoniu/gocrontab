package worker

import (
	"context"
	"os/exec"
)

type ShellExecutor struct {
	Shell   string
	Command string
}

var G_shellExecutor *ShellExecutor

func init() {
	G_shellExecutor = &ShellExecutor{}
}

//执行命令
func (this *ShellExecutor) Exec(ctx context.Context, command string) (output []byte, err error) {
	var (
		cmd *exec.Cmd
	)
	//构造cmd
	cmd = exec.CommandContext(ctx, G_config.ExecuteShell, "-c", command)
	//命令执行 并获取输出
	output, err = cmd.CombinedOutput()
	return
}
