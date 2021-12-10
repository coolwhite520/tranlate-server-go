package main

import (
	_ "translate-server/datamodels"
	"translate-server/docker"
	"translate-server/http"
	_ "translate-server/logext"
	"translate-server/services"
)

func main()  {
	// 判断是否激活，如果没有激活的话就先不启动docker相关的容器
	service := services.NewActivationService()
	_, state := service.ParseKeystoreFile()
	if state == services.Success {
		err := docker.GetInstance().StartDockers()
		if err != nil {
			panic(err)
		}
	}
	// 启动主要服务
	http.StartMainServer()
}
