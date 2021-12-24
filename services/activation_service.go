package services

import (
	"encoding/base64"
	"encoding/json"
	"github.com/denisbrodbeck/machineid"
	"io/ioutil"
	"time"
	"translate-server/datamodels"
	"translate-server/utils"
)

type ActivationService interface {
	GenerateMachineId() (string, datamodels.HttpStatusCode)
	GenerateKeystoreContent(activationInfo datamodels.Activation) (string, datamodels.HttpStatusCode)
	GenerateKeystoreFile(datamodels.Activation) datamodels.HttpStatusCode
	GenerateKeystoreFileByContent(string) datamodels.HttpStatusCode
	ParseKeystoreContent(content string) (*datamodels.Activation, datamodels.HttpStatusCode)
	ParseKeystoreFile() (*datamodels.Activation, datamodels.HttpStatusCode)
}

func NewActivationService() ActivationService {
	return &activation{}
}

const KeyStorePath = "./keystore"
const AppID = "@My_TrAnSLaTe_sErVeR"



type activation struct {
}

// GenerateMachineId 生成机器码
func (a *activation) GenerateMachineId() (string, datamodels.HttpStatusCode) {
	id, err := machineid.ProtectedID(AppID)
	if err != nil {
		return "", datamodels.HttpActivationGenerateError
	}
	return id, datamodels.HttpSuccess
}

func (a activation) GenerateKeystoreContent(activationInfo datamodels.Activation) (string, datamodels.HttpStatusCode) {
	data, err := json.Marshal(activationInfo)
	if err != nil {
		return "", datamodels.HttpActivationGenerateError
	}
	v := utils.Md5V(activationInfo.Sn + AppID)
	encrypt, err := utils.AesEncrypt(data, []byte(v))
	if err != nil {
		return "", datamodels.HttpActivationAESError
	}
	toString := base64.StdEncoding.EncodeToString(encrypt)
	return toString, datamodels.HttpSuccess
}


func (a *activation) GenerateKeystoreFile(activationInfo datamodels.Activation) datamodels.HttpStatusCode {
	content, status := a.GenerateKeystoreContent(activationInfo)
	if status != datamodels.HttpSuccess {
		return status
	}
	ioutil.WriteFile(KeyStorePath, []byte(content), 0666)
	return datamodels.HttpSuccess
}

func (a *activation) GenerateKeystoreFileByContent(content string) datamodels.HttpStatusCode {
	ioutil.WriteFile(KeyStorePath, []byte(content), 0666)
	return datamodels.HttpSuccess
}

func (a *activation) ParseKeystoreContent(content string) (*datamodels.Activation, datamodels.HttpStatusCode) {
	id, status := a.GenerateMachineId()
	if status != datamodels.HttpSuccess {
		return nil, status
	}
	v := utils.Md5V(id + AppID)
	base64Decode, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return nil, datamodels.HttpActivationInvalidateError
	}
	decrypt, err := utils.AesDecrypt(base64Decode, []byte(v))
	if err != nil {
		return nil, datamodels.HttpActivationInvalidateError
	}
	var activationInfo datamodels.Activation
	err = json.Unmarshal(decrypt, &activationInfo)
	if err != nil {
		return nil, datamodels.HttpActivationInvalidateError
	}
	if activationInfo.Sn != id {
		return nil, datamodels.HttpActivationInvalidateError
	}
	if a.isExpired(&activationInfo) {
		return nil, datamodels.HttpActivationExpiredError
	}
	return &activationInfo, datamodels.HttpSuccess
}

func (a *activation) ParseKeystoreFile() (*datamodels.Activation, datamodels.HttpStatusCode) {
	if utils.PathExists(KeyStorePath){
		data, err := ioutil.ReadFile(KeyStorePath)
		if err != nil {
			return nil, datamodels.HttpActivationReadFileError
		}
		return a.ParseKeystoreContent(string(data))
	}
	return nil, datamodels.HttpActivationNotFound
}

func (a *activation) isExpired(activationInfo *datamodels.Activation) bool {
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	return activationInfo.ExpiredAt < currentTime
}




