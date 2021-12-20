package main

import (
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
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
	log.Println(os.Args)
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
	signal.Notify(datamodels.GlobalChannel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2)
	for {
		sig := <-datamodels.GlobalChannel
		log.Printf("signal receive: %v\n", sig)
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM: // 终止进程执行
			log.Println("shutdown")
			signal.Stop(datamodels.GlobalChannel)
			log.Println("graceful shutdown")
			return
		case syscall.SIGUSR2: // 进程热重启
			log.Println("reload")
			err := reload() // 执行热重启函数
			if err != nil {
				log.Fatalf("graceful reload error: %v", err)
			}
			log.Println("graceful reload")
			return
		}
	}
}

func reload() error {
	args := []string{
		"-graceful"}
	cmd := exec.Command(os.Args[0], args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// 新建并执行子进程
	return cmd.Start()
}