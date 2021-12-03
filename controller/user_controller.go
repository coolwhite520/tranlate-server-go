package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"translate-server/jwt"
	"translate-server/services"
)

type UserController struct {
	Ctx         iris.Context
	UserService services.UserService
}


func (u *UserController) BeforeActivation(a mvc.BeforeActivation) {
	//a.Handle("GET", "/info", "GetSomeThing")
}

func (u *UserController) GetLogin() mvc.Result {
	return mvc.View{
		Name: "user/login.html",
		Data: iris.Map{"Title": "User Login"},
	}
}

func (u *UserController) PostLogin() mvc.Result {
	username := u.Ctx.FormValue("username")
	password := u.Ctx.FormValue("password")
	user, b := u.UserService.CheckUser(username, password)
	if b {
		jwt.GenerateToken(u.Ctx, user)
		return mvc.Response{
			Object: map[string]interface{}{
				"code": 200,
				"msg":  "用户登录成功",
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": -100,
			"msg":  "用户名密码错误",
		},
	}
}
