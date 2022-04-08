package services

import (
	"encoding/base64"
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	log "github.com/sirupsen/logrus"
	"github.com/thinkeridea/go-extend/exnet"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"translate-server/config"
	"translate-server/constant"
	"translate-server/datamodels"
	"translate-server/docker"
	"translate-server/middleware"
	"translate-server/structs"
	"translate-server/systeminfo"
	"translate-server/utils"
)

type AdminService interface {
	GetAllTransRecords(ctx iris.Context) mvc.Result
	GetUserList() mvc.Result
	GetUserOperatorRecords(offset, count int) mvc.Result
	DeleteUserOperatorById(Id int64) mvc.Result
	DeleteAllUserOperator() mvc.Result
	DeleteById(Id int64) mvc.Result
	AddNewUser(ctx iris.Context) mvc.Result
	ModifyMark(ctx iris.Context) mvc.Result
	ModifyPassword(ctx iris.Context) mvc.Result
	Repair() mvc.Result
	GetComponents() mvc.Result
	UploadUpgradeFile(ctx iris.Context) mvc.Result
	UpgradeComponent(ctx iris.Context) mvc.Result
	DeleteIpTableRecord(Id int64) mvc.Result
	AddIpTableRecord(ctx iris.Context) mvc.Result
	GetIpTableType() mvc.Result
	GetIpTableRecords() mvc.Result
	SetIpTableType(ctx iris.Context) mvc.Result
	GetSystemCpuMemDiskDetail() mvc.Result
	GetSysInfo(ctx iris.Context) mvc.Result
	LookupContainerLogs(ctx iris.Context)
	LookupSystemLogs(ctx iris.Context)
}

func  NewAdminService() AdminService {
	return &adminService{}
}

type adminService struct {
	
}

// GetAllTransRecords 获取所有用户的翻译记录
func (a *adminService) GetAllTransRecords(ctx iris.Context) mvc.Result {
	offset := ctx.Params().GetIntDefault("offset", 0)
	count := ctx.Params().GetIntDefault("count", 0)
	total, records, err := datamodels.QueryTranslateRecords(offset, count)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpMysqlQueryError,
				"msg":  err.Error(),
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": constant.HttpSuccess,
			"msg":  constant.HttpSuccess.String(),
			"data": map[string]interface{}{
				"list": records,
				"total": total,
			},
		},
	}
}

// GetUserList 获取用户列表
func (a *adminService) GetUserList() mvc.Result {
	users, err := datamodels.QueryAllUsers()
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpMysqlQueryError,
				"msg":  err.Error(),
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": constant.HttpSuccess,
			"msg":  constant.HttpSuccess.String(),
			"data": users,
		},
	}
}

// GetUserOperatorRecords 获取用户操作记录
func (a *adminService) GetUserOperatorRecords(offset, count int) mvc.Result {
	total, records, err := datamodels.QueryUserOperatorRecords(offset, count )
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpMysqlQueryError,
				"msg":  err.Error(),
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": constant.HttpSuccess,
			"msg":  constant.HttpSuccess.String(),
			"data": map[string]interface{}{
				"list": records,
				"total": total,
			},
		},
	}
}

// DeleteUserOperatorById 删除用户
func (a *adminService) DeleteUserOperatorById(Id int64) mvc.Result {
	err := datamodels.DeleteUserOperatorRecord(Id)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpMysqlDelError,
				"msg":  err.Error(),
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": constant.HttpSuccess,
			"msg":  constant.HttpSuccess.String(),
		},
	}
}

// DeleteAllUserOperator 删除用户操作记录
func (a *adminService) DeleteAllUserOperator() mvc.Result {
	err := datamodels.DeleteAllUserOperatorRecords()
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpMysqlDelError,
				"msg":  err.Error(),
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": constant.HttpSuccess,
			"msg":  constant.HttpSuccess.String(),
		},
	}
}
// DeleteById 删除用户
func (a *adminService) DeleteById(Id int64) mvc.Result {
	err := datamodels.DeleteUserById(Id)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpMysqlDelError,
				"msg":  err.Error(),
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": constant.HttpSuccess,
			"msg":  constant.HttpSuccess.String(),
		},
	}
}

