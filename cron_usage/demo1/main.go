package main

import (
	"fmt"
	"github.com/gorhill/cronexpr"
	"time"
)

func main() {
	var (
		exp *cronexpr.Expression
		err error
		now time.Time
	)
	now = time.Now()
	if exp, err = cronexpr.Parse("*/5 * * * * *"); err != nil {
		fmt.Println("cron error:", err)
		return
	}
	fmt.Println(exp.Next(now))
}
