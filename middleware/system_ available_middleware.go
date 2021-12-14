package middleware

import (
	"github.com/kataras/iris/v12"
	"translate-server/datamodels"
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
					"code": datamodels.HttpDockerInitializing,
					"msg": datamodels.HttpDockerInitializing.String(),
					"percent": docker.GetInstance().GetPercent(),
				})
			return
		} else if docker.GetInstance().GetStatus() == docker.RepairingStatus{
			ctx.JSON(
				map[string]interface{}{
					"code": datamodels.HttpDockerRepairing,
					"msg": datamodels.HttpDockerRepairing.String(),
					"percent": docker.GetInstance().GetPercent(),
				})
			return
		} else {
			ctx.JSON(
				map[string]interface{}{
					"code": datamodels.HttpDockerServiceException,
					"msg": datamodels.HttpDockerServiceException.String(),
				})
			return
		}
	}
	ctx.Next()
}