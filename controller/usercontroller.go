package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"translate-server/models"
	"translate-server/repositories"
)

type UserController struct {
	Ctx iris.Context
}

func (u *UserController) BeforeActivation(a mvc.BeforeActivation)  {
	//a.Handle("GET", "/info", "GetSomeThing")
}


func (u *UserController) PostLogin() mvc.Result {

	return mvc.Response{}
}

func (u *UserController) Get() mvc.Result{
	users, err := repositories.QueryAllUsers()
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code" : -100,
				"error": err.Error(),
			},
		}
	}
	return mvc.Response{
		Object: users,
	}
}

func (u *UserController) Post() mvc.Result {
	var newUser models.User
	err := u.Ctx.ReadJSON(&newUser)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code" : -100,
				"error": err.Error(),
			},
		}
	}
	err = repositories.InsertUser(newUser)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code" : -100,
				"error": err.Error(),
			},
		}
	}
	return mvc.Response{
		Code: 200,
	}
}