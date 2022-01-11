package services

import (
	"github.com/Unknwon/goconfig"
	"translate-server/utils"
)

func InitConfigIniFile() error {
	if !utils.PathExists("./config.ini") {
		content := `
[IP_TABLE]
type=
`
		cfg, err := goconfig.LoadFromData([]byte(content))
		if err != nil {
			return err
		}
		err = goconfig.SaveConfigFile(cfg, "./config.ini")
		if err != nil {
			return err
		}
	}
	return nil
}
