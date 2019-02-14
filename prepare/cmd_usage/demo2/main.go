package main

import (
	"fmt"
	"os/exec"
)

func main() {
	var (
		cmd    *exec.Cmd
		output []byte
		err    error
	)

	cmd = exec.Command("E:\\cygwin\\bin\\bash.exe", "-c", "/usr/bin/sleep 5;/usr/bin/ls -l")

	if output, err = cmd.CombinedOutput(); err != nil {
		fmt.Println("CMD ERROR:", err, string(output))
		return
	}

	fmt.Println(string(output))

}
