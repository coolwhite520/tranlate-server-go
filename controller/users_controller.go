package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"translate-server/datamodels"
	"translate-server/middleware"
	"translate-server/services"
)


type UsersController struct {
	Ctx         iris.Context
	UserService services.UserService
}


func (u *UsersController) BeforeActivation(b mvc.BeforeActivation) {
	b.Router().Use(middleware.CheckActivationMiddleware, middleware.IsSystemAvailable, middleware.CheckLoginMiddleware, middleware.CheckSuperMiddleware)
	b.Handle("DELETE","/{id: int64}", "DeleteById")
}


// Get 获取用户列表
func (u *UsersController) Get() mvc.Result {
	users, err := u.UserService.QueryAllUsers()
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code":  datamodels.HttpUsersQueryError,
				"msg": err.Error(),
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code":  datamodels.HttpSuccess,
			"msg": datamodels.HttpSuccess.String(),
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
				"code":  datamodels.HttpUsersDeleteError,
				"msg": err.Error(),
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code":  datamodels.HttpSuccess,
			"msg": datamodels.HttpSuccess.String(),
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
				"code":  datamodels.HttpUsersAddError,
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
				"code":  datamodels.HttpUsersExistSameUserNameError,
				"msg": "存在相同的用户名",
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code":  datamodels.HttpSuccess,
			"msg": datamodels.HttpSuccess.String(),
		},
	}
}
