package main

import (
	"fmt"
	"os/exec"
)

func main() {

	cmd := exec.Command("E:\\cygwin\\bin\\bash.exe", "-c", "echo 1")
	err := cmd.Run()
	fmt.Println(err)
}
