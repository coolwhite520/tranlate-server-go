package apis

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"translate-server/config"
)



func PyTransSpecialFile(rowId int64, srcFile, desFile, srcLang, desLang string) error {
	compList, err := config.GetInstance().GetComponentList(false)
	if err != nil {
		return err
	}
	port := "5001"
	for _, v := range compList {
		if v.ImageName == "plugins" {
			port = v.HostPort
			break
		}
	}
	url := fmt.Sprintf("http://%s:%s/trans_file", config.ProxyUrl, port)
	client := &http.Client{}
	var req *http.Request

	var bodyData struct {
		RowId     int64  `json:"row_id"`
		SrcFile   string `json:"src_file"`
		DesFile   string `json:"des_file"`
		SrcLang   string `json:"src_lang"`
		DesLang   string `json:"des_lang"`
	}
	bodyData.RowId = rowId
	bodyData.SrcLang = srcLang
	bodyData.DesLang = desLang
	bodyData.SrcFile, _ = filepath.Abs(srcFile)
	bodyData.DesFile, _ = filepath.Abs(desFile)

	data, _ := json.Marshal(bodyData)
	req, err = http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		log.Errorln(err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	var a struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data string `json:"data"`
	}
	err = json.Unmarshal(body, &a)
	if err != nil {
		return err
	}
	if a.Code != 200 {
		return errors.New(a.Msg)
	}
	return nil
}


func PyConvertSpecialFile(srcFile, convertType string) error {
	compList, err := config.GetInstance().GetComponentList(false)
	if err != nil {
		return err
	}
	port := "5001"
	for _, v := range compList {
		if v.ImageName == "plugins" {
			port = v.HostPort
			break
		}
	}
	url := fmt.Sprintf("http://%s:%s/convert_file", config.ProxyUrl, port)
	client := &http.Client{}
	var req *http.Request

	var bodyData struct {
		SrcFile   string `json:"src_file"`
		DesFile   string `json:"des_file"`
		ConvertType      string `json:"convert_type"`
	}
	bodyData.SrcFile, _ = filepath.Abs(srcFile)
	bodyData.ConvertType = convertType

	data, _ := json.Marshal(bodyData)
	req, err = http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		log.Errorln(err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	var a struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data string `json:"data"`
	}
	err = json.Unmarshal(body, &a)
	if err != nil {
		return err
	}
	if a.Code != 200 {
		return errors.New(a.Msg)
	}
	return nil
}
