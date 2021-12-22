package rpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"translate-server/config"
	"unicode"
)
func IsChineseChar(str string) bool {
	for _, r := range str {
		if unicode.Is(unicode.Scripts["Han"], r) {
			return true
		}
	}
	return false
}
func  postWithMultiPartData(url string, body io.Reader, filename string) (resp *http.Response, err error) {
	var buffer = new(bytes.Buffer)
	var writer  = multipart.NewWriter(buffer)
	w, err := writer.CreateFormFile("image", filename)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(w, body)
	if err != nil {
		return nil, err
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	client:=&http.Client{}
	return client.Post(url, writer.FormDataContentType(), buffer)
}
func OcrParseFile(filePathName string) (string, error) {
	info, err := os.Stat(filePathName)
	if err != nil {
		log.Error(err)
		return "", err
	}
	f, err := os.Open(filePathName)
	if err != nil {
		log.Error(err)
		return "", err
	}
	systemConfig, err := config.GetInstance().ParseSystemConfigFile(false)
	if err != nil {
		return "", err
	}
	var port string
	for _, v := range systemConfig.ComponentList {
		if v.ImageName == "tesseract" {
			port = v.HostPort
			break
		}
	}
	url := fmt.Sprintf("http://localhost:%s/upload", port)
	resp, err := postWithMultiPartData(url, f, info.Name())
	if err != nil {
		log.Error("resp err: ", err)
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", errors.New("resp status:" + fmt.Sprint(resp.StatusCode))
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err.Error())
		return "", err
	}

	var a struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Content string `json:"content"`
	}
	err = json.Unmarshal(body, &a)
	if err != nil {
		return "", err
	}
	if a.Code != 200 {
		return "", errors.New(a.Msg)
	}
	var all string
	if IsChineseChar(a.Content) {
		all = strings.ReplaceAll(a.Content, " ", "")
	} else {
		all = a.Content
	}

	return all, nil
}

