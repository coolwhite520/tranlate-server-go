package activation

import (
	"github.com/kataras/iris/v12"
	"time"
	"translate-server/datamodels"
)

func CheckActivationMiddleware(ctx iris.Context) {
	newActivation := NewActivation()
	id, _ := newActivation.GenerateMachineId()
	activationInfo := datamodels.ActivationInfo{
		UserName:        "panda",
		SupportLangList: []string{"zh", "en", "ur"},
		CreatedAt:       time.Now().Format("2006-01-02 15:04:05"),
		ExpiredAt:       time.Date(2099, 1, 1, 1, 1, 1, 1, time.Local).Format("2006-01-02 15:04:05"),
		MachineId:       id,
	}
	content, state := newActivation.GenerateKeystoreContent(activationInfo)
	if state != Success {

	}
	_, state = newActivation.ParseKeystoreFile()
	if state != Success {
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
