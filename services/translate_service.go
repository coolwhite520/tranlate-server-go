package services

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"translate-server/constant"
	"translate-server/datamodels"
	"translate-server/datamodels/translate_models"
	"translate-server/structs"
	"translate-server/utils"
)

type TranslateService interface {
	GetLangList() mvc.Result
	GetAllLangList() mvc.Result
    PostTranslateFile(ctx iris.Context) mvc.Result
	PostTranslateContent(ctx iris.Context) mvc.Result
	PostUpload(ctx iris.Context) mvc.Result
	PostDownFile(ctx iris.Context)
	GetAllRecords(ctx iris.Context) mvc.Result
	GetRecordsByType(ctx iris.Context) mvc.Result
	PostDeleteRecord(ctx iris.Context) mvc.Result
}

func NewTranslateService() TranslateService {
	return &translateService{}
}

type translateService struct {
}

func (t *translateService) GetLangList() mvc.Result {
	file, state := datamodels.NewActivationModel().ParseKeystoreFile()
	if state != constant.HttpSuccess {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": state,
				"msg":  state.String(),
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": constant.HttpSuccess,
			"msg":  constant.HttpSuccess.String(),
			"data": file.SupportLangList,
		},
	}
}
func (t *translateService) GetAllLangList() mvc.Result {
	return mvc.Response{
		Object: map[string]interface{}{
			"code": constant.HttpSuccess,
			"msg":  constant.HttpSuccess.String(),
			"data": constant.AllLanguageList,
		},
	}
}


func (t *translateService) PostTranslateFile(ctx iris.Context) mvc.Result {
	var req struct {
		SrcLang  string `json:"src_lang"`
		DesLang  string `json:"des_lang"`
		RecordId int64  `json:"record_id"`
	}
	err := ctx.ReadJSON(&req)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpJsonParseError,
				"msg":  constant.HttpJsonParseError.String(),
			},
		}
	}
	if b, list :=  datamodels.NewActivationModel().IsSupportLang(req.DesLang, req.SrcLang); !b {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpLanguageNotSupport,
				"msg":  fmt.Sprintf("不支持的语言，当前版本支持的语言列表为%v", list),
			},
		}
	}
	a := ctx.Values().Get("User")
	user, _ := (a).(structs.User)
	go func() {
		translate_models.TranslateFile(req.SrcLang, req.DesLang, req.RecordId, user.Id)
	}()
	return mvc.Response{
		Object: map[string]interface{}{
			"code": constant.HttpSuccess,
			"msg":  constant.HttpSuccess.String(),
		},
	}
}
func (t *translateService) PostTranslateContent(ctx iris.Context) mvc.Result {
	var req struct {
		SrcLang string `json:"src_lang"`
		DesLang string `json:"des_lang"`
		Content string `json:"content"`
	}
	err := ctx.ReadJSON(&req)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpJsonParseError,
				"msg":  constant.HttpJsonParseError.String(),
			},
		}
	}
	// 判断是否为空
	content:= strings.Trim(req.Content, " ")
	if len(content) == 0 {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpJsonParseError,
				"msg":  "传递的内容为空",
			},
		}
	}
	if b, list := datamodels.NewActivationModel().IsSupportLang(req.DesLang, req.SrcLang); !b {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpLanguageNotSupport,
				"msg":  fmt.Sprintf("不支持的语言，当前版本支持的语言列表为%v", list),
			},
		}
	}
	a := ctx.Values().Get("User")
	user, _ := (a).(structs.User)
	outputContent, err := translate_models.TranslateContent(req.SrcLang, req.DesLang, content, user.Id)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpTranslateError,
				"msg":  err.Error(),
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": constant.HttpSuccess,
			"msg":  constant.HttpSuccess.String(),
			"data": outputContent,
		},
	}
}


func (t *translateService) PostUpload(ctx iris.Context) mvc.Result {
	list, err := translate_models.ReceiveFiles(ctx)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpUploadFileError,
				"msg":  err.Error(),
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": constant.HttpSuccess,
			"msg":  constant.HttpSuccess.String(),
			"data": list,
		},
	}
}

