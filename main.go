package main

import (
	"context"
	"errors"
	"flag"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
	"translate-server/datamodels"
	_ "translate-server/datamodels"
	"translate-server/docker"
	_ "translate-server/logext"
	"translate-server/server"
	"translate-server/services"
)
var (
	srv   *http.Server
	listener net.Listener
	graceful = flag.Bool("graceful", false, "listen on fd open 3 (internal use only)")
)

func main()  {
	go func() {
		// 不管激活不激活都要启动web的docker镜像，否则用户没有页面可以访问
		err := docker.GetInstance().StartDefaultWebpageDocker()
		if err != nil {
			log.Fatal(err)
			return
		}
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
	log.Println(os.Args)
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

	go func() {
		server.StartMainServer(listener)
	}()
	// 信号
	signalHandler()
}

func reload() error {
	tl, ok := listener.(*net.TCPListener)
	if !ok {
		return errors.New("listener is not tcp listener")
	}
	f, err := tl.File()
	if err != nil {
		return err
	}
	args := []string{"-graceful"}
	cmd := exec.Command(os.Args[0], args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// put socket FD at the first entry
	cmd.ExtraFiles = []*os.File{f}
	return cmd.Start()
}

func signalHandler() {
	signal.Notify(datamodels.GlobalChannel, syscall.SIGINT, syscall.SIGTERM,  syscall.SIGQUIT)
	for {
		sig := <-datamodels.GlobalChannel
		log.Printf("signal: %v", sig)

		// timeout context for shutdown
		ctx, _ := context.WithTimeout(context.Background(), 10 * time.Second)
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			// stop
			log.Printf("stop")
			signal.Stop(datamodels.GlobalChannel)
			srv.Shutdown(ctx)
			log.Printf("graceful shutdown")
			return
		case syscall.SIGQUIT:
			// reload
			log.Printf("reload")
			err := reload()
			if err != nil {
				log.Fatalf("graceful restart error: %v", err)
			}
			srv.Shutdown(ctx)
			log.Printf("graceful reload")
			return
		}
	}
}