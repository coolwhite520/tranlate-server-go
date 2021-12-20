package main

import (
	log "github.com/sirupsen/logrus"
	"os/signal"
	"syscall"
	"translate-server/datamodels"
	_ "translate-server/datamodels"
	"translate-server/docker"
	"translate-server/http"
	_ "translate-server/logext"
	"translate-server/services"
)

func main()  {

	//err := docker.GetInstance().StartDefaultWebpageDocker()
	//if err != nil {
	//	log.Fatal(err)
	//	return
	//}
	//log.Println(os.Args)
	go func() {
		// 判断是否激活，如果没有激活的话就先不启动docker相关的容器
		service := services.NewActivationService()
		_, state := service.ParseKeystoreFile()
		if state == datamodels.HttpSuccess {
			docker.GetInstance().SetStatus(docker.RepairingStatus)
			err := docker.GetInstance().StartDockers()
			if err != nil {
				panic(err)
			}
			docker.GetInstance().SetStatus(docker.NormalStatus)
		}
	}()
	// 启动主要服务
	go func() {
		http.StartMainServer()
	}()
	// 信号
	handleSignal()
}

func handleSignal() {
	// 监听信号
	signal.Notify(datamodels.GlobalChannel, syscall.SIGINT, syscall.SIGTERM)
	for {
		sig := <-datamodels.GlobalChannel
		log.Printf("signal receive: %v\n", sig)
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM: // 终止进程执行
			log.Println("shutdown")
			signal.Stop(datamodels.GlobalChannel)
			log.Println("graceful shutdown")
			return
		}
	}
}