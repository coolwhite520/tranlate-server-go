package middleware

import (
	"github.com/kataras/iris/v12"
	"translate-server/constant"
	"translate-server/docker"
)

func IsSystemAvailable(ctx iris.Context) {
	ok, err := docker.GetInstance().IsALlRunningStatus()
	if err != nil {
		return
	}
	if !ok  {
		if  docker.GetInstance().GetStatus() == docker.InitializingStatus {
			ctx.JSON(
				map[string]interface{}{
					"code":    constant.HttpDockerInitializing,
					"msg":     constant.HttpDockerInitializing.String(),
					"percent": docker.GetInstance().GetPercent(),
				})
			return
		} else if docker.GetInstance().GetStatus() == docker.RepairingStatus{
			ctx.JSON(
				map[string]interface{}{
					"code":    constant.HttpDockerRepairing,
					"msg":     constant.HttpDockerRepairing.String(),
					"percent": docker.GetInstance().GetPercent(),
				})
			return
		} else if docker.GetInstance().GetStatus() == docker.ErrorStatus {
			ctx.JSON(
				map[string]interface{}{
					"code": constant.HttpDockerServiceException,
					"msg":  constant.HttpDockerServiceException.String(),
				})
			return
		}
	}
	ctx.Next()
}