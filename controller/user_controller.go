package controller

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"github.com/thinkeridea/go-extend/exnet"
	"translate-server/config"
	"translate-server/datamodels"
	"translate-server/middleware"
	"translate-server/services"
)

type UserController struct {
	Ctx         iris.Context
	UserService services.UserService
}

func (u *UserController) BeforeActivation(b mvc.BeforeActivation) {
	//b.Router().Use(middleware.CheckActivationMiddleware, middleware.IsSystemAvailable)
	// 只有登录以后，才可以进行密码修改
	b.Handle("POST", "/password", "PostPassword", middleware.CheckLoginMiddleware)
	b.Handle("POST", "/logoff", "PostLogoff", middleware.CheckLoginMiddleware)
	b.Handle("GET", "/favor", "GetQueryUserFavor", middleware.CheckLoginMiddleware)
	b.Handle("POST", "/favor", "PostAddUserFavor", middleware.CheckLoginMiddleware)
}

func (u *UserController) PostAddUserFavor() mvc.Result {
	var newUserReq struct {
		Favor string `json:"favor"`
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
	a := u.Ctx.Values().Get("User")
	user, _ := (a).(datamodels.User)
	err = u.UserService.InsertOrReplaceUserFavor(user.Id, newUserReq.Favor)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": datamodels.HttpMysqlAddError,
				"msg": err.Error(),
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

func (u *UserController) GetQueryUserFavor() mvc.Result {
	a := u.Ctx.Values().Get("User")
	user, _ := (a).(datamodels.User)
	favor, err := u.UserService.QueryUserFavorById(user.Id)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": datamodels.HttpMysqlQueryError,
				"msg": err.Error(),
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": datamodels.HttpSuccess,
			"msg":  datamodels.HttpSuccess.String(),
			"data": favor,
		},
	}
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
				"code": datamodels.HttpJsonParseError,
				"msg": datamodels.HttpJsonParseError.String(),
			},
		}
	}
	if newUserReq.NewPassword != newUserReq.SecondPassword {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": datamodels.HttpUserTwicePwdNotSame,
				"msg": datamodels.HttpUserTwicePwdNotSame.String(),
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
		} else {
			return mvc.Response{
				Object: map[string]interface{}{
					"code": datamodels.HttpUserPwdError,
					"msg":  "原始密码输入有误，请重新输入",
				},
			}
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": datamodels.HttpUserExpired,
			"msg":  datamodels.HttpUserExpired.String(),
		},
	}
}

func (u *UserController) PostLogoff() mvc.Result {
	var newUserReq struct {
		UserId int64 `json:"user_id"`
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

	record := datamodels.UserOperatorRecord{
		UserId:   newUserReq.UserId,
		Ip:       exnet.ClientIP(u.Ctx.Request()),
		Operator: "logoff",
	}
	u.UserService.AddUserOperatorRecord(record)

	return mvc.Response{}
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
				"code": datamodels.HttpJsonParseError,
				"msg": datamodels.HttpJsonParseError.String(),
			},
		}
	}
	user, b := u.UserService.CheckUser(newUserReq.Username, newUserReq.Password)
	if user == nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": datamodels.HttpUserNoThisUserError,
				"msg": datamodels.HttpUserNoThisUserError.String(),
			},
		}
	}
	if b {
		token, _, err := middleware.GenerateToken(*user)
		if err != nil {
			return mvc.Response{
				Object: map[string]interface{}{
					"code": datamodels.HttpJwtTokenGenerateError,
					"msg":  "服务器生成JWT错误",
				},
			}
		}
		//记录到操作表中
		record := datamodels.UserOperatorRecord{
			UserId:   user.Id,
			Ip:       exnet.ClientIP(u.Ctx.Request()),
			Operator: "login",
		}
		u.UserService.AddUserOperatorRecord(record)
		//Authorization: Bearer $token
		ver := config.GetInstance().GetSystemVer()
		u.Ctx.Header("Authorization", fmt.Sprintf("Bearer %s", token))
		return mvc.Response{
			Object: map[string]interface{}{
				"code": datamodels.HttpSuccess,
				"msg":  datamodels.HttpSuccess.String(),
				"user": map[string]interface{}{
					"avatar": "",
					"name": user.Username,
					"ip": record.Ip,
					"user_id": user.Id,
					"isSuper": user.IsSuper,
					"sysVer": ver,
				},
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": datamodels.HttpUserNameOrPwdError,
			"msg":  "密码错误",
		},
	}
}

