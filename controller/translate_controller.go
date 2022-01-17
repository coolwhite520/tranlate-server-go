package controller

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"io/ioutil"
	"strings"
	"translate-server/datamodels"
	"translate-server/middleware"
	"translate-server/services"
	"translate-server/utils"
)

type TranslateController struct {
	Ctx               iris.Context
	TranslateService  services.TranslateService
	ActivationService services.ActivationService
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
	file, state := t.ActivationService.ParseKeystoreFile()
	if state != datamodels.HttpSuccess {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": state,
				"msg":  state.String(),
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": datamodels.HttpSuccess,
			"msg":  datamodels.HttpSuccess.String(),
			"data": file.SupportLangList,
		},
	}
}

// GetAllLangList 获取支持的语言
func (t *TranslateController) GetAllLangList() mvc.Result {
	return mvc.Response{
		Object: map[string]interface{}{
			"code": datamodels.HttpSuccess,
			"msg":  datamodels.HttpSuccess.String(),
			"data": datamodels.AllLanguageList,
		},
	}
}

// PostTranslateFile 解析文件 单个文件的解析
func (t *TranslateController) PostTranslateFile() mvc.Result {
	var req struct {
		SrcLang  string `json:"src_lang"`
		DesLang  string `json:"des_lang"`
		RecordId int64  `json:"record_id"`
	}
	err := t.Ctx.ReadJSON(&req)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": datamodels.HttpJsonParseError,
				"msg":  datamodels.HttpJsonParseError.String(),
			},
		}
	}
	if b, list :=  t.ActivationService.IsSupportLang(req.DesLang, req.SrcLang); !b {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": datamodels.HttpLanguageNotSupport,
				"msg":  fmt.Sprintf("不支持的语言，当前版本支持的语言列表为%v", list),
			},
		}
	}
	a := t.Ctx.Values().Get("User")
	user, _ := (a).(datamodels.User)
	go func() {
		t.TranslateService.TranslateFile(req.SrcLang, req.DesLang, req.RecordId, user.Id)
	}()
	return mvc.Response{
		Object: map[string]interface{}{
			"code": datamodels.HttpSuccess,
			"msg":  datamodels.HttpSuccess.String(),
		},
	}
}

// PostTranslateContent 解析文本
func (t *TranslateController) PostTranslateContent() mvc.Result {
	var req struct {
		SrcLang string `json:"src_lang"`
		DesLang string `json:"des_lang"`
		Content string `json:"content"`
	}
	err := t.Ctx.ReadJSON(&req)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": datamodels.HttpJsonParseError,
				"msg":  datamodels.HttpJsonParseError.String(),
			},
		}
	}

	// 判断是否为空
	content:= strings.Trim(req.Content, " ")
	if len(content) == 0 {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": datamodels.HttpJsonParseError,
				"msg":  "传递的内容为空",
			},
		}
	}
	if b, list := t.ActivationService.IsSupportLang(req.DesLang, req.SrcLang); !b {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": datamodels.HttpLanguageNotSupport,
				"msg":  fmt.Sprintf("不支持的语言，当前版本支持的语言列表为%v", list),
			},
		}
	}
	a := t.Ctx.Values().Get("User")
	user, _ := (a).(datamodels.User)
	outputContent, err := t.TranslateService.TranslateContent(req.SrcLang, req.DesLang, content, user.Id)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": datamodels.HttpTranslateError,
				"msg":  err.Error(),
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": datamodels.HttpSuccess,
			"msg":  datamodels.HttpSuccess.String(),
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
				"msg":  err.Error(),
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": datamodels.HttpSuccess,
			"msg":  datamodels.HttpSuccess.String(),
			"data": list,
		},
	}
}

