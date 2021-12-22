package rpc

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"translate-server/config"
	"translate-server/rpc/mytika"
)


// TikaParseFile 根据文件路径进行文本提取
func TikaParseFile(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	systemConfig, err := config.GetInstance().ParseSystemConfigFile(false)
	if err != nil {
		return "", err
	}
	var port string
	for _, v := range systemConfig.ComponentList {
		if v.ImageName == "tk" {
			port = v.HostPort
			break
		}
	}
	url := fmt.Sprintf("http://localhost:%s", port)
	client := mytika.NewClient(nil, url)
	body, err := client.Parse(context.Background(), f)
	if err != nil {
		return "", err
	}
	return body, nil
}
