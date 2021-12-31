package services

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/denisbrodbeck/machineid"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"syscall"
	"translate-server/datamodels"
	"translate-server/utils"
)

type ActivationService interface {
	GetMachineId() string
	GenerateKeystoreContent(activationInfo datamodels.Activation) (string, datamodels.HttpStatusCode)
	GenerateKeystoreFileByContent(string) datamodels.HttpStatusCode
	ParseKeystoreContent(content string) (*datamodels.Activation, datamodels.HttpStatusCode)
	ParseKeystoreFile() (*datamodels.Activation, datamodels.HttpStatusCode)
	IsSupportLang(srcLang, desLang string) (bool, []datamodels.SupportLang)

	GenerateExpiredFile(datamodels.KeystoreExpired) datamodels.HttpStatusCode
	ParseExpiredContent(content string) (*datamodels.KeystoreExpired, datamodels.HttpStatusCode)
	ParseExpiredFile() (*datamodels.KeystoreExpired, datamodels.HttpStatusCode)
}

func NewActivationService() (ActivationService, error) {
	p := new(activation)
	id, err := p.generateMachineId()
	if err != nil {
		return nil, err
	}
	p.currentMachineId = id
	p.expiredFilePath = fmt.Sprintf("./.%s", id)
	return p, nil
}

const KeyStorePath = "./keystore"
const AppID = "@My_TrAnSLaTe_sErVeR"


type activation struct {
	currentMachineId string
	expiredFilePath string
}

// GenerateMachineId 生成机器码
func (a *activation) generateMachineId() (string, error) {
	id, err := machineid.ProtectedID(AppID)
	if err != nil {
		log.Errorln(err)
		return "", err
	}
	return id, err
}

func (a *activation) GetMachineId() string {
	return a.currentMachineId
}

func isIn(target string, strArray []datamodels.SupportLang) bool {
	for _, element := range strArray {
		if target == element.EnName {
			return true
		}
	}
	return false
}

func (a *activation)IsSupportLang(srcLang, desLang string) (bool, []datamodels.SupportLang) {
	file, state := a.ParseKeystoreFile()
	if state != datamodels.HttpSuccess {
		return false, file.SupportLangList
	}
	if !isIn(srcLang, file.SupportLangList) {
		return false, file.SupportLangList
	}
	if !isIn(desLang, file.SupportLangList) {
		return false, file.SupportLangList
	}
	return true, file.SupportLangList
}

func (a activation) GenerateKeystoreContent(activationInfo datamodels.Activation) (string, datamodels.HttpStatusCode) {
	data, err := json.Marshal(activationInfo)
	if err != nil {
		log.Errorln(err)
		return "", datamodels.HttpActivationGenerateError
	}
	v := utils.Md5V(activationInfo.Sn + AppID)
	encrypt, err := utils.AesEncrypt(data, []byte(v))
	if err != nil {
		log.Errorln(err)
		return "", datamodels.HttpActivationAESError
	}
	toString := base64.StdEncoding.EncodeToString(encrypt)
	return toString, datamodels.HttpSuccess
}


func (a *activation) GenerateKeystoreFileByContent(content string) datamodels.HttpStatusCode {
	ioutil.WriteFile(KeyStorePath, []byte(content), 0666)
	return datamodels.HttpSuccess
}

func (a *activation) ParseKeystoreContent(content string) (*datamodels.Activation, datamodels.HttpStatusCode) {
	v := utils.Md5V(a.currentMachineId + AppID)
	base64Decode, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return nil, datamodels.HttpActivationInvalidateError
	}
	decrypt, err := utils.AesDecrypt(base64Decode, []byte(v))
	if err != nil {
		log.Errorln(err)
		return nil, datamodels.HttpActivationInvalidateError
	}
	var activationInfo datamodels.Activation
	err = json.Unmarshal(decrypt, &activationInfo)
	if err != nil {
		log.Errorln(err)
		return nil, datamodels.HttpActivationInvalidateError
	}
	if activationInfo.Sn != a.currentMachineId {
		return nil, datamodels.HttpActivationInvalidateError
	}
	return &activationInfo, datamodels.HttpSuccess
}

