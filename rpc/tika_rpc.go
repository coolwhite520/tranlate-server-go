package rpc

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"translate-server/rpc/mytika"
)

const TikaExternalPort = "9998"


// TikaParseFile 根据文件路径进行文本提取
func TikaParseFile(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	fmt.Println(f.Name())
	client := mytika.NewClient(nil, "http://localhost:" + TikaExternalPort)
	body, err := client.Parse(context.Background(), f)
	if err != nil {
		return "", err
	}
	return body, nil
}
