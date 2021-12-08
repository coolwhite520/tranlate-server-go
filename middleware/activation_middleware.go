package middleware

import (
	"github.com/kataras/iris/v12"
	"time"
	"translate-server/datamodels"
	"translate-server/services"
)

func CheckActivationMiddleware(ctx iris.Context) {
	newActivation := services.NewActivationService()
	id, _ := newActivation.GenerateMachineId()
	var langList []datamodels.SupportLang
	langList = append(langList, datamodels.SupportLang{
		EnName: "en",
		CnName: "英语",
	}, datamodels.SupportLang{
		EnName: "cn",
		CnName: "中文",
	})
	activationInfo := datamodels.Activation{
		UserName:        "panda",
		SupportLangList: langList,
		CreatedAt:       time.Now().Format("2006-01-02 15:04:05"),
		ExpiredAt:       time.Date(2099, 1, 1, 1, 1, 1, 1, time.Local).Format("2006-01-02 15:04:05"),
		MachineId:       id,
	}
	content, state := newActivation.GenerateKeystoreContent(activationInfo)
	if state != services.Success {

	}
	_, state = newActivation.ParseKeystoreFile()
	if state != services.Success {
		ctx.JSON(
			map[string]interface{}{
				"code":      -100,
				"machineId": id,
				"msg":       state.String(),
				"keystore":  content,
			})
		return
	}
	ctx.Next()
}
