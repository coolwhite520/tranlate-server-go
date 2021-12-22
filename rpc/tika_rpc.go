package rpc

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"translate-server/rpc/mytika"
)


// TikaParseFile 根据文件路径进行文本提取
func TikaParseFile(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	fmt.Println(f.Name())
	client := mytika.NewClient(nil, "http://localhost:9998")
	body, err := client.Parse(context.Background(), f)
	if err != nil {
		return "", err
	}
	return body, nil
}
