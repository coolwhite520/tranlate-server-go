package apis

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"translate-server/apis/mytika"
	"translate-server/config"
)


// TikaParseFile 根据文件路径进行文本提取
func TikaParseFile(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	compList, err := config.GetInstance().GetComponentList(false)
	if err != nil {
		return "", err
	}
	port := "9998"
	for _, v := range compList {
		if v.ImageName == "tk" {
			port = v.HostPort
			break
		}
	}
	url := fmt.Sprintf("http://%s:%s", config.ProxyUrl, port)
	client := mytika.NewClient(nil, url)
	body, err := client.Parse(context.Background(), f)
	if err != nil {
		return "", err
	}
	return body, nil
}
