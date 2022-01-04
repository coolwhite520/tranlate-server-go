package controller

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"translate-server/config"
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


func (a *AdminController) BeforeActivation(b mvc.BeforeActivation) {
	b.Router().Use(middleware.CheckLoginMiddleware, middleware.CheckSuperMiddleware, middleware.CheckActivationMiddleware) //  middleware.IsSystemAvailable
	b.Handle("DELETE","/{id: int64}", "DeleteById")
	b.Handle("POST","/upload", "PostUploadUpgradeFile")
	b.Handle("POST","/upgrade", "PostUpgradeComponent")
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
		Mark     string `json:"mark"`
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
		Mark:           newUserReq.Mark,
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

func (a *AdminController) PostMark() mvc.Result {
	var newUserReq struct {
		Id    int64 `json:"id"`
		Mark    string `json:"mark"`
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
	var user datamodels.User
	user.Id = newUserReq.Id
	user.Mark = newUserReq.Mark
	err = a.UserService.UpdateUserMark(user)
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

func Filter(comment string, list []string) [] string  {
	var newList []string
	for _, v:= range  list{
		if v == comment {
			continue
		}
		newList = append(newList, v)
	}
	return newList
}

//GetComponents 获取组件列表
func (a *AdminController) GetComponents() mvc.Result {
	type resultData struct{
		Name string `json:"name"`
		CurrentVersion string `json:"current_version"`
		Versions []string `json:"versions"`
	}
	type ResultDataList []resultData
	var retList ResultDataList
	list, err := config.GetInstance().GetComponentList(false)
	for _, v := range list {
		var item resultData
		versions := config.GetInstance().GetCompVersions(v.ImageName)
		item.Versions = versions
		item.Name = v.ImageName
		item.CurrentVersion = v.ImageVersion
		retList = append(retList, item)
	}

	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": datamodels.HttpFileNotFoundError,
				"msg":  err.Error(),
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": datamodels.HttpSuccess,
			"msg": datamodels.HttpSuccess.String(),
			"data": retList,
		},
	}
}

// PostUploadUpgradeFile 升级文件必须是zip格式，压缩包里面包含一个同名的 xxx.dat（记录升级文件的信息也就是ComponentInfo结构） 和一个xxx.tar 文件
func (a *AdminController) PostUploadUpgradeFile() mvc.Result{
	fileName := a.Ctx.FormValue("fileName")
	fileMd5 := a.Ctx.FormValue("fileMd5")
	order, _ := strconv.Atoi(a.Ctx.FormValue("order"))
	total, _ := strconv.Atoi(a.Ctx.FormValue("total"))
	file, _, err := a.Ctx.FormFile("file")
	if err != nil {
		return nil
	}
	tempDir := os.TempDir()
	dir := fmt.Sprintf("%s/%s", tempDir, fileName)
	if !utils.PathExists(dir) {
		os.MkdirAll(dir, 0666)
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
	// 开始合并文件并解压
	if total == order {
		mergeFilePathname := fmt.Sprintf("%s/%s", dir, fileName)
		mergeFile, _ := os.Create(mergeFilePathname)
		for i := 1; i <= total; i++ {
			slicedFilePathName := fmt.Sprintf("%s/%s-%d", dir, fileName, i)
			slicedFile, _ := os.Open(slicedFilePathName)
			_, err := io.Copy(mergeFile, slicedFile)
			if err != nil {
				return mvc.Response{
					Object: map[string]interface{}{
						"code": datamodels.HttpUploadFileError,
						"msg":  "系统错误，文件copy错误",
					},
				}
			}
			slicedFile.Close()
			os.Remove(slicedFilePathName)
		}
		mergeFile.Close()
		// 解压缩zip包 到一个目录中
		err:= utils.Unzip(mergeFilePathname, dir)
		if err != nil {
			return mvc.Response{
				Object: map[string]interface{}{
					"code": datamodels.HttpUploadFileError,
					"msg":  "系统错误，解压文件失败" + err.Error(),
				},
			}
		}
		os.Remove(mergeFilePathname)
		// 解析.dat文件
		var datName string
		fs, _ := ioutil.ReadDir(dir)
		for _, v := range fs {
			// 遍历得到文件名
			if strings.Contains(v.Name(), ".dat"){
				datName = v.Name()
				break
			}
		}
		datConfig := fmt.Sprintf("%s/%s", dir, datName)
		compInfo, err := config.GetInstance().ParseComponentDatFile(datConfig)
		if err != nil {
			return mvc.Response{
				Object: map[string]interface{}{
					"code": datamodels.HttpUploadFileError,
					"msg":  "系统错误，解析配置失败",
				},
			}
		}
		// 移动到指定目录
		desDir := fmt.Sprintf("./components/%s/%s", compInfo.ImageName, compInfo.ImageVersion)
		if utils.PathExists(desDir) {
			err := os.RemoveAll(desDir)
			if err != nil {
				return mvc.Response{
					Object: map[string]interface{}{
						"code": datamodels.HttpUploadFileError,
						"msg":  "系统错误，删除已有组件失败",
					},
				}
			}
		}
		err = os.Rename(dir, desDir)
		if err != nil {
			return mvc.Response{
				Object: map[string]interface{}{
					"code": datamodels.HttpUploadFileError,
					"msg":  "系统错误，移动目录失败",
				},
			}
		}
		return mvc.Response{
			Object: map[string]interface{}{
				"code": datamodels.HttpSuccess,
				"msg":  datamodels.HttpSuccess.String(),
				"data": compInfo,
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


// PostUpgradeComponent 进行组件的升级
func (a *AdminController) PostUpgradeComponent() mvc.Result {
	var newUserReq struct {
		Name          string `json:"name"`
		CurrentVersion string `json:"current_version"`
		UpVersion     string  `json:"up_version"`
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
	// 移除容器
	err = docker.GetInstance().RemoveContainer(newUserReq.Name, newUserReq.CurrentVersion)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": datamodels.HttpDockerServiceException,
				"msg": err.Error(),
			},
		}
	}
	// 移除镜像
	err = docker.GetInstance().RemoveImage(newUserReq.Name, newUserReq.CurrentVersion)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": datamodels.HttpDockerServiceException,
				"msg": err.Error(),
			},
		}
	}
	//
	dat := fmt.Sprintf("./components/%s/%s/%s.dat", newUserReq.Name, newUserReq.UpVersion, newUserReq.Name)
	compInfo, err := config.GetInstance().ParseComponentDatFile(dat)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": datamodels.HttpFileOpenError,
				"msg": err.Error(),
			},
		}
	}
	// 加载新的镜像
	err = docker.GetInstance().LoadImage(*compInfo)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": datamodels.HttpDockerServiceException,
				"msg": err.Error(),
			},
		}
	}
	// 启动容器
	err = docker.GetInstance().StartContainer(*compInfo)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": datamodels.HttpDockerServiceException,
				"msg": err.Error(),
			},
		}
	}
	// 如果是mysql 模块升级，那么需要重新初始化数据库,
	// ?????现在不需要这么做了，因为已经挂载到本地文件系统了。！！！！！
	//if compInfo.ImageName == "mysql" {
	//	services.InitDb()
	//}
	// 修改versions.ini
	config.GetInstance().SetSectionKeyValue("components", newUserReq.Name, newUserReq.UpVersion)

	//重启dockerd防止由于firewalld导致的dockerd链条缺失的问题
	// 可能需要手动重启，不知道为什么golang的cmd调用不好使
	return mvc.Response{
		Object: map[string]interface{}{
			"code": datamodels.HttpSuccess,
			"msg": datamodels.HttpSuccess.String(),
		},
	}
}