// PostDownFile 下载文件
func (t *TranslateController) PostDownFile() {
	a := t.Ctx.Values().Get("User")
	user, _ := (a).(datamodels.User)
	var req struct {
		Id   int64 `json:"id"`   // recordId
		Type int   `json:"type"` // 分别： 0: 原始文件、1：抽取的内容文件、2：翻译结果文件
	}
	err := t.Ctx.ReadJSON(&req)
	if err != nil {
		t.Ctx.JSON(
			map[string]interface{}{
				"code": datamodels.HttpJsonParseError,
				"msg":  datamodels.HttpJsonParseError.String(),
			})
		return
	}
	// 判断这个文件是否属于这个人
	record, err := t.TranslateService.QueryTranslateRecordById(req.Id, user.Id)
	if err != nil {
		t.Ctx.JSON(
			map[string]interface{}{
				"code": datamodels.HttpRecordGetError,
				"msg":  datamodels.HttpRecordGetError.String(),
			})
		return
	}
	if record == nil {
		t.Ctx.JSON(
			map[string]interface{}{
				"code": datamodels.HttpRecordGetError,
				"msg":  "您访问的资源不存在",
			})
		return
	}
	// 通过反射获取FieldName对应的值
	srcDir := fmt.Sprintf("%s/%d/%s", services.UploadDir, user.Id, record.DirRandId)
	extractDir := fmt.Sprintf("%s/%d/%s", services.ExtractDir, user.Id, record.DirRandId)
	translatedDir := fmt.Sprintf("%s/%d/%s", services.OutputDir, user.Id, record.DirRandId)

	var filePathName string
	if req.Type == 0 {
		filePathName = fmt.Sprintf("%s/%s", srcDir, record.FileName)
	} else if req.Type == 1 {
		filePathName = fmt.Sprintf("%s/%s.txt", extractDir, record.FileName)
	} else if req.Type == 2 {
		filePathName = fmt.Sprintf("%s/%s.txt", translatedDir, record.FileName)
	} else {
		t.Ctx.JSON(
			map[string]interface{}{
				"code": datamodels.HttpFileNotFoundError,
				"msg":  "文件不存在",
			})
		return
	}
	if !utils.PathExists(filePathName) {
		t.Ctx.JSON(
			map[string]interface{}{
				"code": datamodels.HttpFileNotFoundError,
				"msg":  "文件不存在",
			},
		)
		return
	}
	bytes, err := ioutil.ReadFile(filePathName)
	if err != nil {
		t.Ctx.JSON(
			map[string]interface{}{
				"code": datamodels.HttpFileOpenError,
				"msg":  err.Error(),
			},
		)
		return
	}
	t.Ctx.ResponseWriter().Write(bytes)
}

func (t *TranslateController) GetAllRecords() mvc.Result {
	a := t.Ctx.Values().Get("User")
	user, _ := (a).(datamodels.User)
	records, err := t.TranslateService.QueryTranslateRecordsByUserId(user.Id)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": datamodels.HttpRecordGetError,
				"msg":  err.Error(),
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": datamodels.HttpSuccess,
			"msg":  datamodels.HttpSuccess.String(),
			"data": records,
		},
	}
}

func (t *TranslateController) GetRecordsByType() mvc.Result {
	transType := t.Ctx.Params().GetIntDefault("type", 0)
	offset := t.Ctx.Params().GetIntDefault("offset", 0)
	count := t.Ctx.Params().GetIntDefault("count", 0)
	a := t.Ctx.Values().Get("User")
	user, _ := (a).(datamodels.User)
	var total int
	var records []datamodels.Record
	var err error
	if transType == 3 {
		total, records, err = t.TranslateService.QueryTranslateFileRecordsByUserId(user.Id, offset, count)
	} else {
		total, records, err = t.TranslateService.QueryTranslateRecordsByUserIdAndType(user.Id, transType, offset, count)
	}
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": datamodels.HttpRecordGetError,
				"msg":  err.Error(),
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": datamodels.HttpSuccess,
			"msg":  datamodels.HttpSuccess.String(),
			"data": map[string]interface{}{
				"list": records,
				"total": total,
			},
		},
	}
}

func (t *TranslateController) PostDeleteRecord() mvc.Result {
	var req struct {
		RecordId int64 `json:"record_id"`
	}
	err := t.Ctx.ReadJSON(&req)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": datamodels.HttpJsonParseError,
				"msg":  datamodels.HttpJsonParseError.String(),
			},
		}
	}
	a := t.Ctx.Values().Get("User")
	user, _ := (a).(datamodels.User)
	err = t.TranslateService.DeleteTranslateRecordById(req.RecordId, user.Id, true)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": datamodels.HttpRecordDelError,
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
