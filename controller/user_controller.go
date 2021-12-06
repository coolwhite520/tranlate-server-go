package controller

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"translate-server/activation"
	"translate-server/datamodels"
	"translate-server/jwt"
	"translate-server/services"
)

type UserController struct {
	Ctx         iris.Context
	UserService services.UserService
}

func (u *UserController) BeforeActivation(b mvc.BeforeActivation) {
	b.Router().Use(activation.CheckActivationMiddleware)
	// 只有登录以后，才可以进行密码修改
	b.Handle("POST", "/password", "PostPassword", jwt.CheckLoginMiddleware)
}

// PostPassword /api/user/password
func (u *UserController) PostPassword() mvc.Result {
	var newUserReq struct {
		OldPassword    string `json:"old_password"`
		NewPassword    string `json:"new_password"`
		SecondPassword string `json:"second_password"`
	}
	err := u.Ctx.ReadJSON(&newUserReq)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": -100,
				"msg":  err.Error(),
			},
		}
	}
	if newUserReq.NewPassword != newUserReq.SecondPassword {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": -100,
				"msg":  "两次密码不一致，请重新输入",
			},
		}
	}
	a := u.Ctx.Values().Get("User")
	if user, ok := (a).(datamodels.User); ok {
		_, b := u.UserService.CheckUser(user.Username, newUserReq.OldPassword)
		if b {
			user.HashedPassword, _ = datamodels.GeneratePassword(newUserReq.NewPassword)
			err = u.UserService.UpdateUserPassword(user)
			if err != nil {
				return mvc.Response{
					Object: map[string]interface{}{
						"code": -100,
						"msg":  err.Error(),
					},
				}
			}
			return mvc.Response{
				Object: map[string]interface{}{
					"code": 200,
					"msg":  "success, 重定向导login进行重新登录，清理掉header中的信息",
				},
			}
		} else {
			return mvc.Response{
				Object: map[string]interface{}{
					"code": -100,
					"msg":  "原始密码输入有误，请重新输入",
				},
			}
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": -100,
			"msg":  "登录信息有误，请重新登录",
		},
	}
}

// PostLogin /api/user/login
func (u *UserController) PostLogin() mvc.Result {
	var newUserReq struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	err := u.Ctx.ReadJSON(&newUserReq)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": -100,
				"msg":  err.Error(),
			},
		}
	}
	user, b := u.UserService.CheckUser(newUserReq.Username, newUserReq.Password)
	if b {
		token, err := jwt.GenerateToken(user)
		if err != nil {
			return mvc.Response{
				Object: map[string]interface{}{
					"code": 500,
					"msg":  "服务器错误",
				},
			}
		}
		//Authorization: Bearer $token
		u.Ctx.Header("Authorization", fmt.Sprintf("Bearer %s", token))
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

// 修改密码
