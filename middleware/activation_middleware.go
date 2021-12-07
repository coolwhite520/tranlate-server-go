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
	activationInfo := datamodels.Activation{
		UserName:          "panda",
		SupportLangList:   []string{"zh", "en", "fr"},
		SupportLangListCn: []string{"中文", "英文", "法文"},
		CreatedAt:         time.Now().Format("2006-01-02 15:04:05"),
		ExpiredAt:         time.Date(2099, 1, 1, 1, 1, 1, 1, time.Local).Format("2006-01-02 15:04:05"),
		MachineId:         id,
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

func isIn(target string, strArray []string) bool {
	for _, element := range strArray {
		if target == element {
			return true
		}
	}
	return false
}
