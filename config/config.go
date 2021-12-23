package config

import (
	"encoding/json"
	"fmt"
	"github.com/Unknwon/goconfig"
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
	compList datamodels.ComponentList
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
	//	FileName:      "web.tar",
	//	ImageName:     "web",
	//	ImageVersion:  "3.2.12",
	//	ExposedPort:   "8080",
	//	HostPort:      "8080",
	//	DefaultRun:    true,
	//}
	tika := datamodels.ComponentInfo{
		FileName:      "tk.tar",
		ImageName:     "tk",
		ImageVersion:  "1.5.2",
		ExposedPort:   "9998",
		HostPort:      "9998",
		DefaultRun:    false,
	}
	//core := datamodels.ComponentInfo{
	//	FileName:      "core.tar",
	//	ImageName:     "core",
	//	ImageVersion:  "4.2.3",
	//	ExposedPort:   "5000",
	//	HostPort:      "5000",
	//	DefaultRun:    false,
	//}
	//ocr := datamodels.ComponentInfo{
	//	FileName:      "ocr.tar",
	//	ImageName:     "ocr",
	//	ImageVersion:  "1.8.5",
	//	ExposedPort:   "9090",
	//	HostPort:      "9090",
	//	DefaultRun:    false,
	//}
	//configList = append(configList, web, tika, translate, tesseract)
	configList = append(configList, tika)

	for _, v:= range configList{
		filename := fmt.Sprintf("./%s.dat", v.ImageName)
		i.GenerateComponentDatFile(v, filename)
	}
	return nil
}
func (i ConfigureLoader) GetCompVersions(compName string) []string {
	compPath := fmt.Sprintf("./components/%s", compName)
	fs, _ := ioutil.ReadDir(compPath)
	var comps []string
	for _, v := range fs {
		// 遍历得到文件名
		if v.IsDir() {
			//查看是否存在.dat 和 .tar 文件
			datFile := fmt.Sprintf("%s/%s/%s.dat", compPath, v.Name(), compName)
			tarFile := fmt.Sprintf("%s/%s/%s.tar", compPath, v.Name(), compName)
			if utils.PathExists(datFile) && utils.PathExists(tarFile) {
				comps = append(comps, v.Name())
			}
		}
	}
	return comps
}
// GetComponentList 获取当前系统的组件信息
func (i *ConfigureLoader) GetComponentList(reload bool) (datamodels.ComponentList, error) {
	if !reload {
		if i.compList != nil {
			return i.compList, nil
		}
	}
	m, err := i.parseSystemIniFile()
	if err != nil {
		return nil, err
	}
	var compListTemp datamodels.ComponentList
	for k, v := range m {
		datFilePath := fmt.Sprintf("./components/%s/%s/%s.dat", k, v, k)
		comp, err := i.ParseComponentDatFile(datFilePath)
		if err != nil {
			return nil, err
		}
		compListTemp = append(compListTemp, *comp)
	}
	i.compList = compListTemp
	return i.compList, nil
}
// parseSystemIniFile 解析ini文件
func (i *ConfigureLoader) parseSystemIniFile() (map[string]string, error) {
	cfg, err := goconfig.LoadConfigFile("./versions.ini")
	if err != nil {
		return nil, err
	}
	return cfg.GetSection("components")
}

func (i *ConfigureLoader) SetSectionKeyValue(sectionName, key, value string) error  {
	cfg, err := goconfig.LoadConfigFile("./versions.ini")
	if err != nil {
		return  err
	}
	cfg.SetValue(sectionName, key, value)
	return nil
}

// GetSystemVer 解析ini文件
func (i *ConfigureLoader) GetSystemVer() (string, error) {
	cfg, err := goconfig.LoadConfigFile("./versions.ini")
	if err != nil {
		return "", err
	}
	return cfg.GetValue("system", "version")
}

// GenerateComponentDatFile 由我们自己控制
func (i *ConfigureLoader) GenerateComponentDatFile(comp datamodels.ComponentInfo, componentConfigPath string) error {
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

// ParseComponentDatFile 解析组件内配置文件
func (i *ConfigureLoader) ParseComponentDatFile(componentConfigPath string) (*datamodels.ComponentInfo, error) {
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