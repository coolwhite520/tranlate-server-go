package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"translate-server/middleware"
	"translate-server/services"
)

type AdminController struct {
	Ctx          iris.Context
	AdminService services.AdminService
}

func (a *AdminController) BeforeActivation(b mvc.BeforeActivation) {
	b.Router().Use(middleware.CheckLoginMiddleware, middleware.CheckSuperMiddleware, middleware.CheckActivationMiddleware) //  middleware.IsSystemAvailable
	b.Handle("DELETE", "/{id: int64}", "DeleteById")
	b.Handle("GET", "/ops/{offset: int}/{count: int}", "GetUserOperatorRecords")
	b.Handle("GET", "/sysinfo", "GetSysInfo")
	b.Handle("DELETE", "/ops/{id: int64}", "DeleteUserOperatorById")
	b.Handle("DELETE", "/ops", "DeleteAllUserOperator")
	b.Handle("POST", "/upload", "UploadUpgradeFile")
	b.Handle("POST", "/upgrade", "UpgradeComponent")
	b.Handle("GET", "/ip_table", "GetIpTableRecords")
	b.Handle("POST", "/ip_table", "AddIpTableRecord")
	b.Handle("DELETE", "/ip_table/{id: int64}", "DeleteIpTableRecord")
	b.Handle("POST", "/ip_table_type", "SetIpTableType")
	b.Handle("GET", "/ip_table_type", "GetIpTableType")
	b.Handle("GET", "/all_records/{offset: uint64}/{count: uint64}", "GetAllTransRecords")
}
//GetSysInfo 获取系统信息
func (a *AdminController) GetSysInfo() mvc.Result {
	return a.AdminService.GetSysInfo(a.Ctx)
}

// GetAllTransRecords 获取所有用户的翻译记录
func (a *AdminController) GetAllTransRecords() mvc.Result {
	return a.AdminService.GetAllTransRecords(a.Ctx)
}

// Get 获取用户列表
func (a *AdminController) Get() mvc.Result {
	return a.AdminService.GetUserList()
}

// GetUserOperatorRecords 获取用户操作记录
func (a *AdminController) GetUserOperatorRecords(offset, count int) mvc.Result {
	return a.AdminService.GetUserOperatorRecords(offset, count)
}

// DeleteUserOperatorById 删除用户
func (a *AdminController) DeleteUserOperatorById(Id int64) mvc.Result {
	return a.AdminService.DeleteUserOperatorById(Id)
}

// DeleteAllUserOperator 删除用户操作记录
func (a *AdminController) DeleteAllUserOperator() mvc.Result {
	return a.AdminService.DeleteAllUserOperator()
}

// DeleteById 删除用户
func (a *AdminController) DeleteById(Id int64) mvc.Result {
	return a.AdminService.DeleteById(Id)
}

// Post 新增用户
func (a *AdminController) Post() mvc.Result {
	return a.AdminService.AddNewUser(a.Ctx)
}

// PostMark 修改个人备注
func (a *AdminController) PostMark() mvc.Result {
	return a.AdminService.ModifyMark(a.Ctx)
}

// PostPassword 修改密码
func (a *AdminController) PostPassword() mvc.Result {
	return a.AdminService.ModifyPassword(a.Ctx)
}

//PostRepair 管理员调用系统修复
func (a *AdminController) PostRepair() mvc.Result {
	return a.AdminService.Repair()
}

//GetComponents 获取组件列表
func (a *AdminController) GetComponents() mvc.Result {
	return a.AdminService.GetComponents()
}

// UploadUpgradeFile 升级文件必须是zip格式，压缩包里面包含一个同名的 xxx.dat（记录升级文件的信息也就是ComponentInfo结构） 和一个xxx.tar 文件
func (a *AdminController) UploadUpgradeFile() mvc.Result {
	return a.AdminService.UploadUpgradeFile(a.Ctx)
}

// UpgradeComponent 进行组件的升级
func (a *AdminController) UpgradeComponent() mvc.Result {
	return a.AdminService.UpgradeComponent(a.Ctx)
}

// AddIpTableRecord 新增ip记录
func (a *AdminController) AddIpTableRecord() mvc.Result {
	return a.AdminService.AddIpTableRecord(a.Ctx)
}

// DeleteIpTableRecord  删除ip记录
func (a *AdminController) DeleteIpTableRecord(Id int64) mvc.Result {
	return a.AdminService.DeleteIpTableRecord(Id)
}

// GetIpTableRecords  获取ip记录
func (a *AdminController) GetIpTableRecords() mvc.Result {
	return a.AdminService.GetIpTableRecords()
}

//GetIpTableType 获取ip黑白类型
func (a *AdminController) GetIpTableType() mvc.Result {
	return a.AdminService.GetIpTableType()
}

//SetIpTableType 设置ip黑白类型
func (a *AdminController) SetIpTableType() mvc.Result {
	return a.AdminService.SetIpTableType(a.Ctx)
}
