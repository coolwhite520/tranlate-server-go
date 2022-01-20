package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"translate-server/middleware"
	"translate-server/services"
)

type TranslateController struct {
	Ctx               iris.Context
	TranslateService  services.TranslateService
}

func (t *TranslateController) BeforeActivation(b mvc.BeforeActivation) {
	b.Router().Use(middleware.CheckLoginMiddleware,
		middleware.CheckActivationMiddleware,
		middleware.IsSystemAvailable,
		middleware.FileLimiterMiddleware)
	b.Handle("GET", "/lang", "GetLangList")              // 获取支持的语言列表
	b.Handle("GET", "/allLang", "GetAllLangList")              // 获取支持的语言列表
	b.Handle("GET", "/records", "GetAllRecords")         // 获取所有的翻译记录
	b.Handle("GET", "/records/{type: uint64}/{offset: uint64}/{count: uint64}", "GetRecordsByType")         // 获取所有的翻译记录
	b.Handle("POST", "/upload", "PostUpload")            // 文件上传
	b.Handle("POST", "/content", "PostTranslateContent") // 文本翻译
	b.Handle("POST", "/file", "PostTranslateFile")       // 执行文件翻译
	b.Handle("POST", "/delete", "PostDeleteRecord")      // 删除某一条记录
	b.Handle("POST", "/down", "PostDownFile")            // 下载文件
}

// GetLangList 获取支持的语言
func (t *TranslateController) GetLangList() mvc.Result {
	return t.TranslateService.GetLangList()
}

// GetAllLangList 获取支持的语言
func (t *TranslateController) GetAllLangList() mvc.Result {
	return t.TranslateService.GetAllLangList()
}

// PostTranslateFile 解析文件 单个文件的解析
func (t *TranslateController) PostTranslateFile() mvc.Result {
	return t.TranslateService.PostTranslateFile(t.Ctx)
}

// PostTranslateContent 解析文本
func (t *TranslateController) PostTranslateContent() mvc.Result {
	return t.TranslateService.PostTranslateContent(t.Ctx)
}

// PostUpload 文件上传
func (t *TranslateController) PostUpload() mvc.Result {
	return t.TranslateService.PostUpload(t.Ctx)
}

// PostDownFile 下载文件
func (t *TranslateController) PostDownFile() {
	t.TranslateService.PostDownFile(t.Ctx)
}

// GetAllRecords 获取所有翻译记录
func (t *TranslateController) GetAllRecords() mvc.Result {
	return t.TranslateService.GetAllRecords(t.Ctx)
}

//GetRecordsByType 根据类型获取翻译记录
func (t *TranslateController) GetRecordsByType() mvc.Result {
	return t.TranslateService.GetRecordsByType(t.Ctx)
}
//PostDeleteRecord 删除翻译记录
func (t *TranslateController) PostDeleteRecord() mvc.Result {
	return t.TranslateService.PostDeleteRecord(t.Ctx)
}
