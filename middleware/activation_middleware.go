package middleware

import (
	"github.com/kataras/iris/v12"
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
	// 是否被永久失效 或 已经过期
	bannedList, _ := newActivation.ParseBannedFile()
	if bannedList != nil {
		for _, v := range bannedList {
			if v.Id == activationInfo.CreatedAt && v.State == structs.ProofStateForceBanned {
				ctx.JSON(
					map[string]interface{}{
						"code": constant.HttpActivationInvalidateError,
						"sn":   sn,
						"msg":  constant.HttpActivationInvalidateError.String(),
					})
				return
			}
			if v.Id == activationInfo.CreatedAt && v.State == structs.ProofStateExpired {
				ctx.JSON(
					map[string]interface{}{
						"code": constant.HttpActivationExpiredError,
						"sn":   sn,
						"msg":  constant.HttpActivationExpiredError.String(),
					})
				return
			}
		}
	}

	ctx.Next()
}
