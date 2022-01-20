package middleware

import (
	"github.com/kataras/iris/v12"
	"translate-server/docker"
	"translate-server/structs"
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
					"code": structs.HttpDockerInitializing,
					"msg": structs.HttpDockerInitializing.String(),
					"percent": docker.GetInstance().GetPercent(),
				})
			return
		} else if docker.GetInstance().GetStatus() == docker.RepairingStatus{
			ctx.JSON(
				map[string]interface{}{
					"code": structs.HttpDockerRepairing,
					"msg": structs.HttpDockerRepairing.String(),
					"percent": docker.GetInstance().GetPercent(),
				})
			return
		} else {
			ctx.JSON(
				map[string]interface{}{
					"code": structs.HttpDockerServiceException,
					"msg": structs.HttpDockerServiceException.String(),
				})
			return
		}
	}
	ctx.Next()
}