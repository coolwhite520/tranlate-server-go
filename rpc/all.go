package rpc

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os/exec"
	"strings"
)
/**
其中字符串为容器ID:
docker exec -it d27bd3008ad9 /bin/bash

1.停用全部运行中的容器:
docker stop $(docker ps -q)

2.删除全部容器：
docker rm $(docker ps -aq)
 */

type DockerInfo struct {
	DockerName string `json:"docker_name"`
	FilePathName string `json:"file_path_name"`
	RumCmdLine string `json:"cmd_line"`
}

var DockerList []DockerInfo

func init()  {
	DockerList = append(DockerList, DockerInfo{
		DockerName: "apache/tika",
		FilePathName:   "./tika.tar",
		RumCmdLine: fmt.Sprintf("docker run -d -p 9998:%s apache/tika", TikaExternalPort),
	})
}

func ImportAllImages() {
	log.Info("ImportAllImages begin...")
	for _,v := range DockerList{
		if !isHavingImage(v.DockerName) {
			importImage(v.FilePathName)
			if isHavingImage(v.DockerName) {
				log.Info("importImage ", v.DockerName, "success")
			} else {
				log.Fatal("importImage", v.DockerName, "failed")
			}
		}
	}
	log.Info("ImportAllImages end...")
}

// StopAllRunningDockers 停止
//docker stop $(docker ps -a | awk '{ print $1}' | tail -n +2) 可以关闭所有服务
func StopAllRunningDockers()  {
	if isHavingDockers() {
		runCmd("docker stop $(docker ps -a | awk '{ print $1}' | tail -n +2)")
	}
}

// ImportImage 根据文件路径导入镜像
func importImage(filePathName string) bool {
	log.Info("importImage ", filePathName)
	cmdLine := fmt.Sprintf("docker load < %s", filePathName)
	runCmd(cmdLine)
	return true
}

// IsHavingImage 查看是否存在某个镜像
func isHavingImage(imageName string) bool {
	log.Info("isHavingImage ", imageName)
	cmdLine := fmt.Sprintf("docker images | awk '{ print $1}'")
	s := runCmd(cmdLine)
	arrays := strings.Split(s, "\n")
	for _, v := range arrays {
		if strings.Contains(v, imageName) {
			log.Info("isHavingImage ", imageName, " true")
			return true
		}
	}
	log.Info("isHavingImage ", imageName, " false")
	return false
}

// isHavingDockers 查看是否存在容器
func isHavingDockers() bool {
	output := runCmd("docker ps -a | awk '{ print $1}' | tail -n +2")
	if len(output) > 0 {
		return true
	}
	return false
}

// runCmd 返回标准输出的内容
func runCmd(cmdLine string) string {
	cmd := exec.Command("bash","-c", cmdLine)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal("isHavingDockers: " + err.Error())
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal("isHavingDockers: " + err.Error())
	}
	if err := cmd.Start(); err != nil {
		log.Fatal("isHavingDockers: " + err.Error())
	}
	bytesErr, err := ioutil.ReadAll(stderr)
	if err != nil {
		log.Fatal("isHavingDockers: " + err.Error())
	}
	if len(bytesErr) != 0 {
		log.Fatal("isHavingDockers: " + err.Error())
	}
	bytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		log.Fatal("isHavingDockers: " + err.Error())
	}
	if err := cmd.Wait(); err != nil {
		log.Fatal("isHavingDockers: " + err.Error())
	}
	return string(bytes)
}
