package rpc

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"translate-server/rpc/mytika"
)

const ExternalPort = "9998"

// StartTikaServer 启动服务tika的docker服务，并返回容器ID
func StartTikaServer() {
	params := fmt.Sprintf("docker run -d -p 9998:%s apache/tika", ExternalPort)
	cmd := exec.Command("bash", "-c", params)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal("StartTikaServer:", err)
	}
	defer stdout.Close()
	if err := cmd.Start(); err != nil {
		log.Fatal("StartTikaServer:",err)
	}
	opBytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		log.Fatal("StartTikaServer:",err)
	}
	log.Info("StartTikaServer start success ! Id:", string(opBytes))
}

// TikaParseFile 根据文件路径进行文本提取
func TikaParseFile(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	fmt.Println(f.Name())
	client := mytika.NewClient(nil, "http://localhost:" + ExternalPort)
	body, err := client.Parse(context.Background(), f)
	if err != nil {
		return "", err
	}
	return body, nil
}
