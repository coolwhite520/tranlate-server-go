package middleware

import (
	"github.com/kataras/iris/v12"
	"translate-server/structs"
)

// CheckSuperMiddleware 当前的controller为用户管理模块，需要超级用户
func CheckSuperMiddleware(Ctx iris.Context) {
	a:= Ctx.Values().Get("User")
	if user, ok := (a).(structs.User); ok && user.IsSuper {
		Ctx.Next()
		return
	}
	Ctx.JSON(map[string]interface{}{
		"code":  structs.HttpUserForbidden,
		"msg": structs.HttpUserForbidden.String(),
	})
}
