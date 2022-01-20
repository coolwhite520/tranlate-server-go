package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"translate-server/middleware"
	"translate-server/services"
)

type UserController struct {
	Ctx         iris.Context
	UserService services.UserService
}

func (u *UserController) BeforeActivation(b mvc.BeforeActivation) {
	//b.Router().Use(middleware.CheckActivationMiddleware, middleware.IsSystemAvailable)
	// 只有登录以后，才可以进行密码修改
	b.Handle("POST", "/password", "PostPassword", middleware.CheckLoginMiddleware)
	b.Handle("POST", "/logoff", "PostLogoff", middleware.CheckLoginMiddleware)
	b.Handle("GET", "/favor", "GetQueryUserFavor", middleware.CheckLoginMiddleware)
	b.Handle("POST", "/favor", "PostAddUserFavor", middleware.CheckLoginMiddleware)
}

func (u *UserController) PostAddUserFavor() mvc.Result {
	return u.UserService.PostAddUserFavor(u.Ctx)
}

func (u *UserController) GetQueryUserFavor() mvc.Result {
	return u.UserService.GetQueryUserFavor(u.Ctx)
}

// PostPassword /api/user/password
func (u *UserController) PostPassword() mvc.Result {
	return u.UserService.PostPassword(u.Ctx)
}

func (u *UserController) PostLogoff() mvc.Result {
	return u.UserService.PostLogoff(u.Ctx)
}

// PostLogin /api/user/login
func (u *UserController) PostLogin() mvc.Result {
	return u.UserService.PostLogin(u.Ctx)
}

