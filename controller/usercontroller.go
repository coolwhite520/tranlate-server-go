package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"translate-server/models"
)

type UserController struct {
	Ctx iris.Context
}

func (u *UserController) Get() mvc.Result{
	users, err := models.QueryAllUsers()
	if err != nil {
		return mvc.Response{
			ContentType: "application/json",
			Err: err,
		}
	}
	return mvc.Response{
		ContentType: "application/json",
		Object: users,
	}
}

func (u *UserController) Post() mvc.Result {
	var newUser models.User
	err := u.Ctx.ReadJSON(&newUser)
	if err != nil {
		return mvc.Response{
			ContentType: "application/json",
			Err: err,
		}
	}
	err = models.InsertUser(newUser)
	if err != nil {
		return mvc.Response{
			ContentType: "application/json",
			Err: err,
		}
	}
	return mvc.Response{
		ContentType: "application/json",
		Code: 200,
	}
}