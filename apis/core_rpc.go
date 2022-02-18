package apis

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"
	"translate-server/config"
	"translate-server/utils"
)
var SignKey = "Today I want to eat noodle."

func PyTranslate(srcLang, desLang, content string) (string, error) {
	compList, err := config.GetInstance().GetComponentList(false)
	if err != nil {
		return "", err
	}
	port := "5000"
	for _, v := range compList {
		if v.ImageName == "core" {
			port = v.HostPort
			break
		}
	}
	url := fmt.Sprintf("http://%s:%s/translate", config.ProxyUrl, port)
	client := &http.Client{}
	var req *http.Request

	var bodyData struct {
		SrcLang   string `json:"src_lang"`
		DesLang   string `json:"des_lang"`
		Content   string `json:"content"`
		Timestamp int64  `json:"timestamp"`
		Sign      string `json:"sign"`
	}
	bodyData.SrcLang = srcLang
	bodyData.DesLang = desLang
	bodyData.Content = content
	bodyData.Timestamp = time.Now().Unix()

	s := fmt.Sprintf("src_lang=%s&des_lang=%s&content=%s&timestamp=%v", bodyData.SrcLang, bodyData.DesLang, bodyData.Content, bodyData.Timestamp)
	sign := utils.GenerateHmacSign([]byte(s), []byte(SignKey))
	bodyData.Sign = sign
	data, _ := json.Marshal(bodyData)
	req, err = http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		log.Errorln(err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err.Error())
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
		return "", err
	}
	var a struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data string `json:"data"`
	}
	err = json.Unmarshal(body, &a)
	if err != nil {
		return "", err
	}
	if a.Code != 200 {
		return "", errors.New(a.Msg)
	}
	return a.Data, nil
}

// PyTokenize 句子切分
func PyTokenize(srcLang, content string) ([]string, error) {
	compList, err := config.GetInstance().GetComponentList(false)
	if err != nil {
		return nil, err
	}
	port := "5000"
	for _, v := range compList {
		if v.ImageName == "core" {
			port = v.HostPort
			break
		}
	}
	url := fmt.Sprintf("http://%s:%s/tokenize", config.ProxyUrl, port)
	client := &http.Client{}
	var req *http.Request

	var bodyData struct {
		SrcLang   string `json:"src_lang"`
		Content   string `json:"content"`
	}
	bodyData.SrcLang = srcLang
	bodyData.Content = content
	data, _ := json.Marshal(bodyData)
	req, err = http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	var a struct {
		Code int    `json:"code"`
		List  []string `json:"list"`
		Len int `json:"len"`
		Msg string `json:"msg"`
	}
	err = json.Unmarshal(body, &a)
	if err != nil {
		return nil, err
	}
	if a.Code != 200 {
		return nil, errors.New(a.Msg)
	}
	return a.List, nil
}
