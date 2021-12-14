package controller

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"translate-server/datamodels"
	"translate-server/middleware"
	"translate-server/services"
)

type TranslateController struct {
	Ctx iris.Context
	TranslateService services.TranslateService
	ActivationService services.ActivationService
}


func isIn(target string, strArray []datamodels.SupportLang) bool {
	for _, element := range strArray {
		if target == element.EnName {
			return true
		}
	}
	return false
}

func IsSupportLang(srcLang, desLang string) (bool, []datamodels.SupportLang) {
	newActivation := services.NewActivationService()
	file, state := newActivation.ParseKeystoreFile()
	if state != datamodels.HttpSuccess {
		return false, file.SupportLangList
	}
	if !isIn(srcLang, file.SupportLangList) {
		return false, file.SupportLangList
	}
	if !isIn(desLang, file.SupportLangList) {
		return false, file.SupportLangList
	}
	return true, file.SupportLangList
}


func (t *TranslateController) BeforeActivation(b mvc.BeforeActivation) {
	b.Router().Use(middleware.CheckActivationMiddleware,
		middleware.IsSystemAvailable,
		middleware.CheckLoginMiddleware,
		middleware.FileLimiterMiddleware)
	b.Handle("GET", "/lang", "GetLangList") // 获取支持的语言列表
	b.Handle("GET", "/records", "GetAllRecords") // 获取所有的翻译记录
	b.Handle("POST", "/upload", "PostUpload") // 文件上传
	b.Handle("POST", "/content", "PostTranslateContent") // 文本翻译
	b.Handle("POST", "/file", "PostTranslateFile") // 执行文件翻译
	b.Handle("POST", "/delete", "PostDeleteRecord") // 删除某一条记录
}

// GetLangList 获取支持的语言
func (t *TranslateController) GetLangList() mvc.Result {
	file, state := t.ActivationService.ParseKeystoreFile()
	if state != datamodels.HttpSuccess {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": state,
				"msg": state.String(),
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": datamodels.HttpSuccess,
			"msg": datamodels.HttpSuccess.String(),
			"data":file.SupportLangList,
		},
	}
}

// PostTranslateFile 解析文件 单个文件的解析
func (t *TranslateController) PostTranslateFile() mvc.Result {
	var req struct{
		SrcLang string `json:"src_lang"`
		DesLang string `json:"des_lang"`
		RecordId int64 `json:"record_id"`
	}
	err := t.Ctx.ReadJSON(&req)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": datamodels.HttpJsonParseError,
				"msg": datamodels.HttpJsonParseError.String(),
			},
		}
	}
	if b, list := IsSupportLang(req.DesLang, req.SrcLang); !b {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": datamodels.HttpLanguageNotSupport,
				"msg": fmt.Sprintf("不支持的语言，当前版本支持的语言列表为%v", list),
			},
		}
	}
	a:= t.Ctx.Values().Get("User")
	user, _ := (a).(datamodels.User)
	go func() {
		t.TranslateService.TranslateFile(req.SrcLang, req.DesLang, req.RecordId, user.Id)
	}()
	return mvc.Response{
		Object: map[string]interface{}{
			"code": datamodels.HttpSuccess,
			"msg": datamodels.HttpSuccess.String(),
		},
	}
}
// PostTranslateContent 解析文本
func (t *TranslateController) PostTranslateContent() mvc.Result {
	var req struct{
		SrcLang string `json:"src_lang"`
		DesLang string `json:"des_lang"`
		Content string `json:"content"`
	}
	err := t.Ctx.ReadJSON(&req)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": datamodels.HttpJsonParseError,
				"msg": datamodels.HttpJsonParseError.String(),
			},
		}
	}
	if b, list := IsSupportLang(req.DesLang, req.SrcLang); !b {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": datamodels.HttpLanguageNotSupport,
				"msg": fmt.Sprintf("不支持的语言，当前版本支持的语言列表为%v", list),
			},
		}
	}
	a:= t.Ctx.Values().Get("User")
	user, _ := (a).(datamodels.User)
	outputContent, err := t.TranslateService.TranslateContent(req.SrcLang, req.DesLang, req.Content, user.Id)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": datamodels.HttpTranslateError,
				"msg": err.Error(),
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": datamodels.HttpSuccess,
			"msg": datamodels.HttpSuccess.String(),
			"data": outputContent,
		},
	}
}

// PostUpload 文件上传
func (t *TranslateController) PostUpload() mvc.Result {
	list, err := t.TranslateService.ReceiveFiles(t.Ctx)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": datamodels.HttpUploadFileError,
				"msg": err.Error(),
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": datamodels.HttpSuccess,
			"msg": datamodels.HttpSuccess.String(),
			"data": list,
		},
	}
}

func (t TranslateController) GetAllRecords() mvc.Result {
	a:= t.Ctx.Values().Get("User")
	user, _ := (a).(datamodels.User)
	records, err := t.TranslateService.QueryTranslateRecordsByUserId(user.Id)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": datamodels.HttpRecordGetError,
				"msg": err.Error(),
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": datamodels.HttpSuccess,
			"msg": datamodels.HttpSuccess.String(),
			"data": records,
		},
	}
}

func (t *TranslateController) PostDeleteRecord() mvc.Result {
	var req struct{
		RecordId int64 `json:"record_id"`
	}
	err := t.Ctx.ReadJSON(&req)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": datamodels.HttpJsonParseError,
				"msg": datamodels.HttpJsonParseError.String(),
			},
		}
	}
	a:= t.Ctx.Values().Get("User")
	user, _ := (a).(datamodels.User)
	err = t.TranslateService.DeleteTranslateRecordById(req.RecordId, user.Id, true)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": datamodels.HttpRecordDelError,
				"msg": err.Error(),
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": datamodels.HttpSuccess,
			"msg": datamodels.HttpSuccess.String(),
		},
	}
}
