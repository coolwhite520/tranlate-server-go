package middleware

import (
	"github.com/kataras/iris/v12"
	"time"
	"translate-server/datamodels"
	"translate-server/structs"
)

func CheckActivationMiddleware(ctx iris.Context) {
	newActivation := datamodels.NewActivationModel()
	sn := newActivation.GetMachineId()
	activationInfo, state := newActivation.ParseKeystoreFile()
	if state != structs.HttpSuccess {
		ctx.JSON(
			map[string]interface{}{
				"code":      state,
				"sn":        sn,
				"msg":       state.String(),
			})
		return
	}
	expiredInfo, state := newActivation.ParseExpiredFile()
	// 用户失误或故意删除了/usr/bin/${machineID}的文件，我们再替他生成回来
	if state == structs.HttpActivationNotFound {
		expiredInfo = new(structs.KeystoreExpired)
		expiredInfo.Sn = activationInfo.Sn
		expiredInfo.CreatedAt = activationInfo.CreatedAt
		expiredInfo.LeftTimeSpan = activationInfo.UseTimeSpan - (time.Now().Unix() - activationInfo.CreatedAt)
		if expiredInfo.LeftTimeSpan <= 0 {
			ctx.JSON(
				map[string]interface{}{
					"code":      structs.HttpActivationExpiredError,
					"sn":        sn,
					"msg":       structs.HttpActivationExpiredError.String(),
				})
			return
		}
		newActivation.GenerateExpiredFile(*expiredInfo)
	} else if state == structs.HttpSuccess {
		if expiredInfo.LeftTimeSpan <= 0 {
			ctx.JSON(
				map[string]interface{}{
					"code":      structs.HttpActivationExpiredError,
					"sn":        sn,
					"msg":       structs.HttpActivationExpiredError.String(),
				})
			return
		}
	} else {
		ctx.JSON(
			map[string]interface{}{
				"code":      state,
				"sn":        sn,
				"msg":       state.String(),
			})
		return
	}
	ctx.Next()
}
