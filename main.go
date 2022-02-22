package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"os"
	"time"
	"translate-server/config"
	"translate-server/constant"
	"translate-server/datamodels"
	"translate-server/docker"
	_ "translate-server/logext"
	"translate-server/server"
	_ "translate-server/structs"
)

var (
	srv      *http.Server
	listener net.Listener
	graceful = flag.Bool("graceful", false, "listen on fd open 3 (internal use only)")
)

func init()  {
	//config.GetInstance().TestGenerateConfigFile()
	//return
	err := docker.GetInstance().CreatePrivateNetwork()
	if err != nil {
		panic(err)
	}
	docker.GetInstance().SetStatus(docker.RepairingStatus)
	// StartDockers 内部会判断是否已经是激活的状态
	err = docker.GetInstance().StartDockers()
	if err != nil {
		panic(err)
	}
	docker.GetInstance().SetStatus(docker.NormalStatus)
	datamodels.InitMysql()
	datamodels.InitRedis()
	err = config.InitConfigIniFile()
	if err != nil {
		panic(err)
	}
}

func main() {
	go func() {
		model := datamodels.NewActivationModel()
		for  {
			// 每隔10分钟，减少一下剩余可用时间
			time.Sleep(time.Minute * 10 )
			expiredInfo, state:= model.ParseExpiredFile()
			if state == constant.HttpSuccess {
				expiredInfo.LeftTimeSpan = expiredInfo.LeftTimeSpan - 10 * 60
				model.GenerateExpiredFile(*expiredInfo)
			}
		}
	}()
	srv = &http.Server{Addr: ":7777"}
	log.Println("server will listening on : http://localhost:7777")
	var err error
	if *graceful {
		log.Print("main: Listening to existing file descriptor 3.")
		f := os.NewFile(3, "")
		listener, err = net.FileListener(f)
	} else {
		log.Print("main: Listening on a new file descriptor.")
		listener, err = net.Listen("tcp", srv.Addr)
	}
	if err != nil {
		log.Fatalf("listener error: %v", err)
	}
	server.StartMainServer(listener)
}


