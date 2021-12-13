package middleware

import (
	"github.com/kataras/iris/v12"
	"translate-server/docker"
)

func IsSystemAvailable(ctx iris.Context) {
	ok, err := docker.GetInstance().IsALlRunningStatus()
	if err != nil {
		return
	}
	if !ok  {
		if  docker.GetInstance().GetStatus() == docker.Initializing {
			ctx.JSON(
				map[string]interface{}{
					"code": -100,
					"msg": "当前系统正在进行初始化,大约需要几分钟，请稍后...",
					"percent": 19,
				})
			return
		} else {
			ctx.JSON(
				map[string]interface{}{
					"code": -100,
					"msg": "当前系统服务异常，请联系管理员...",
				})
			return
		}
	}
	ctx.Next()
}