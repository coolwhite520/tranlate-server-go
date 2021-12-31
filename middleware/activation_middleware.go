package middleware

import (
	"github.com/kataras/iris/v12"
	log "github.com/sirupsen/logrus"
	"time"
	"translate-server/datamodels"
	"translate-server/services"
)

func CheckActivationMiddleware(ctx iris.Context) {
	newActivation, err := services.NewActivationService()
	if err != nil {
		log.Errorln(err)
		panic(err)
	}
	sn := newActivation.GetMachineId()
	activationInfo, state := newActivation.ParseKeystoreFile()
	if state != datamodels.HttpSuccess {
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
	if state == datamodels.HttpActivationNotFound {
		expiredInfo = new(datamodels.KeystoreExpired)
		expiredInfo.Sn = activationInfo.Sn
		expiredInfo.CreatedAt = activationInfo.CreatedAt
		expiredInfo.LeftTimeSpan = activationInfo.UseTimeSpan - (time.Now().Unix() - activationInfo.CreatedAt)
		if expiredInfo.LeftTimeSpan <= 0 {
			ctx.JSON(
				map[string]interface{}{
					"code":      datamodels.HttpActivationExpiredError,
					"sn":        sn,
					"msg":       datamodels.HttpActivationExpiredError.String(),
				})
			return
		}
		newActivation.GenerateExpiredFile(*expiredInfo)
	} else if state == datamodels.HttpSuccess {
		if expiredInfo.LeftTimeSpan <= 0 {
			ctx.JSON(
				map[string]interface{}{
					"code":      datamodels.HttpActivationExpiredError,
					"sn":        sn,
					"msg":       datamodels.HttpActivationExpiredError.String(),
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