// Post 新增用户
func (a *adminService) AddNewUser(ctx iris.Context) mvc.Result {
	var newUserReq struct{
		Username string `json:"username"`
		Password string `json:"password"`
		Mark     string `json:"mark"`
	}
	err := ctx.ReadJSON(&newUserReq)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpMysqlAddError,
				"msg":  err.Error(),
			},
		}
	}
	newUserReq.Username = strings.Trim(newUserReq.Username, " ")
	newUserReq.Password = strings.Trim(newUserReq.Password, " ")

	if len(newUserReq.Username) < 5 {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpMysqlAddError,
				"msg":  "用户名必须大于4位",
			},
		}
	}
	if len(newUserReq.Password) < 5 {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpMysqlAddError,
				"msg":  "密码必须大于4位",
			},
		}
	}
	password, _ := structs.GeneratePassword(newUserReq.Password)
	newUser := structs.User{
		Username:       newUserReq.Username,
		HashedPassword: password,
		Mark:           newUserReq.Mark,
		IsSuper:        false,
	}
	err = datamodels.InsertUser(newUser)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpUsersExistSameUserNameError,
				"msg":  "存在相同的用户名",
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": constant.HttpSuccess,
			"msg":  constant.HttpSuccess.String(),
		},
	}
}

// ModifyMark 修改个人备注
func (a *adminService) ModifyMark(ctx iris.Context) mvc.Result {
	var newUserReq struct {
		Id    int64 `json:"id"`
		Mark    string `json:"mark"`
	}
	err := ctx.ReadJSON(&newUserReq)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpJsonParseError,
				"msg":  constant.HttpJsonParseError.String(),
			},
		}
	}
	var user structs.User
	user.Id = newUserReq.Id
	user.Mark = newUserReq.Mark
	err = datamodels.UpdateUserMark(user)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpUserUpdatePwdError,
				"msg":  err.Error(),
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": constant.HttpSuccess,
			"msg":  constant.HttpSuccess.String(),
		},
	}
}

// PostPassword 修改密码
func (a *adminService) ModifyPassword(ctx iris.Context) mvc.Result {
	var newUserReq struct {
		Id    int64 `json:"id"`
		NewPassword    string `json:"new_password"`
	}
	err := ctx.ReadJSON(&newUserReq)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpJsonParseError,
				"msg":  constant.HttpJsonParseError.String(),
			},
		}
	}
	if len(newUserReq.NewPassword) < 5 {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpMysqlAddError,
				"msg":  "密码必须大于4位",
			},
		}
	}
	var user structs.User
	user.Id = newUserReq.Id
	user.HashedPassword, _ = structs.GeneratePassword(newUserReq.NewPassword)
	err = datamodels.UpdateUserPassword(user)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpUserUpdatePwdError,
				"msg":  err.Error(),
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": constant.HttpSuccess,
			"msg":  constant.HttpSuccess.String(),
		},
	}
}

