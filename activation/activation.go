package activation

import (
	"github.com/denisbrodbeck/machineid"
	"io/ioutil"
	"os"
	"translate-server/datamodels"
)

type Activation interface {
	GenerateMachineId() (string, State)
	VerifyAuthenticationFile() (*datamodels.ActivationInfo, State)
}

func NewActivation() Activation {
	return &activation{}
}

const KeyStorePath = "./keystore"

type State int

const (
	NotFound State = iota // value --> 0
	ReadFileError
	ParseError
	ExpiredError
	GenerateError
	Success
)

type activation struct {
}

// GenerateMachineId 生成机器码
func (a *activation) GenerateMachineId() (string, State) {
	id, err := machineid.ProtectedID("@My_TrAnSLaTe_sErVeR")
	if err != nil {
		return "", GenerateError
	}
	return id, Success
}

func (a *activation) VerifyAuthenticationFile() (*datamodels.ActivationInfo, State) {
	_, err := a.pathExists(KeyStorePath)
	if err != nil {
		return nil, NotFound
	}
	return a.parseKeyStoreFile()
}

func (a *activation) parseKeyStoreFile() (*datamodels.ActivationInfo, State) {
	data, err := ioutil.ReadFile(KeyStorePath)
	if err != nil {
		return nil, ReadFileError
	}
	content := string(data)

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

