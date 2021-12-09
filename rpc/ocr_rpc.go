package rpc

import (
	"github.com/otiai10/gosseract/v2"
)
func OcrParseFile(filePath string) (string,error) {
	client := gosseract.NewClient()
	defer client.Close()
	client.SetImage(filePath)
	return  client.Text()
}