//Repair 管理员调用系统修复
func (a *adminService) Repair() mvc.Result{
	err := docker.GetInstance().CreatePrivateNetwork()
	docker.GetInstance().SetStatus(docker.RepairingStatus)
	err = docker.GetInstance().StartDockers()
	if err != nil {
		docker.GetInstance().SetStatus(docker.NormalStatus)
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpDockerServiceException,
				"msg":  constant.HttpDockerServiceException.String(),
			},
		}
	}
	docker.GetInstance().SetStatus(docker.NormalStatus)
	return mvc.Response{
		Object: map[string]interface{}{
			"code": constant.HttpSuccess,
			"msg":  constant.HttpSuccess.String(),
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

type ContainerState int64
const (
	ContainerAllGood ContainerState =  iota
	ContainerNotRun
	ContainerNotInPrivateNet
)

func (h ContainerState) String() string {
	switch h {
	case ContainerAllGood:
		return "运行良好"
	case ContainerNotRun:
		return "已停止"
	case ContainerNotInPrivateNet:
		return "网络错误"
	default:
		return ""
	}
}

type resultData struct{
	Name string `json:"name"`
	CurrentVersion string `json:"current_version"`
	Versions []string `json:"versions"`
	CompsState ContainerState `json:"comps_state"`
	CompsStateDescribe string `json:"comps_state_describe"`
}
type ResultDataList []resultData

func (r ResultDataList) Len() int           { return len(r) }
func (r ResultDataList) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r ResultDataList) Less(i, j int) bool { return r[i].Name < r[j].Name }

//GetComponents 获取组件列表
func (a *adminService) GetComponents() mvc.Result {
	var retList ResultDataList
	list, err := config.GetInstance().GetComponentList(false)
	for _, v := range list {
		var item resultData
		versions := config.GetInstance().GetCompVersions(v.ImageName)
		item.Versions = versions
		item.Name = v.ImageName
		item.CurrentVersion = v.ImageVersion
		running, _ := docker.GetInstance().IsContainerRunning(v.ImageName, v.ImageVersion)
		net, _ := docker.GetInstance().IsInPrivateNet(v.ImageName)
		if running && net {
			item.CompsState = ContainerAllGood
		} else if running && !net {
			item.CompsState = ContainerNotInPrivateNet
		} else {
			item.CompsState = ContainerNotRun
		}
		item.CompsStateDescribe = item.CompsState.String()
		retList = append(retList, item)
	}
	sort.Sort(retList)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpFileNotFoundError,
				"msg":  err.Error(),
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": constant.HttpSuccess,
			"msg":  constant.HttpSuccess.String(),
			"data": retList,
		},
	}
}

func (a *adminService) GetSystemCpuMemDiskDetail() mvc.Result {
	info := systeminfo.GetSystemInfo()
	return mvc.Response{
		Object: map[string]interface{}{
			"code": constant.HttpSuccess,
			"msg":  constant.HttpSuccess.String(),
			"data": info,
		},
	}
}

func (a *adminService) LookupSystemLogs(ctx iris.Context) {
	srcFileName := "./logs"
	now := time.Now().Format("2006_01_02_15_04_05")
	fileName := fmt.Sprintf("trans_logs_%s.zip", now)
	desFileName := fmt.Sprintf("%s/%s", os.TempDir(), fileName)
	err := utils.ZipFile(srcFileName, desFileName)
	if err != nil {
		ctx.JSON(
			map[string]interface{}{
				"code": constant.HttpFileOpenError,
				"msg":  err.Error(),
			},
		)
		return
	}
	ctx.SendFile(desFileName, fileName)

}

func (a *adminService) LookupContainerLogs(ctx iris.Context) {
	var newUserReq struct {
		Name          string `json:"name"`
	}
	err := ctx.ReadJSON(&newUserReq)
	if err != nil {
		ctx.JSON(
			map[string]interface{}{
				"code": constant.HttpJsonParseError,
				"msg":  err.Error(),
			},
		)
		return
	}
	name := newUserReq.Name
	logs, err := docker.GetInstance().LogsContainer(name)
	if err != nil {
		ctx.JSON(
			map[string]interface{}{
				"code": constant.HttpFileOpenError,
				"msg":  err.Error(),
			},
		)
		return
	}
	logStr := base64.StdEncoding.EncodeToString(logs)
	now := time.Now().Format("2006_01_02_15_04_05")
	tempDir := os.TempDir()
	dir := fmt.Sprintf("%s/%s/%s", tempDir, name, now)
	if !utils.PathExists(dir) {
		os.MkdirAll(dir, os.ModePerm)
	}
	fileName := fmt.Sprintf("%s/%s.txt", dir, name)
	desFileName := fmt.Sprintf("%s/%s.zip", dir, name)
	err = ioutil.WriteFile(fileName, []byte(logStr), os.ModePerm)
	if err != nil {
		ctx.JSON(
			map[string]interface{}{
				"code": constant.HttpFileOpenError,
				"msg":  err.Error(),
			},
		)
		return
	}
	err = utils.ZipFile(fileName, desFileName)
	if err != nil {
		ctx.JSON(
			map[string]interface{}{
				"code": constant.HttpFileOpenError,
				"msg":  err.Error(),
			},
		)
		return
	}
	ctx.SendFile(desFileName, name + ".zip")
}

