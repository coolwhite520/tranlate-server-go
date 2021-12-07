package utils

import (
	"net/http"
	"os"
)

func GetFileContentType(filepath string) (string, error) {
	ff, err := os.Open(filepath)
	if err != nil {
		panic(err)
	}
	defer ff.Close()

	buffer := make([]byte, 512)
	_, err = ff.Read(buffer)
	if err != nil {
		return "", err
	}
	contentType := http.DetectContentType(buffer)
	return contentType, nil
}