func (a *activation) ParseKeystoreFile() (*datamodels.Activation, datamodels.HttpStatusCode) {
	if utils.PathExists(KeyStorePath){
		data, err := ioutil.ReadFile(KeyStorePath)
		if err != nil {
			log.Errorln(err)
			return nil, datamodels.HttpActivationReadFileError
		}
		return a.ParseKeystoreContent(string(data))
	}
	return nil, datamodels.HttpActivationNotFound
}



func (a *activation) GenerateExpiredFile(keystoreExpired datamodels.KeystoreExpired) datamodels.HttpStatusCode {
	data, err := json.Marshal(keystoreExpired)
	if err != nil {
		return datamodels.HttpActivationGenerateError
	}
	v := utils.Md5V(keystoreExpired.Sn + AppID)
	encrypt, err := utils.AesEncrypt(data, []byte(v))
	if err != nil {
		return datamodels.HttpActivationAESError
	}
	content := base64.StdEncoding.EncodeToString(encrypt)
	f, err := os.Create(a.expiredFilePath)
	if err != nil {
		log.Errorln(err)
		return datamodels.HttpActivationGenerateError
	}
	defer f.Close()
	// 非阻塞模式下，加共享锁
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_SH | syscall.LOCK_NB); err != nil {
		log.Errorln(err)
		return datamodels.HttpActivationGenerateError
	}
	_, err = f.WriteString(content)
	if err != nil {
		log.Errorln(err)
		return datamodels.HttpActivationGenerateError
	}
	// 解锁
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_UN); err != nil {
		log.Errorln(err)
		return datamodels.HttpActivationGenerateError
	}

	return datamodels.HttpSuccess
}

func (a *activation) ParseExpiredContent(content string) (*datamodels.KeystoreExpired, datamodels.HttpStatusCode) {
	v := utils.Md5V(a.currentMachineId + AppID)
	base64Decode, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		log.Errorln(err)
		return nil, datamodels.HttpActivationInvalidateError
	}
	decrypt, err := utils.AesDecrypt(base64Decode, []byte(v))
	if err != nil {
		log.Errorln(err)
		return nil, datamodels.HttpActivationInvalidateError
	}
	var expired datamodels.KeystoreExpired
	err = json.Unmarshal(decrypt, &expired)
	if err != nil {
		log.Errorln(err)
		return nil, datamodels.HttpActivationInvalidateError
	}
	if expired.Sn != a.currentMachineId {
		return nil, datamodels.HttpActivationInvalidateError
	}
	return &expired, datamodels.HttpSuccess
}

func (a *activation) ParseExpiredFile() (*datamodels.KeystoreExpired, datamodels.HttpStatusCode) {
	if utils.PathExists(a.expiredFilePath){
		f, err := os.Open(a.expiredFilePath)
		if err != nil {
			log.Errorln(err)
			return nil, datamodels.HttpActivationReadFileError
		}
		defer f.Close()
		// 非阻塞模式下，加共享锁
		if err := syscall.Flock(int(f.Fd()), syscall.LOCK_SH | syscall.LOCK_NB); err != nil {
			log.Errorln(err)
			return nil, datamodels.HttpActivationReadFileError
		}
		all, err := ioutil.ReadAll(f)
		if err != nil {
			log.Errorln(err)
			return nil, datamodels.HttpActivationReadFileError
		}
		// 解锁
		if err := syscall.Flock(int(f.Fd()), syscall.LOCK_UN); err != nil {
			log.Errorln(err)
			return nil, datamodels.HttpActivationReadFileError
		}
		return a.ParseExpiredContent(string(all))
	}
	return nil, datamodels.HttpActivationNotFound
}