// UploadUpgradeFile 升级文件必须是zip格式，压缩包里面包含一个同名的 xxx.dat（记录升级文件的信息也就是ComponentInfo结构） 和一个xxx.tar 文件
func (a *adminService) UploadUpgradeFile(ctx iris.Context) mvc.Result{
	fileName := ctx.FormValue("fileName")
	fileMd5 := ctx.FormValue("fileMd5")
	order, _ := strconv.Atoi(ctx.FormValue("order"))
	total, _ := strconv.Atoi(ctx.FormValue("total"))
	file, _, err := ctx.FormFile("file")
	if err != nil {
		return nil
	}
	tempDir := os.TempDir()
	dir := fmt.Sprintf("%s/%s", tempDir, fileName)
	if !utils.PathExists(dir) {
		os.MkdirAll(dir, os.ModePerm)
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
				"code": constant.HttpUploadFileError,
				"msg":  err.Error(),
			},
		}
	}

	md5, err := utils.GetFileMd5(filePathName)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpUploadFileError,
				"msg":  err.Error(),
			},
		}
	}
	if fileMd5 != md5 {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpUploadFileError,
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
						"code": constant.HttpUploadFileError,
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
					"code": constant.HttpUploadFileError,
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
					"code": constant.HttpUploadFileError,
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
						"code": constant.HttpUploadFileError,
						"msg":  "系统错误，删除已有组件失败",
					},
				}
			}
		}
		err = os.Rename(dir, desDir)
		if err != nil {
			return mvc.Response{
				Object: map[string]interface{}{
					"code": constant.HttpUploadFileError,
					"msg":  "系统错误，移动目录失败",
				},
			}
		}
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpSuccess,
				"msg":  constant.HttpSuccess.String(),
				"data": compInfo,
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": constant.HttpSuccess,
			"msg":  constant.HttpSuccess.String(),
		},
	}
}


// UpgradeComponent 进行组件的升级
func (a *adminService) UpgradeComponent(ctx iris.Context) mvc.Result {
	var newUserReq struct {
		Name          string `json:"name"`
		CurrentVersion string `json:"current_version"`
		UpVersion     string  `json:"up_version"`
	}
	err := ctx.ReadJSON(&newUserReq)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpJsonParseError,
				"msg":  constant.HttpJsonParseError.String(),
			},
		}
	}
	// 移除容器
	err = docker.GetInstance().RemoveContainer(newUserReq.Name, newUserReq.CurrentVersion)
	if err != nil {
		log.Errorln(err)
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpDockerServiceException,
				"msg":  err.Error(),
			},
		}
	}
	// 移除镜像
	err = docker.GetInstance().RemoveImage(newUserReq.Name, newUserReq.CurrentVersion)
	if err != nil {
		log.Errorln(err)
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpDockerServiceException,
				"msg":  err.Error(),
			},
		}
	}
	//
	dat := fmt.Sprintf("./components/%s/%s/%s.dat", newUserReq.Name, newUserReq.UpVersion, newUserReq.Name)
	compInfo, err := config.GetInstance().ParseComponentDatFile(dat)
	if err != nil {
		log.Errorln(err)
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpFileOpenError,
				"msg":  err.Error(),
			},
		}
	}
	// 加载新的镜像
	err = docker.GetInstance().LoadImage(*compInfo)
	if err != nil {
		log.Errorln(err)
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpDockerServiceException,
				"msg":  err.Error(),
			},
		}
	}
	// 启动容器
	id, err := docker.GetInstance().CreateAndStartContainer(*compInfo)
	if err != nil {
		log.Errorln(err)
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpDockerServiceException,
				"msg":  err.Error(),
			},
		}
	}
	//添加到networking里面
	if compInfo.ImageName != "web" {
		err = docker.GetInstance().JoinPrivateNetwork(id)
		if err != nil {
			log.Errorln(err)
			return mvc.Response{
				Object: map[string]interface{}{
					"code": constant.HttpDockerServiceException,
					"msg":  err.Error(),
				},
			}
		}
	}

	// 修改versions.ini
	config.GetInstance().SetSectionKeyValue("components", newUserReq.Name, newUserReq.UpVersion)
	config.GetInstance().GetComponentList(true)

	// 如果是mysql 、 redis 模块升级，那么需要重新初始化数据库,因为之前的连接已经断开了
	//if compInfo.ImageName == "mysql" {
	//	datamodels.InitMysql()
	//}
	//if compInfo.ImageName == "redis" {
	//	datamodels.InitRedis()
	//}

	//重启dockerd防止由于firewalld导致的dockerd链条缺失的问题
	// 可能需要手动重启，不知道为什么golang的cmd调用不好使
	return mvc.Response{
		Object: map[string]interface{}{
			"code": constant.HttpSuccess,
			"msg":  constant.HttpSuccess.String(),
		},
	}
}

