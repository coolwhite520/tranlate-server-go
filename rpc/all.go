package rpc

import (
	log "github.com/sirupsen/logrus"
	"os/exec"
)

// StopAllServer 启动服务tika的docker服务，并返回容器ID
//docker stop $(docker ps -a | awk '{ print $1}' | tail -n +2) 可以关闭所有服务
func StopAllServer()  {
	cmd := exec.Command("bash","-c", "docker stop $(docker ps -a | awk '{ print $1}' | tail -n +2)")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal("StopAllServer:", err)
	}
	defer stdout.Close()
	if err = cmd.Start(); err != nil {
		log.Fatal("StopAllServer:", err)
	}
	err = cmd.Wait()
	if err != nil {
		log.Fatal("StopAllServer:", err)
	}
}