package controller

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"io"
	"os"
	"strconv"
	"strings"
	"translate-server/datamodels"
	"translate-server/docker"
	"translate-server/middleware"
	"translate-server/services"
	"translate-server/utils"
)


type AdminController struct {
	Ctx         iris.Context
	UserService services.UserService
}

const UpgradeDirNew = "./upgrade_new"
const UpgradeDirBak = "./upgrade_bak"

func (a *AdminController) BeforeActivation(b mvc.BeforeActivation) {
	b.Router().Use(middleware.CheckLoginMiddleware, middleware.CheckSuperMiddleware, middleware.CheckActivationMiddleware) //  middleware.IsSystemAvailable
	b.Handle("DELETE","/{id: int64}", "DeleteById")
	b.Handle("POST","/upload", "PostUploadUpgradeFile")

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

func (a *AdminController) PostUploadUpgradeFile() mvc.Result{
	fileName := a.Ctx.FormValue("fileName")
	fileMd5 := a.Ctx.FormValue("fileMd5")
	order, _ := strconv.Atoi(a.Ctx.FormValue("order"))
	total, _ := strconv.Atoi(a.Ctx.FormValue("total"))
	file, _, err := a.Ctx.FormFile("file")
	if err != nil {
		return nil
	}
	dir := fmt.Sprintf("%s/%s", UpgradeDirNew, fileName)
	if !utils.PathExists(dir) {
		os.MkdirAll(dir, 0777)
	}
	filePathName := fmt.Sprintf("%s/%s-%d", dir, fileName, order)
	create, err := os.Create(filePathName)
	if err != nil {
		return nil
	}
	_, err = io.Copy(create, file)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": datamodels.HttpUploadFileError,
				"msg":  err.Error(),
			},
		}
	}

	md5, err := utils.GetFileMd5(filePathName)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": datamodels.HttpUploadFileError,
				"msg":  err.Error(),
			},
		}
	}

	if fileMd5 != md5 {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": datamodels.HttpUploadFileError,
				"msg":  "md5不一致，请重新上传",
			},
		}
	}
	if total == order {
		mergeFile := fmt.Sprintf("%s/%s", dir, fileName)
		f, _ := os.Create(mergeFile)
		for i := 1; i <= total; i++ {
			filePathName := fmt.Sprintf("%s/%s-%d", dir, fileName, i)
			tempf, _ := os.Open(filePathName)
			io.Copy(f, tempf)
		}
		f.Close()
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": datamodels.HttpSuccess,
			"msg":  datamodels.HttpSuccess.String(),
		},
	}
}

