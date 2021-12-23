package middleware

import (
	"github.com/kataras/iris/v12"
	"translate-server/datamodels"
	"translate-server/services"
)

func CheckActivationMiddleware(ctx iris.Context) {
	newActivation := services.NewActivationService()
	sn, _ := newActivation.GenerateMachineId()
	_, state := newActivation.ParseKeystoreFile()
	if state != datamodels.HttpSuccess {
		ctx.JSON(
			map[string]interface{}{
				"code":      state,
				"sn":        sn,
				"msg":       state.String(),
				//"keystore":  content,
			})
		return
	}
	ctx.Next()
}