func (a *adminService) AddIpTableRecord(ctx iris.Context) mvc.Result {
	var record structs.IpTableRecord
	err := ctx.ReadJSON(&record)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpJsonParseError,
				"msg":  constant.HttpJsonParseError.String(),
			},
		}
	}
	err = datamodels.AddIpTblRecord(record)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpMysqlAddError,
				"msg":  err.Error(),
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": constant.HttpSuccess,
			"msg":  constant.HttpSuccess.String(),
		},
	}
}
func (a *adminService) DeleteIpTableRecord(Id int64) mvc.Result {
	err := datamodels.DelIpTblRecord(Id)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpMysqlDelError,
				"msg":  err.Error(),
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": constant.HttpSuccess,
			"msg":  constant.HttpSuccess.String(),
		},
	}
}
func (a *adminService) GetIpTableRecords() mvc.Result  {
	records, err := datamodels.QueryIpTblRecords()
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpMysqlQueryError,
				"msg":  err.Error(),
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": constant.HttpSuccess,
			"msg":  constant.HttpSuccess.String(),
			"data": records,
		},
	}
}

func (a *adminService) GetIpTableType() mvc.Result {
	tableType, err := datamodels.GetIpTableType()
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpJsonParseError,
				"msg":  err,
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": constant.HttpSuccess,
			"msg":  constant.HttpSuccess.String(),
			"data": tableType,
		},
	}
}

func (a *adminService) SetIpTableType(ctx iris.Context) mvc.Result {
	var record struct{
		Type string `json:"type"`
	}
	err := ctx.ReadJSON(&record)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpJsonParseError,
				"msg":  constant.HttpJsonParseError.String(),
			},
		}
	}
	ipAddr := exnet.ClientIP(ctx.Request())
	// 白名单的时候先把自己加入到名单中
	if record.Type == "white" {
		records, err := datamodels.QueryIpTblRecords()
		if err != nil {
			return mvc.Response{
				Object: map[string]interface{}{
					"code": constant.HttpMysqlQueryError,
					"msg":  err,
				},
			}
		}
		if!middleware.IsInWhiteList(ipAddr, records) {
			err := datamodels.AddIpTblRecord(structs.IpTableRecord{
				Ip:   ipAddr,
				Type: "white",
			})
			if err != nil {
				return mvc.Response{
					Object: map[string]interface{}{
						"code": constant.HttpMysqlAddError,
						"msg":  err,
					},
				}
			}
		}
	}
	_, err = datamodels.SetIpTableType(record.Type)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpJsonParseError,
				"msg":  err,
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": constant.HttpSuccess,
			"msg":  constant.HttpSuccess.String(),
		},
	}
}

func (a *adminService) GetSysInfo(ctx iris.Context) mvc.Result {
	newActivation := datamodels.NewActivationModel()
	activationInfo, state := newActivation.ParseKeystoreFile()
	if state != constant.HttpSuccess {
		ctx.JSON(
			map[string]interface{}{
				"code":      state,
				"msg":       state.String(),
			})
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": constant.HttpSuccess,
			"msg":  constant.HttpSuccess.String(),
			"data": activationInfo,
		},
	}
}