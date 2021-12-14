package rpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strings"
)

func OcrParseFile(filePathName string) (string, error) {
	base := path.Base(filePathName)
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	fw, err := w.CreateFormFile("image", base)
	if err != nil {
		log.Error(err)
		return "", err
	}
	f, err := os.Open(filePathName)
	if err != nil {
		log.Error(err)
		return "", err
	}
	bin, err := ioutil.ReadAll(f)
	if err != nil {
		log.Error(err)
		return "", err
	}
	_, err = fw.Write(bin)
	if err != nil {
		log.Error(err)
		return "", err
	}
	w.Close()

	req, err := http.NewRequest("POST", "http://localhost:9090/upload", buf)
	if err != nil {
		log.Error("req err: ", err)
		return "", err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
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
	all := strings.ReplaceAll(a.Content, " ", "")
	return all, nil
}

