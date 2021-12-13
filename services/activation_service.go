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
	GenerateMachineId() (string, State)
	GenerateKeystoreContent(activationInfo datamodels.Activation) (string, State)
	GenerateKeystoreFile(datamodels.Activation) State
	GenerateKeystoreFileByContent(string) State
	ParseKeystoreContent(content string) (*datamodels.Activation, State)
	ParseKeystoreFile() (*datamodels.Activation, State)
}

func NewActivationService() ActivationService {
	return &activation{}
}

const KeyStorePath = "./keystore"
const AppID = "@My_TrAnSLaTe_sErVeR"

type State int

const (
	NotFound State = iota // value --> 0
	ReadFileError
	ExpiredError
	GenerateError
	AESError
	InvalidateError
	Success
)

func (s State) String() string {
	switch s {
	case NotFound: return "未找到激活文件"
	case ReadFileError: return "读取文件失败"
	case ExpiredError: return "授权已经过期"
	case GenerateError: return "生成机器码错误，请联系管理员"
	case AESError: return "Aes加解密错误"
	case InvalidateError: return "无效的激活文件"
	default:         return "成功"
	}
}

type activation struct {
}

// GenerateMachineId 生成机器码
func (a *activation) GenerateMachineId() (string, State) {
	id, err := machineid.ProtectedID(AppID)
	if err != nil {
		return "", GenerateError
	}
	return id, Success
}

func (a activation) GenerateKeystoreContent(activationInfo datamodels.Activation) (string, State) {
	data, err := json.Marshal(activationInfo)
	if err != nil {
		return "", GenerateError
	}
	v := utils.Md5V(activationInfo.MachineId + AppID)
	encrypt, err := utils.AesEncrypt(data, []byte(v))
	if err != nil {
		return "", AESError
	}
	toString := base64.StdEncoding.EncodeToString(encrypt)
	return toString, Success
}


func (a *activation) GenerateKeystoreFile(activationInfo datamodels.Activation) State {
	content, state := a.GenerateKeystoreContent(activationInfo)
	if state != Success {
		return state
	}
	ioutil.WriteFile(KeyStorePath, []byte(content), 0777)
	return Success
}

func (a *activation) GenerateKeystoreFileByContent(content string) State {
	ioutil.WriteFile(KeyStorePath, []byte(content), 0777)
	return Success
}

func (a *activation) ParseKeystoreContent(content string) (*datamodels.Activation, State) {
	id, state := a.GenerateMachineId()
	if state != Success {
		return nil, GenerateError
	}
	v := utils.Md5V(id + AppID)
	base64Decode, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return nil, InvalidateError
	}
	decrypt, err := utils.AesDecrypt(base64Decode, []byte(v))
	if err != nil {
		return nil, InvalidateError
	}
	var activationInfo datamodels.Activation
	err = json.Unmarshal(decrypt, &activationInfo)
	if err != nil {
		return nil, InvalidateError
	}
	if activationInfo.MachineId != id {
		return nil, InvalidateError
	}
	if a.isExpired(&activationInfo) {
		return nil, ExpiredError
	}
	return &activationInfo, Success
}

func (a *activation) ParseKeystoreFile() (*datamodels.Activation, State) {
	if utils.PathExists(KeyStorePath){
		data, err := ioutil.ReadFile(KeyStorePath)
		if err != nil {
			return nil, ReadFileError
		}
		return a.ParseKeystoreContent(string(data))
	}
	return nil, NotFound
}

func (a *activation) isExpired(activationInfo *datamodels.Activation) bool {
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	return activationInfo.ExpiredAt < currentTime
}



