package main

import (
	"translate-server/datamodels"
	_ "translate-server/datamodels"
	"translate-server/docker"
	"translate-server/http"
	_ "translate-server/logext"
	"translate-server/services"
)

func main()  {
	// 判断是否激活，如果没有激活的话就先不启动docker相关的容器
	//err := docker.GetInstance().StartDefaultWebpageDocker()
	//if err != nil {
	//	log.Fatal(err)
	//	return
	//}
	go func() {
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
	http.StartMainServer()
}
