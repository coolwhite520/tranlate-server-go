package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	log "github.com/sirupsen/logrus"
	"translate-server/activation"
	"translate-server/datamodels"
	"translate-server/jwt"
	"translate-server/services"
)


type UsersController struct {
	Ctx         iris.Context
	UserService services.UserService
}


func (u *UsersController) BeforeActivation(b mvc.BeforeActivation) {
	b.Router().Use(activation.CheckActivationMiddleware)
	b.Router().Use(jwt.CheckTokenMiddleware)
}

func (u *UsersController) Get() mvc.Result {
	a:= u.Ctx.Values().Get("User")
	log.Info(a)
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
