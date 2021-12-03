package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"translate-server/datamodels"
	"translate-server/jwt"
	"translate-server/services"
)


type UsersController struct {
	Ctx         iris.Context
	UserService services.UserService
}


func (u *UsersController) BeforeActivation(a mvc.BeforeActivation) {
	//a.Handle("GET", "/info", "GetSomeThing")
	a.Router().Use(jwt.ParseToken)
}


func (u *UsersController) Get() mvc.Result {
	users, err := u.UserService.QueryAllUsers()
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code":  -100,
				"error": err.Error(),
			},
		}
	}
	return mvc.Response{
		Object: users,
	}
}

func (u *UsersController) Post() mvc.Result {
	var newUser datamodels.User
	err := u.Ctx.ReadJSON(&newUser)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code":  -100,
				"error": err.Error(),
			},
		}
	}
	err = u.UserService.InsertUser(newUser)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code":  -100,
				"error": err.Error(),
			},
		}
	}
	return mvc.Response{
		Code: 200,
	}
}
