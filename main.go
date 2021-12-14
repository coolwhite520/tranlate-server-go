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
	service := services.NewActivationService()
	_, state := service.ParseKeystoreFile()
	if state == datamodels.HttpSuccess {
		//docker.GetInstance().RemoveAllContainer()
		//docker.GetInstance().RemoveImage("48616f72e41a")
		err := docker.GetInstance().StartDockers()
		if err != nil {
			panic(err)
		}
	}
	// 启动主要服务
	http.StartMainServer()
}
