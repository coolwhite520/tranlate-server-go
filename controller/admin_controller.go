package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"strings"
	"translate-server/datamodels"
	"translate-server/docker"
	"translate-server/middleware"
	"translate-server/services"
)


type AdminController struct {
	Ctx         iris.Context
	UserService services.UserService
}


func (a *AdminController) BeforeActivation(b mvc.BeforeActivation) {
	b.Router().Use(middleware.CheckLoginMiddleware, middleware.CheckSuperMiddleware, middleware.CheckActivationMiddleware) //  middleware.IsSystemAvailable
	b.Handle("DELETE","/{id: int64}", "DeleteById")
}


// Get 获取用户列表
func (a *AdminController) Get() mvc.Result {
	users, err := a.UserService.QueryAllUsers()
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
func (a *AdminController) DeleteById(Id int64) mvc.Result {
	err := a.UserService.DeleteUserById(Id)
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
func (a *AdminController) Post() mvc.Result {
	var newUserReq struct{
		Username string `json:"username"`
		Password string `json:"password"`
	}
	err := a.Ctx.ReadJSON(&newUserReq)
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
	err = a.UserService.InsertUser(newUser)
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

func (a *AdminController) PostPassword() mvc.Result {
	var newUserReq struct {
		Id    int64 `json:"id"`
		NewPassword    string `json:"new_password"`
	}
	err := a.Ctx.ReadJSON(&newUserReq)
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
	err = a.UserService.UpdateUserPassword(user)
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

func (a *AdminController) PostRepair() mvc.Result{
	docker.GetInstance().SetStatus(docker.RepairingStatus)
	err := docker.GetInstance().StartDockers()
	if err != nil {
		docker.GetInstance().SetStatus(docker.NormalStatus)
		return mvc.Response{
			Object: map[string]interface{}{
				"code": datamodels.HttpDockerServiceException,
				"msg":  datamodels.HttpDockerServiceException.String(),
			},
		}
	}
	docker.GetInstance().SetStatus(docker.NormalStatus)
	return mvc.Response{
		Object: map[string]interface{}{
			"code": datamodels.HttpSuccess,
			"msg":  datamodels.HttpSuccess.String(),
		},
	}
}