package middleware

import (
	"github.com/kataras/iris/v12"
	"time"
	"translate-server/constant"
	"translate-server/datamodels"
	"translate-server/structs"
)

func CheckActivationMiddleware(ctx iris.Context) {
	newActivation := datamodels.NewActivationModel()
	sn := newActivation.GetMachineId()
	activationInfo, state := newActivation.ParseKeystoreFile()
	if state != constant.HttpSuccess {
		ctx.JSON(
			map[string]interface{}{
				"code":      state,
				"sn":        sn,
				"msg":       state.String(),
			})
		return
	}

	// 是否被永久失效了
	bannedInfo, _ := newActivation.ParseBannedFile()
	if bannedInfo != nil {
		for _, v := range bannedInfo.Ids {
			if v == activationInfo.CreatedAt {
				ctx.JSON(
					map[string]interface{}{
						"code": constant.HttpActivationInvalidateError,
						"sn":   sn,
						"msg":  constant.HttpActivationInvalidateError.String(),
					})
				return
			}
		}
	}

	// 是否过期了
	expiredInfo, state := newActivation.ParseExpiredFile()
	// 用户失误或故意删除了/usr/bin/${machineID}的文件，我们再替他生成回来
	if state == constant.HttpActivationNotFound {
		expiredInfo = new(structs.KeystoreExpired)
		expiredInfo.Sn = activationInfo.Sn
		expiredInfo.CreatedAt = activationInfo.CreatedAt
		expiredInfo.LeftTimeSpan = activationInfo.UseTimeSpan - (time.Now().Unix() - activationInfo.CreatedAt)
		if expiredInfo.LeftTimeSpan <= 0 {
			ctx.JSON(
				map[string]interface{}{
					"code": constant.HttpActivationExpiredError,
					"sn":   sn,
					"msg":  constant.HttpActivationExpiredError.String(),
				})
			return
		}
		newActivation.GenerateExpiredFile(*expiredInfo)
	} else if state == constant.HttpSuccess {
		if expiredInfo.LeftTimeSpan <= 0 {
			ctx.JSON(
				map[string]interface{}{
					"code": constant.HttpActivationExpiredError,
					"sn":   sn,
					"msg":  constant.HttpActivationExpiredError.String(),
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
