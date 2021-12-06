package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
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
	b.Router().Use(jwt.CheckLoginMiddleware)
	b.Router().Use(CheckSuperMiddleware)
	b.Handle("DELETE","/{id: int64}", "DeleteById")
}

// CheckSuperMiddleware 当前的controller为用户管理模块，需要超级用户
func CheckSuperMiddleware(Ctx iris.Context) {
	a:= Ctx.Values().Get("User")
	if user, ok := (a).(datamodels.User); ok && user.IsSuper {
		Ctx.Next()
		return
	}
	Ctx.JSON(map[string]interface{}{
		"code":  -100,
		"msg": "权限不足，禁止访问",
	})
}

// Get 获取用户列表
func (u *UsersController) Get() mvc.Result {
	users, err := u.UserService.QueryAllUsers()
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code":  -100,
				"msg": err.Error(),
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code":  200,
			"msg": "success",
			"data": users,
		},
	}
}
// DeleteById 删除用户
func (u *UsersController) DeleteById(Id int64) mvc.Result {
	err := u.UserService.DeleteUserById(Id)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code":  -100,
				"msg": err.Error(),
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code":  200,
			"msg": "success",
		},
	}
}

// Post 新增用户
func (u *UsersController) Post() mvc.Result {
	var newUserReq struct{
		Username string `json:"username"`
		Password string `json:"password"`
	}
	err := u.Ctx.ReadJSON(&newUserReq)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code":  -100,
				"msg": err.Error(),
			},
		}
	}
	password, _ := datamodels.GeneratePassword(newUserReq.Password)
	newUser := datamodels.User{
		Username:       newUserReq.Username,
		HashedPassword: password,
		IsSuper:        false,
	}
	err = u.UserService.InsertUser(newUser)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code":  -100,
				"msg": "存在相同的用户名",
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code":  200,
			"msg": "success",
		},
	}
}
