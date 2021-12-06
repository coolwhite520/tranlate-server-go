package activation

import (
	"encoding/base64"
	"encoding/json"
	"github.com/denisbrodbeck/machineid"
	"io/ioutil"
	"os"
	"time"
	"translate-server/datamodels"
)

type Activation interface {
	GenerateMachineId() (string, State)
	GenerateKeystoreContent(activationInfo datamodels.ActivationInfo) (string, State)
	GenerateKeystoreFile(datamodels.ActivationInfo) State
	GenerateKeystoreFileByContent(string) State
	ParseKeystoreContent(content string) (*datamodels.ActivationInfo, State)
	ParseKeystoreFile() (*datamodels.ActivationInfo, State)
}

func NewActivation() Activation {
	return &activation{}
}

const KeyStorePath = "./keystore"
const AppID = "@My_TrAnSLaTe_sErVeR"

type State int

const (
	NotFound State = iota // value --> 0
	ReadFileError
	ParseError
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
	case ParseError: return "解析文件失败"
	case ExpiredError: return "授权已经过期"
	case GenerateError: return "生成错误"
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

func (a activation) GenerateKeystoreContent(activationInfo datamodels.ActivationInfo) (string, State) {
	data, err := json.Marshal(activationInfo)
	if err != nil {
		return "", GenerateError
	}
	v := md5V(activationInfo.MachineId + AppID)
	encrypt, err := AesEncrypt(data, []byte(v))
	if err != nil {
		return "",AESError
	}
	toString := base64.StdEncoding.EncodeToString(encrypt)
	return toString, Success
}


func (a *activation) GenerateKeystoreFile(activationInfo datamodels.ActivationInfo) State {
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

func (a *activation) ParseKeystoreContent(content string) (*datamodels.ActivationInfo, State) {
	id, state := a.GenerateMachineId()
	if state != Success {
		return nil, GenerateError
	}
	v := md5V(id + AppID)
	base64Decode, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return nil, ParseError
	}
	decrypt, err := AesDecrypt(base64Decode, []byte(v))
	if err != nil {
		return nil, ParseError
	}
	var activationInfo datamodels.ActivationInfo
	err = json.Unmarshal(decrypt, &activationInfo)
	if err != nil {
		return nil, ParseError
	}
	if activationInfo.MachineId != id {
		return nil, InvalidateError
	}
	if a.isExpired(&activationInfo) {
		return nil, ExpiredError
	}
	return &activationInfo, Success
}

func (a *activation) ParseKeystoreFile() (*datamodels.ActivationInfo, State) {
	if ok, err := a.pathExists(KeyStorePath); ok && err == nil{
		data, err := ioutil.ReadFile(KeyStorePath)
		if err != nil {
			return nil, ReadFileError
		}
		return a.ParseKeystoreContent(string(data))
	}
	return nil, NotFound
}

func (a *activation) isExpired(activationInfo *datamodels.ActivationInfo) bool {
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	return activationInfo.ExpiredAt < currentTime
}


// PathExists 判断文件是否存在
func (a *activation) pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

