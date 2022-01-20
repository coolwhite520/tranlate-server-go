package services

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"github.com/thinkeridea/go-extend/exnet"
	"translate-server/config"
	"translate-server/datamodels"
	"translate-server/middleware"
	"translate-server/structs"
)


type UserService interface {
	PostLogin(ctx iris.Context) mvc.Result
	PostAddUserFavor(ctx iris.Context) mvc.Result
	GetQueryUserFavor(ctx iris.Context) mvc.Result
	PostPassword(ctx iris.Context) mvc.Result
	PostLogoff(ctx iris.Context) mvc.Result
}

func NewUserService() UserService  {
	return &userService{}
}

type userService struct {

}
func (u *userService) PostAddUserFavor(ctx iris.Context) mvc.Result {
	var newUserReq struct {
		Favor string `json:"favor"`
	}
	err := ctx.ReadJSON(&newUserReq)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": structs.HttpJsonParseError,
				"msg": structs.HttpJsonParseError.String(),
			},
		}
	}
	a := ctx.Values().Get("User")
	user, _ := (a).(structs.User)

	err = datamodels.InsertOrReplaceUserFavor(user.Id, newUserReq.Favor)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": structs.HttpMysqlAddError,
				"msg": err.Error(),
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": structs.HttpSuccess,
			"msg":  structs.HttpSuccess.String(),
		},
	}
}

func (u *userService) GetQueryUserFavor(ctx iris.Context) mvc.Result {
	a := ctx.Values().Get("User")
	user, _ := (a).(structs.User)
	favor, err := datamodels.QueryUserFavorById(user.Id)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": structs.HttpMysqlQueryError,
				"msg": err.Error(),
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": structs.HttpSuccess,
			"msg":  structs.HttpSuccess.String(),
			"data": favor,
		},
	}
}


// PostPassword /api/user/password
func (u *userService) PostPassword(ctx iris.Context) mvc.Result {
	var newUserReq struct {
		OldPassword    string `json:"old_password"`
		NewPassword    string `json:"new_password"`
		SecondPassword string `json:"second_password"`
	}
	err := ctx.ReadJSON(&newUserReq)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": structs.HttpJsonParseError,
				"msg": structs.HttpJsonParseError.String(),
			},
		}
	}
	if newUserReq.NewPassword != newUserReq.SecondPassword {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": structs.HttpUserTwicePwdNotSame,
				"msg": structs.HttpUserTwicePwdNotSame.String(),
			},
		}
	}
	a := ctx.Values().Get("User")
	if user, ok := (a).(structs.User); ok {
		_, b := datamodels.CheckUser(user.Username, newUserReq.OldPassword)
		if b {
			user.HashedPassword, _ = structs.GeneratePassword(newUserReq.NewPassword)
			err = datamodels.UpdateUserPassword(user)
			if err != nil {
				return mvc.Response{
					Object: map[string]interface{}{
						"code": structs.HttpUserUpdatePwdError,
						"msg":  err.Error(),
					},
				}
			}
			return mvc.Response{
				Object: map[string]interface{}{
					"code": structs.HttpSuccess,
					"msg":  structs.HttpSuccess.String(),
				},
			}
		} else {
			return mvc.Response{
				Object: map[string]interface{}{
					"code": structs.HttpUserPwdError,
					"msg":  "原始密码输入有误，请重新输入",
				},
			}
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": structs.HttpUserExpired,
			"msg":  structs.HttpUserExpired.String(),
		},
	}
}

func (u *userService) PostLogoff(ctx iris.Context) mvc.Result {
	var newUserReq struct {
		UserId int64 `json:"user_id"`
	}
	err := ctx.ReadJSON(&newUserReq)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": structs.HttpJsonParseError,
				"msg": structs.HttpJsonParseError.String(),
			},
		}
	}

	record := structs.UserOperatorRecord{
		UserId:   newUserReq.UserId,
		Ip:       exnet.ClientIP(ctx.Request()),
		Operator: "logoff",
	}
	datamodels.AddUserOperatorRecord(record)

	return mvc.Response{}
}



// PostLogin /api/user/login
func (u *userService) PostLogin(ctx iris.Context) mvc.Result {
	var newUserReq struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	err := ctx.ReadJSON(&newUserReq)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": structs.HttpJsonParseError,
				"msg": structs.HttpJsonParseError.String(),
			},
		}
	}
	user, b := datamodels.CheckUser(newUserReq.Username, newUserReq.Password)
	if user == nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": structs.HttpUserNoThisUserError,
				"msg": structs.HttpUserNoThisUserError.String(),
			},
		}
	}
	if b {
		token, _, err := middleware.GenerateToken(*user)
		if err != nil {
			return mvc.Response{
				Object: map[string]interface{}{
					"code": structs.HttpJwtTokenGenerateError,
					"msg":  "服务器生成JWT错误",
				},
			}
		}
		//记录到操作表中
		record := structs.UserOperatorRecord{
			UserId:   user.Id,
			Ip:       exnet.ClientIP(ctx.Request()),
			Operator: "login",
		}
		datamodels.AddUserOperatorRecord(record)
		//Authorization: Bearer $token
		ver := config.GetInstance().GetSystemVer()
		ctx.Header("Authorization", fmt.Sprintf("Bearer %s", token))
		return mvc.Response{
			Object: map[string]interface{}{
				"code": structs.HttpSuccess,
				"msg":  structs.HttpSuccess.String(),
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
			"code": structs.HttpUserNameOrPwdError,
			"msg":  "密码错误",
		},
	}
}