// PostDownFile 下载文件
func (t *translateService) PostDownFile(ctx iris.Context) {
	a := ctx.Values().Get("User")
	user, _ := (a).(structs.User)
	var req struct {
		Id   int64 `json:"id"`   // recordId
		Type int   `json:"type"` // 分别： 0: 原始文件、1：抽取的内容文件、2：翻译结果文件
	}
	err := ctx.ReadJSON(&req)
	if err != nil {
		ctx.JSON(
			map[string]interface{}{
				"code": constant.HttpJsonParseError,
				"msg":  constant.HttpJsonParseError.String(),
			})
		return
	}
	// 判断这个文件是否属于这个人
	var record *structs.Record
	if user.IsSuper {
		// 超级管理员可以查询任何人的记录
		record, err = datamodels.QueryTranslateRecordById(req.Id)
	} else {
		// 普通用户只能查询自己ID对应的记录
		record, err = datamodels.QueryTranslateRecordByIdAndUserId(req.Id, user.Id)
	}
	if err != nil {
		ctx.JSON(
			map[string]interface{}{
				"code": constant.HttpRecordGetError,
				"msg":  constant.HttpRecordGetError.String(),
			})
		return
	}
	if record == nil {
		ctx.JSON(
			map[string]interface{}{
				"code": constant.HttpRecordGetError,
				"msg":  "您访问的资源不存在",
			})
		return
	}
	//
	srcDir := fmt.Sprintf("%s/%d/%s", structs.UploadDir, user.Id, record.DirRandId)
	extractDir := fmt.Sprintf("%s/%d/%s", structs.ExtractDir, user.Id, record.DirRandId)
	translatedDir := fmt.Sprintf("%s/%d/%s", structs.OutputDir, user.Id, record.DirRandId)

	var filePathName string
	if req.Type == 0 {
		filePathName = fmt.Sprintf("%s/%s%s", srcDir, record.FileName, record.FileExt)
	} else if req.Type == 1 {
		filePathName = fmt.Sprintf("%s/%s%s", extractDir, record.FileName,record.OutFileExt)
	} else if req.Type == 2 {
		filePathName = fmt.Sprintf("%s/%s%s", translatedDir, record.FileName, record.OutFileExt)
	} else {
		ctx.JSON(
			map[string]interface{}{
				"code": constant.HttpFileNotFoundError,
				"msg":  "文件不存在",
			})
		return
	}
	if !utils.PathExists(filePathName) {
		ctx.JSON(
			map[string]interface{}{
				"code": constant.HttpFileNotFoundError,
				"msg":  "文件不存在",
			},
		)
		return
	}
	bytes, err := ioutil.ReadFile(filePathName)
	if err != nil {
		ctx.JSON(
			map[string]interface{}{
				"code": constant.HttpFileOpenError,
				"msg":  err.Error(),
			},
		)
		return
	}
	ctx.ResponseWriter().Write(bytes)
}


func (t *translateService) GetAllRecords(ctx iris.Context) mvc.Result {
	a := ctx.Values().Get("User")
	user, _ := (a).(structs.User)
	records, err := datamodels.QueryTranslateRecordsByUserId(user.Id)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpRecordGetError,
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

func (t *translateService) GetRecordsByType(ctx iris.Context) mvc.Result {
	transType := ctx.Params().GetIntDefault("type", 0)
	offset := ctx.Params().GetIntDefault("offset", 0)
	count := ctx.Params().GetIntDefault("count", 0)
	a := ctx.Values().Get("User")
	user, _ := (a).(structs.User)
	var total int
	var records []structs.Record
	var err error
	if transType == 3 {
		total, records, err = datamodels.QueryTranslateFileRecordsByUserId(user.Id, offset, count)
	} else {
		total, records, err = datamodels.QueryTranslateRecordsByUserIdAndType(user.Id, transType, offset, count)
	}
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpRecordGetError,
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

func (t *translateService) PostDeleteRecord(ctx iris.Context) mvc.Result {
	var req struct {
		RecordId int64 `json:"record_id"`
	}
	err := ctx.ReadJSON(&req)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpJsonParseError,
				"msg":  constant.HttpJsonParseError.String(),
			},
		}
	}
	a := ctx.Values().Get("User")
	user, _ := (a).(structs.User)
	byId, err2 := datamodels.QueryTranslateRecordByIdAndUserId(req.RecordId, user.Id)
	if err2 != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpMysqlQueryError,
				"msg":  constant.HttpMysqlQueryError.String(),
			},
		}
	}
	if byId.ContentType != "" {
		srcDir := fmt.Sprintf("%s/%d/%s", structs.UploadDir, user.Id, byId.DirRandId)
		extractDir := fmt.Sprintf("%s/%d/%s", structs.ExtractDir, user.Id, byId.DirRandId)
		translatedDir := fmt.Sprintf("%s/%d/%s", structs.OutputDir, user.Id, byId.DirRandId)
		srcFilePathName := path.Join(srcDir, byId.FileName+byId.FileExt)
		middleFilePathName := path.Join(extractDir, byId.FileName+byId.FileExt)
		desFilePathName := path.Join(translatedDir, byId.FileName+byId.FileExt)
		if utils.PathExists(srcFilePathName) {
			os.Remove(srcFilePathName)
		}
		if utils.PathExists(middleFilePathName) {
			os.Remove(middleFilePathName)
		}
		if utils.PathExists(desFilePathName) {
			os.Remove(desFilePathName)
		}
	}
	err = datamodels.DeleteTranslateRecordById(req.RecordId, user.Id)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpRecordDelError,
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

