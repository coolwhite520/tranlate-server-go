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
	go func() {
		service := services.NewActivationService()
		_, state := service.ParseKeystoreFile()
		if state == datamodels.HttpSuccess {
			err := docker.GetInstance().StartDockers()
			if err != nil {
				panic(err)
			}
		}
	}()
	// 启动主要服务
	http.StartMainServer()
}
