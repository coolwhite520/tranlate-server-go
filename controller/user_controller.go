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
	b.Handle("POST", "/password", "ModifyPassword", middleware.CheckLoginMiddleware)
	b.Handle("POST", "/logoff", "Logoff", middleware.CheckLoginMiddleware)
	b.Handle("GET", "/favor", "QueryUserFavor", middleware.CheckLoginMiddleware)
	b.Handle("POST", "/favor", "AddUserFavor", middleware.CheckLoginMiddleware)
}

func (u *UserController) AddUserFavor() mvc.Result {
	return u.UserService.AddUserFavor(u.Ctx)
}

func (u *UserController) QueryUserFavor() mvc.Result {
	return u.UserService.QueryUserFavor(u.Ctx)
}

// ModifyPassword /api/user/password
func (u *UserController) ModifyPassword() mvc.Result {
	return u.UserService.ModifyPassword(u.Ctx)
}

func (u *UserController) Logoff() mvc.Result {
	return u.UserService.Logoff(u.Ctx)
}

// Login /api/user/login
func (u *UserController) Login() mvc.Result {
	return u.UserService.Login(u.Ctx)
}

