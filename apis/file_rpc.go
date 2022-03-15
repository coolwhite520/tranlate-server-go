package apis

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"translate-server/config"
)

func RpcTransFile(rowId int64, dataAbsPath string) error {
	compList, err := config.GetInstance().GetComponentList(false)
	if err != nil {
		return err
	}
	port := "5001"
	for _, v := range compList {
		if v.ImageName == "file" {
			port = v.HostPort
			break
		}
	}
	url := fmt.Sprintf("http://%s:%s/trans_file", config.ProxyUrl, port)
	client := &http.Client{}
	var req *http.Request

	var bodyData struct {
		Id     int64  `json:"id"`
		DataAbsPath string `json:"dataAbsPath"`
	}
	bodyData.Id = rowId
	bodyData.DataAbsPath = dataAbsPath


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
		Message  string `json:"message"`
		Data string `json:"data"`
	}
	err = json.Unmarshal(body, &a)
	if err != nil {
		return err
	}
	if a.Code != 200 {
		return errors.New(a.Message)
	}
	return nil
}
