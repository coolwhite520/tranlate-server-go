package config

import (
	"encoding/json"
	"io/ioutil"
	"sync"
	"translate-server/datamodels"
	"translate-server/utils"
)

var instance *ConfigureLoader
var once sync.Once

type ConfigureLoader struct {
	secret string
	systemConfigFilePath string
	systemConfig *datamodels.SystemConfig
}

func GetInstance() *ConfigureLoader {
	once.Do(func() {
		instance = &ConfigureLoader{}
		instance.secret = "ecf274d323fab23667a2ccd7904803c8"
		instance.systemConfigFilePath = "./config.dat"
	})
	return instance
}
// TestGenerateConfigFile 自己测试的时候使用
func (i *ConfigureLoader) TestGenerateConfigFile() error {
	var configList []datamodels.ComponentInfo
	//web := datamodels.ComponentInfo{
	//	FileName:      "nginx-web.tar",
	//	ImageName:     "nginx-web",
	//	ContainerName: "nginx-web",
	//	ImageVersion:  "v1",
	//	FileMd5:       "",
	//	ExposedPort:   "8080",
	//	HostPort:      "8080",
	//	DefaultRun:    true,
	//}
	tika := datamodels.ComponentInfo{
		FileName:      "tika.tar",
		ImageName:     "tika",
		ContainerName: "tika",
		ImageVersion:  "v1",
		FileMd5:       "",
		ExposedPort:   "9998",
		HostPort:      "9998",
		DefaultRun:    false,
	}
	translate := datamodels.ComponentInfo{
		FileName:      "translate.tar",
		ImageName:     "translate",
		ContainerName: "translate",
		ImageVersion:  "v1",
		FileMd5:       "",
		ExposedPort:   "5000",
		HostPort:      "5000",
		DefaultRun:    false,
	}
	tesseract := datamodels.ComponentInfo{
		FileName:      "tesseract.tar",
		ImageName:     "tesseract",
		ContainerName: "tesseract",
		ImageVersion:  "v1",
		FileMd5:       "",
		ExposedPort:   "9090",
		HostPort:      "9090",
		DefaultRun:    false,
	}
	//configList = append(configList, web, tika, translate, tesseract)
	configList = append(configList, tika, translate, tesseract)
	var systemConfig datamodels.SystemConfig
	systemConfig.ComponentList = configList
	systemConfig.SystemVersion = "3.2.1"
	return i.GenerateSystemConfigFile(systemConfig)
}
// GenerateSystemConfigFile 由我们自己控制
func (i *ConfigureLoader) GenerateSystemConfigFile(systemConfig datamodels.SystemConfig) error {
	marshal, err := json.Marshal(systemConfig)
	if err != nil {
		return err
	}
	encrypt, err := utils.AesEncrypt(marshal, []byte(i.secret))
	if err != nil {
		return err
	}
	return ioutil.WriteFile(i.systemConfigFilePath, encrypt, 0777)
}

// ParseSystemConfigFile 解析系统配置文件
func (i *ConfigureLoader) ParseSystemConfigFile(reload bool) (*datamodels.SystemConfig, error) {
	if !reload {
		if i.systemConfig != nil {
			return i.systemConfig, nil
		}
	}
	bytes, err := ioutil.ReadFile(i.systemConfigFilePath)
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

// GenerateComponentConfigFile 由我们自己控制
func (i *ConfigureLoader) GenerateComponentConfigFile(comp datamodels.ComponentInfo, componentConfigPath string) error {
	marshal, err := json.Marshal(comp)
	if err != nil {
		return err
	}
	encrypt, err := utils.AesEncrypt(marshal, []byte(i.secret))
	if err != nil {
		return err
	}
	return ioutil.WriteFile(componentConfigPath, encrypt, 0777)
}

// ParseComponentConfigFile 解析组件内配置文件
func (i *ConfigureLoader) ParseComponentConfigFile(componentConfigPath string) (*datamodels.ComponentInfo, error) {
	bytes, err := ioutil.ReadFile(componentConfigPath)
	if err != nil {
		return nil, err
	}
	decrypt, err := utils.AesDecrypt(bytes, []byte(i.secret))
	if err != nil {
		return nil, err
	}
	var componentInfo datamodels.ComponentInfo
	err = json.Unmarshal(decrypt, &componentInfo)
	if err != nil {
		return nil, err
	}
	return &componentInfo, nil
}