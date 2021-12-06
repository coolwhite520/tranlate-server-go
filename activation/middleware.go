package activation

import (
	"github.com/kataras/iris/v12"
	"time"
	"translate-server/datamodels"
)

func CheckActivationMiddleware(ctx iris.Context)  {
	//log.Info(ctx.Path())
	newActivation := NewActivation()
	id, _ := newActivation.GenerateMachineId()
	activationInfo := datamodels.ActivationInfo{
		UserName:        "panda",
		SupportLangList: []string{"zh", "en", "ur"},
		CreatedDate:     time.Now(),
		ExpiredDate:     time.Date(2099, 1, 1, 1, 1, 1, 0, time.Local),
		MachineId:       id,
	}
	content, state := newActivation.GenerateKeystoreContent(activationInfo)
	if state != Success {

	}
	_, state = newActivation.ParseKeystoreFile()
	if state != Success {
		ctx.JSON(
			map[string]interface{}{
				"code": -100,
				"machineId": id,
				"err": state.String(),
				"keystore": content,
			})
		return
	}
	ctx.Next()
}

