package datamodels

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/denisbrodbeck/machineid"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"runtime"
	"sync"
	"syscall"
	"translate-server/structs"
	"translate-server/utils"
)

const KeyStorePath = "./keystore"
const AppID = "@My_TrAnSLaTe_sErVeR"

var instance *activation
var once sync.Once


func NewActivationModel() *activation {
	once.Do(func() {
		instance = new(activation)
		id, _ := instance.generateMachineId()
		instance.currentMachineId = id
		systemType := runtime.GOOS
		if  systemType == "linux"{
			instance.expiredFilePath = fmt.Sprintf("/usr/bin/.%s", id)
		} else {
			instance.expiredFilePath = fmt.Sprintf("./.%s", id)
		}
	})
	return instance
}

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

func (a *activation) isIn(target string, strArray []structs.SupportLang) bool {
	for _, element := range strArray {
		if target == element.EnName {
			return true
		}
	}
	return false
}

func (a *activation) IsSupportLang(srcLang, desLang string) (bool, []structs.SupportLang) {
	file, state := a.ParseKeystoreFile()
	if state != structs.HttpSuccess {
		return false, file.SupportLangList
	}
	if !a.isIn(srcLang, file.SupportLangList) {
		return false, file.SupportLangList
	}
	if !a.isIn(desLang, file.SupportLangList) {
		return false, file.SupportLangList
	}
	return true, file.SupportLangList
}

func (a activation) GenerateKeystoreContent(activationInfo structs.Activation) (string, structs.HttpStatusCode) {
	data, err := json.Marshal(activationInfo)
	if err != nil {
		log.Errorln(err)
		return "", structs.HttpActivationGenerateError
	}
	v := utils.Md5V(activationInfo.Sn + AppID)
	encrypt, err := utils.AesEncrypt(data, []byte(v))
	if err != nil {
		log.Errorln(err)
		return "", structs.HttpActivationAESError
	}
	toString := base64.StdEncoding.EncodeToString(encrypt)
	return toString, structs.HttpSuccess
}


func (a *activation) GenerateKeystoreFileByContent(content string) structs.HttpStatusCode {
	ioutil.WriteFile(KeyStorePath, []byte(content), 0666)
	return structs.HttpSuccess
}

func (a *activation) ParseKeystoreContent(content string) (*structs.Activation, structs.HttpStatusCode) {
	v := utils.Md5V(a.currentMachineId + AppID)
	base64Decode, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return nil, structs.HttpActivationInvalidateError
	}
	decrypt, err := utils.AesDecrypt(base64Decode, []byte(v))
	if err != nil {
		log.Errorln(err)
		return nil, structs.HttpActivationInvalidateError
	}
	var activationInfo structs.Activation
	err = json.Unmarshal(decrypt, &activationInfo)
	if err != nil {
		log.Errorln(err)
		return nil, structs.HttpActivationInvalidateError
	}
	if activationInfo.Sn != a.currentMachineId {
		return nil, structs.HttpActivationInvalidateError
	}
	return &activationInfo, structs.HttpSuccess
}

func (a *activation) ParseKeystoreFile() (*structs.Activation, structs.HttpStatusCode) {
	if utils.PathExists(KeyStorePath){
		data, err := ioutil.ReadFile(KeyStorePath)
		if err != nil {
			log.Errorln(err)
			return nil, structs.HttpActivationReadFileError
		}
		return a.ParseKeystoreContent(string(data))
	}
	return nil, structs.HttpActivationNotFound
}

func (a *activation) GenerateExpiredFile(keystoreExpired structs.KeystoreExpired) structs.HttpStatusCode {
	data, err := json.Marshal(keystoreExpired)
	if err != nil {
		return structs.HttpActivationGenerateError
	}
	v := utils.Md5V(keystoreExpired.Sn + AppID)
	encrypt, err := utils.AesEncrypt(data, []byte(v))
	if err != nil {
		return structs.HttpActivationAESError
	}
	content := base64.StdEncoding.EncodeToString(encrypt)
	f, err := os.Create(a.expiredFilePath)
	if err != nil {
		log.Errorln(err)
		return structs.HttpActivationGenerateError
	}
	defer f.Close()
	// 非阻塞模式下，加共享锁
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_SH | syscall.LOCK_NB); err != nil {
		log.Errorln(err)
		return structs.HttpActivationGenerateError
	}
	_, err = f.WriteString(content)
	if err != nil {
		log.Errorln(err)
		return structs.HttpActivationGenerateError
	}
	// 解锁
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_UN); err != nil {
		log.Errorln(err)
		return structs.HttpActivationGenerateError
	}

	return structs.HttpSuccess
}

func (a *activation) ParseExpiredContent(content string) (*structs.KeystoreExpired, structs.HttpStatusCode) {
	v := utils.Md5V(a.currentMachineId + AppID)
	base64Decode, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		log.Errorln(err)
		return nil, structs.HttpActivationInvalidateError
	}
	decrypt, err := utils.AesDecrypt(base64Decode, []byte(v))
	if err != nil {
		log.Errorln(err)
		return nil, structs.HttpActivationInvalidateError
	}
	var expired structs.KeystoreExpired
	err = json.Unmarshal(decrypt, &expired)
	if err != nil {
		log.Errorln(err)
		return nil, structs.HttpActivationInvalidateError
	}
	if expired.Sn != a.currentMachineId {
		return nil, structs.HttpActivationInvalidateError
	}
	return &expired, structs.HttpSuccess
}

func (a *activation) ParseExpiredFile() (*structs.KeystoreExpired, structs.HttpStatusCode) {
	if utils.PathExists(a.expiredFilePath){
		f, err := os.Open(a.expiredFilePath)
		if err != nil {
			log.Errorln(err)
			return nil, structs.HttpActivationReadFileError
		}
		defer f.Close()
		// 非阻塞模式下，加共享锁
		if err := syscall.Flock(int(f.Fd()), syscall.LOCK_SH | syscall.LOCK_NB); err != nil {
			log.Errorln(err)
			return nil, structs.HttpActivationReadFileError
		}
		all, err := ioutil.ReadAll(f)
		if err != nil {
			log.Errorln(err)
			return nil, structs.HttpActivationReadFileError
		}
		// 解锁
		if err := syscall.Flock(int(f.Fd()), syscall.LOCK_UN); err != nil {
			log.Errorln(err)
			return nil, structs.HttpActivationReadFileError
		}
		return a.ParseExpiredContent(string(all))
	}
	return nil, structs.HttpActivationNotFound
}
