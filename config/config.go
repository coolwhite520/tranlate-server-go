package config

import (
	"encoding/json"
	"io/ioutil"
	"sync"
	"translate-server/datamodels"
	"translate-server/utils"
)

var instance *ImgConfig
var once sync.Once

type ImgConfig struct {
	secret string
	configName string
	systemConfig *datamodels.SystemConfig
}

func GetInstance() *ImgConfig {
	once.Do(func() {
		instance = &ImgConfig{}
		instance.secret = "ecf274d323fab23667a2ccd7904803c8"
		instance.configName = "./config.dat"
	})
	return instance
}
// TestGenerateConfigFile 自己测试的时候使用
func (i *ImgConfig) TestGenerateConfigFile() error {
	var configList []datamodels.ComponentInfo
	web := datamodels.ComponentInfo{
		FileName:      "nginx-web.tar",
		ImageName:     "nginx-web",
		ContainerName: "nginx-web",
		ContainerTag:  "1.0.1",
		FileMd5:       "",
		InternalPort:  "8080",
		ExposePort:    "8080",
		//DefaultRun:    true,
		DefaultRun: false, // 测试的时候需要联合调试，所以就不启动他
	}
	tika := datamodels.ComponentInfo{
		FileName:      "tika.tar",
		ImageName:     "tika",
		ContainerName: "tika",
		ContainerTag:  "1.0.1",
		FileMd5:       "",
		InternalPort:  "9998",
		ExposePort:    "9998",
		DefaultRun:    false,
	}
	translate := datamodels.ComponentInfo{
		FileName:      "translate.tar",
		ImageName:     "translate",
		ContainerName: "translate",
		ContainerTag:  "1.0.1",
		FileMd5:       "",
		InternalPort:  "5000",
		ExposePort:    "5000",
		DefaultRun:    false,
	}
	tesseract := datamodels.ComponentInfo{
		FileName:      "tesseract.tar",
		ImageName:     "tesseract",
		ContainerName: "tesseract",
		ContainerTag:  "1.0.1",
		FileMd5:       "",
		InternalPort:  "9090",
		ExposePort:    "9090",
		DefaultRun:    false,
	}
	configList = append(configList, web, tika, translate, tesseract)
	var systemConfig datamodels.SystemConfig
	systemConfig.ComponentList = configList
	systemConfig.SystemVersion = "3.2.1"
	return i.generateConfigFile(systemConfig)
}
// generateConfigFile 由我们自己控制
func (i *ImgConfig) generateConfigFile(systemConfig datamodels.SystemConfig) error {
	marshal, err := json.Marshal(systemConfig)
	if err != nil {
		return err
	}
	encrypt, err := utils.AesEncrypt(marshal, []byte(i.secret))
	if err != nil {
		return err
	}
	return ioutil.WriteFile(i.configName, encrypt, 0777)
}

func (i *ImgConfig) ParseConfigFile(reload bool) (*datamodels.SystemConfig, error) {
	if !reload {
		if i.systemConfig != nil {
			return i.systemConfig, nil
		}
	}
	bytes, err := ioutil.ReadFile(i.configName)
	if err != nil {
		return nil, err
	}
	decrypt, err := utils.AesDecrypt(bytes, []byte(i.secret))
	if err != nil {
		return nil, err
	}
	var systemConfig datamodels.SystemConfig
	err = json.Unmarshal(decrypt, &systemConfig)
	if err != nil {
		return nil, err
	}
	i.systemConfig = &systemConfig
	return i.systemConfig, nil
}
