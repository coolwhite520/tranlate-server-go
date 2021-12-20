package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"strings"
	"syscall"
	"translate-server/datamodels"
	"translate-server/middleware"
	"translate-server/services"
)


type UsersController struct {
	Ctx         iris.Context
	UserService services.UserService
}


func (u *UsersController) BeforeActivation(b mvc.BeforeActivation) {
	b.Router().Use(middleware.CheckLoginMiddleware, middleware.CheckSuperMiddleware, middleware.CheckActivationMiddleware, middleware.IsSystemAvailable)
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
	newUserReq.Username = strings.Trim(newUserReq.Username, " ")
	newUserReq.Password = strings.Trim(newUserReq.Password, " ")

	if len(newUserReq.Username) < 5 {
		return mvc.Response{
			Object: map[string]interface{}{
				"code":  datamodels.HttpUsersAddError,
				"msg": "用户名必须大于4位",
			},
		}
	}
	if len(newUserReq.Password) < 5 {
		return mvc.Response{
			Object: map[string]interface{}{
				"code":  datamodels.HttpUsersAddError,
				"msg": "密码必须大于4位",
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

func (u *UsersController) PostPassword() mvc.Result {
	var newUserReq struct {
		Id    int64 `json:"id"`
		NewPassword    string `json:"new_password"`
	}
	err := u.Ctx.ReadJSON(&newUserReq)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": datamodels.HttpJsonParseError,
				"msg": datamodels.HttpJsonParseError.String(),
			},
		}
	}
	if len(newUserReq.NewPassword) < 5 {
		return mvc.Response{
			Object: map[string]interface{}{
				"code":  datamodels.HttpUsersAddError,
				"msg": "密码必须大于4位",
			},
		}
	}
	var user datamodels.User
	user.Id = newUserReq.Id
	user.HashedPassword, _ = datamodels.GeneratePassword(newUserReq.NewPassword)
	err = u.UserService.UpdateUserPassword(user)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": datamodels.HttpUserUpdatePwdError,
				"msg":  err.Error(),
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": datamodels.HttpSuccess,
			"msg":  datamodels.HttpSuccess.String(),
		},
	}
}

func (u *UsersController) PostRestart() mvc.Result{
	datamodels.GlobalChannel <- syscall.SIGUSR2
	return mvc.Response{
		Object: map[string]interface{}{
			"code": datamodels.HttpSuccess,
			"msg":  datamodels.HttpSuccess.String(),
		},
	}
}