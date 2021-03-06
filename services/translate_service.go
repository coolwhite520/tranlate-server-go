package services

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
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
	sort.Sort(file.SupportLangList)
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
	if b, list := datamodels.NewActivationModel().IsSupportLang(req.DesLang, req.SrcLang); !b {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpLanguageNotSupport,
				"msg":  fmt.Sprintf("?????????????????????????????????????????????????????????%v", list),
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
	// ??????????????????
	content := strings.Trim(req.Content, " ")
	if len(content) == 0 {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpJsonParseError,
				"msg":  "?????????????????????",
			},
		}
	}
	if b, list := datamodels.NewActivationModel().IsSupportLang(req.DesLang, req.SrcLang); !b {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpLanguageNotSupport,
				"msg":  fmt.Sprintf("?????????????????????????????????????????????????????????%v", list),
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
	var records []structs.Record
	u := ctx.Values().Get("User")
	user, _ := (u).(structs.User)
	// ????????????????????????
	nowUnixMicro := time.Now().UnixMicro()
	DirRandId := fmt.Sprintf("%d", nowUnixMicro)
	userUploadDir := fmt.Sprintf("%s/%d/%s", structs.UploadDir, user.Id, DirRandId)
	if !utils.PathExists(userUploadDir) {
		err := os.MkdirAll(userUploadDir, os.ModePerm)
		if err != nil {
			return mvc.Response{
				Object: map[string]interface{}{
					"code": constant.HttpUploadFileError,
					"msg":  err.Error(),
				},
			}
		}
	}
	files, _, err := ctx.UploadFormFiles(userUploadDir)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpUploadFileError,
				"msg":  err.Error(),
			},
		}
	}
	for _, v := range files {
		filePath := fmt.Sprintf("%s/%s", userUploadDir, v.Filename)
		if !utils.PathExists(filePath) {
			return mvc.Response{
				Object: map[string]interface{}{
					"code": constant.HttpSuccess,
					"msg":  constant.HttpSuccess.String(),
					"data": records,
				},
			}
		}
		contentType, _ := utils.GetFileContentType(filePath)
		fileExt := filepath.Ext(v.Filename)
		fileName := v.Filename[0 : len(v.Filename)-len(fileExt)]
		if strings.Contains(fileName, " ") {
			fileName = strings.ReplaceAll(fileName, " ", "_")
			newFileName := path.Join(userUploadDir, fileName + fileExt)
			err := os.Rename(filePath, newFileName)
			if err != nil {
				return mvc.Response{
					Object: map[string]interface{}{
						"code": constant.HttpUploadFileError,
						"msg":  "????????????????????????" + err.Error(),
					},
				}
			}
		}
		var TransType int
		if strings.Contains(contentType, "image/") {
			TransType = 1
		} else {
			TransType = 2
		}
		record := structs.Record{
			ContentType:   contentType,
			TransType:     TransType,
			FileName:      fileName,
			FileExt:       fileExt,
			DirRandId:     DirRandId,
			State:         structs.TransNoRun,
			StateDescribe: structs.TransNoRun.String(),
			UserId:        user.Id,
		}
		err = datamodels.InsertRecord(&record)
		if err != nil {
			continue
		}
		records = append(records, record)
	}
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
			"data": records,
		},
	}
}

// PostDownFile ????????????
func (t *translateService) PostDownFile(ctx iris.Context) {
	a := ctx.Values().Get("User")
	user, _ := (a).(structs.User)
	var req struct {
		Id   int64 `json:"id"`   // recordId
		Type int   `json:"type"` // ????????? 0: ???????????????1???????????????????????????2?????????????????????
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
	// ???????????????????????????????????????
	var record *structs.Record
	if user.IsSuper {
		// ?????????????????????????????????????????????
		record, err = datamodels.QueryTranslateRecordById(req.Id)
	} else {
		// ??????????????????????????????ID???????????????
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
				"msg":  "???????????????????????????",
			})
		return
	}
	//
	srcDir := fmt.Sprintf("%s/%d/%s", structs.UploadDir, record.UserId, record.DirRandId)
	translatedDir := fmt.Sprintf("%s/%d/%s", structs.OutputDir, record.UserId, record.DirRandId)

	var filePathName string
	if req.Type == 0 {
		filePathName = fmt.Sprintf("%s/%s%s", srcDir, record.FileName, record.FileExt)
	} else if req.Type == 2 {
		filePathName = fmt.Sprintf("%s/%s%s", translatedDir, record.FileName, record.OutFileExt)
	} else {
		ctx.JSON(
			map[string]interface{}{
				"code": constant.HttpFileNotFoundError,
				"msg":  "???????????????",
			})
		return
	}
	if !utils.PathExists(filePathName) {
		log.Errorln(fmt.Sprintf("??????????????????%s", filePathName))
		ctx.JSON(
			map[string]interface{}{
				"code": constant.HttpFileNotFoundError,
				"msg":  "???????????????",
			},
		)
		return
	}
	bytes, err := ioutil.ReadFile(filePathName)
	if err != nil {
		log.Errorln(fmt.Sprintf("??????????????????%s", err.Error()))
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
				"list":  records,
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
	var byId *structs.Record
	if user.IsSuper {
		byId, err = datamodels.QueryTranslateRecordById(req.RecordId)
		if err != nil {
			return mvc.Response{
				Object: map[string]interface{}{
					"code": constant.HttpMysqlQueryError,
					"msg":  constant.HttpMysqlQueryError.String(),
				},
			}
		}
	} else {
		byId, err = datamodels.QueryTranslateRecordByIdAndUserId(req.RecordId, user.Id)
		if err != nil {
			return mvc.Response{
				Object: map[string]interface{}{
					"code": constant.HttpMysqlQueryError,
					"msg":  constant.HttpMysqlQueryError.String(),
				},
			}
		}
	}
	if byId == nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": constant.HttpMysqlQueryError,
				"msg":  constant.HttpMysqlQueryError.String(),
			},
		}
	}

	if byId.TransType != 0 {
		srcDir := fmt.Sprintf("%s/%d/%s", structs.UploadDir, byId.UserId, byId.DirRandId)
		translatedDir := fmt.Sprintf("%s/%d/%s", structs.OutputDir, byId.UserId, byId.DirRandId)
		srcFilePathName := path.Join(srcDir, byId.FileName+byId.FileExt)
		desFilePathName := path.Join(translatedDir, byId.FileName+byId.FileExt)
		if utils.PathExists(srcFilePathName) {
			os.Remove(srcFilePathName)
		}
		if utils.PathExists(desFilePathName) {
			os.Remove(desFilePathName)
		}
	}

	if user.IsSuper {
		datamodels.DeleteTranslateRecordById(req.RecordId)
	} else {
		err = datamodels.DeleteTranslateRecordByIdAndUserId(req.RecordId, user.Id)
	}

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
