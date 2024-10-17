package main

import (
	"fun/worker"
	"runtime"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
)

func main() {
	logx.Info("Hello world!")

	wp := worker.NewWorkerPool(runtime.NumCPU(), 1000)

	sg := service.NewServiceGroup()
	defer sg.Stop()

	sg.Add(wp)
	sg.Start()

}
