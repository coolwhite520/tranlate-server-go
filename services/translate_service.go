package services

import (
	"baliance.com/gooxml/document"
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
	"translate-server/constant"
	"translate-server/datamodels"
	"translate-server/rpc"
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
		t.translateFile(req.SrcLang, req.DesLang, req.RecordId, user.Id)
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
	outputContent, err := t.translateContent(req.SrcLang, req.DesLang, content, user.Id)
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
	list, err := t.receiveFiles(ctx)
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

func (t *translateService) receiveFiles(Ctx iris.Context) ([]structs.Record, error) {
	var records []structs.Record
	u := Ctx.Values().Get("User")
	user, _ := (u).(structs.User)
	// 创建用户的子目录
	nowUnixMicro := time.Now().UnixMicro()
	DirRandId := fmt.Sprintf("%d", nowUnixMicro)
	userUploadDir := fmt.Sprintf("%s/%d/%s", structs.UploadDir, user.Id, DirRandId)
	if !utils.PathExists(userUploadDir) {
		err := os.MkdirAll(userUploadDir, os.ModePerm)
		if err != nil {
			return records, err
		}
	}
	files, _, err := Ctx.UploadFormFiles(userUploadDir)
	if err != nil {
		return records, err
	}
	for _, v := range files {
		filePath := fmt.Sprintf("%s/%s", userUploadDir, v.Filename)
		contentType, _ := utils.GetFileContentType(filePath)
		fileExt := filepath.Ext(v.Filename)
		fileName := v.Filename[0:len(v.Filename) - len(fileExt)]
		var TransType int
		var OutFileExt string
		if strings.Contains(contentType, "image/") {
			TransType = 1
			OutFileExt = ".docx"
		} else {
			TransType = 2
			OutFileExt = fileExt
		}

		record := structs.Record{
			ContentType:   contentType,
			TransType:     TransType,
			FileName:      fileName,
			FileExt:       fileExt,
			DirRandId:     DirRandId,
			CreateAt:      time.Now().Format("2006-01-02 15:04:05"),
			State:         structs.TransNoRun,
			StateDescribe: structs.TransNoRun.String(),
			UserId:        user.Id,
			OutFileExt:    OutFileExt,
		}
		err = datamodels.InsertRecord(&record)
		if err != nil {
			continue
		}
		records = append(records, record)
	}
	return records, nil
}


func (t *translateService) translate(srcLang string, desLang string, content string) (string, string, error) {
	sha1 := utils.Sha1(fmt.Sprintf("%s&%s&%s", content, srcLang, desLang))
	records, err := datamodels.QueryTranslateRecordsBySha1(sha1)
	if err != nil {
		return "", "", err
	}
	var transContent string
	for _, v := range records {
		if v.SrcLang == srcLang && v.DesLang == desLang {
			if v.TransType == 0 {
				transContent = v.OutputContent
				break
			}
		}
	}
	if len(transContent) == 0 {
		transContent, err = rpc.PyTranslate(srcLang, desLang, content)
		if err != nil {
			return "", "", err
		}
	}
	return transContent, sha1, nil
}

// TranslateContent 同步翻译，用户界面卡住，直接返回翻译结果
func (t *translateService) translateContent(srcLang string, desLang string, content string, userId int64) (string, error) {
	transContent, sha1, err := t.translate(srcLang, desLang, content)
	if err != nil {
		return "", err
	}
	var record structs.Record
	record.SrcLang = srcLang
	record.DesLang = desLang
	record.Content = content
	record.ContentType = ""
	record.TransType = 0
	record.UserId = userId
	record.Sha1 = sha1
	record.OutputContent = transContent
	record.State = structs.TransTranslateSuccess
	record.StateDescribe = structs.TransTranslateSuccess.String()
	// 记录到数据库中
	records, err := datamodels.QueryTranslateRecordsBySha1(sha1)
	if err != nil {
		return "", err
	}
	// 如果是自己之前的记录，那么更新一下时间就好
	for _, v := range records {
		if v.UserId == userId && v.TransType == 0 {
			record.Id = v.Id
			record.CreateAt = time.Now().Format("2006-01-02 15:04:05")
			datamodels.UpdateRecord(&record)
			return record.OutputContent, nil
		}
	}
	datamodels.InsertRecord(&record)
	return record.OutputContent, nil
}

func (t *translateService) translateDocxFile(srcLang string, desLang string, record *structs.Record) {
	srcDir := fmt.Sprintf("%s/%d/%s", structs.UploadDir, record.UserId, record.DirRandId)
	translatedDir := fmt.Sprintf("%s/%d/%s", structs.OutputDir, record.UserId, record.DirRandId)
	srcFilePathName := fmt.Sprintf("%s/%s%s", srcDir, record.FileName, record.FileExt)
	record.SrcLang = srcLang
	record.DesLang = desLang
	record.State = structs.TransExtractSuccess
	record.StateDescribe = structs.TransExtractSuccess.String()
	datamodels.UpdateRecord(record)

	doc, err := document.Open(srcFilePathName)
	if err != nil {
		log.Errorln(err)
		return
	}
	paragraphs := doc.Paragraphs()
	for _, p := range paragraphs {
		var content string
		for _, r := range p.Runs() {
			content += r.Text()
		}
		if len(strings.Trim(content, " ")) > 0 {
			for _, r := range p.Runs() {
				p.RemoveRun(r)
			}
			run := p.AddRun()
			transContent, _, _ := t.translate(srcLang, desLang, content)
			run.AddText(transContent)
		}
	}
	headers := doc.Headers()
	for _, h := range headers {
		for _, p := range h.Paragraphs() {
			var content string
			for _, r := range p.Runs() {
				content += r.Text()
			}
			if len(strings.Trim(content, " ")) > 0 {
				for _, r := range p.Runs() {
					p.RemoveRun(r)
				}
				run := p.AddRun()
				transContent, _, _ := t.translate(srcLang, desLang, content)
				run.AddText(transContent)
			}
		}
	}
	tables := doc.Tables()
	for _, tal := range tables {
		for _, r := range tal.Rows() {
			for _, c := range r.Cells() {
				for _, p := range c.Paragraphs() {
					var content string
					for _, r := range p.Runs() {
						content += r.Text()
					}
					if len(strings.Trim(content, " ")) > 0 {
						for _, r := range p.Runs() {
							p.RemoveRun(r)
						}
						run := p.AddRun()
						transContent, _, _ := t.translate(srcLang, desLang, content)
						run.AddText(transContent)
					}
				}

			}
		}
	}

	footers := doc.Footers()
	for _, f := range footers {
		for _, p := range f.Paragraphs() {
			var content string
			for _, r := range p.Runs() {
				content += r.Text()
			}
			if len(strings.Trim(content, " ")) > 0 {
				for _, r := range p.Runs() {
					p.RemoveRun(r)
				}
				run := p.AddRun()
				transContent, _, _ := t.translate(srcLang, desLang, content)
				run.AddText(transContent)
			}
		}

	}
	if !utils.PathExists(translatedDir) {
		err := os.MkdirAll(translatedDir, os.ModePerm)
		if err != nil {
			return
		}
	}
	desFile := fmt.Sprintf("%s/%s%s", translatedDir, record.FileName, record.OutFileExt)
	err = doc.SaveToFile(desFile)
	if err != nil {
		return
	}
	// 计算文件md5
	md5, err := utils.GetFileMd5(srcFilePathName)
	if err != nil {
		return
	}
	// 拼接sha1字符串
	sha1 := utils.Sha1(fmt.Sprintf("%s&%s&%s", md5, srcLang, desLang))
	record.Sha1 = sha1
	record.State = structs.TransTranslateSuccess
	record.StateDescribe = structs.TransTranslateSuccess.String()
	record.Error = ""
	datamodels.UpdateRecord(record)
}

func (t *translateService) translateCommonFile(srcLang string, desLang string, record *structs.Record) {
	srcDir := fmt.Sprintf("%s/%d/%s", structs.UploadDir, record.UserId, record.DirRandId)
	extractDir := fmt.Sprintf("%s/%d/%s", structs.ExtractDir, record.UserId, record.DirRandId)
	translatedDir := fmt.Sprintf("%s/%d/%s", structs.OutputDir, record.UserId, record.DirRandId)
	srcFilePathName := fmt.Sprintf("%s/%s%s", srcDir, record.FileName, record.FileExt)
	// 开始抽取数据
	record.SrcLang = srcLang
	record.DesLang = desLang
	record.State = structs.TransBeginExtract
	record.StateDescribe = structs.TransBeginExtract.String()
	err := datamodels.UpdateRecord(record)
	if err != nil {
		return
	}
	content, err := t.extractContent(record.TransType, srcFilePathName, srcLang)
	if err != nil {
		record.State = structs.TransExtractFailed
		record.StateDescribe = structs.TransExtractFailed.String()
		record.Error = err.Error()
		datamodels.UpdateRecord(record)
		return
	}
	content = strings.Trim(content, " ")
	// 抽取成功，但是是空数据，那么就退出了
	if len(content) == 0 {
		record.State = structs.TransExtractSuccessContentEmpty
		record.StateDescribe = structs.TransExtractSuccessContentEmpty.String()
		datamodels.UpdateRecord(record)
		return
	}
	// 更新状态
	record.State = structs.TransExtractSuccess
	record.StateDescribe = structs.TransExtractSuccess.String()
	datamodels.UpdateRecord(record)
	if !utils.PathExists(extractDir) {
		err := os.MkdirAll(extractDir, os.ModePerm)
		if err != nil {
			record.State = structs.TransExtractFailed
			record.StateDescribe = structs.TransExtractFailed.String()
			record.Error = err.Error()
			datamodels.UpdateRecord(record)
			return
		}
	}
	desFile := fmt.Sprintf("%s/%s%s", extractDir, record.FileName, record.OutFileExt)
	err = ioutil.WriteFile(desFile, []byte(content), 0666)
	if err != nil {
		record.State = structs.TransExtractFailed
		record.StateDescribe = structs.TransExtractFailed.String()
		record.Error = err.Error()
		datamodels.UpdateRecord(record)
		return
	}
	// 更新为开始翻译状态
	record.State = structs.TransBeginTranslate
	record.StateDescribe = structs.TransBeginTranslate.String()
	err = datamodels.UpdateRecord(record)
	if err != nil {
		return
	}
	transContent, sha1, err := t.translate(srcLang, desLang, content)
	if err != nil {
		record.State = structs.TransTranslateFailed
		record.StateDescribe = structs.TransTranslateFailed.String()
		record.Error = err.Error()
		err = datamodels.UpdateRecord(record)
		return
	}
	if !utils.PathExists(translatedDir) {
		err := os.MkdirAll(translatedDir, os.ModePerm)
		if err != nil {
			return
		}
	}
	desFile = fmt.Sprintf("%s/%s%s", translatedDir, record.FileName, record.OutFileExt)
	doc := document.New()
	paragraph := doc.AddParagraph()
	run := paragraph.AddRun()
	run.AddText(transContent)
	err = doc.SaveToFile(desFile)
	if err != nil {
		log.Errorln(err)
		return
	}

	if err != nil {
		return
	}
	record.Sha1 = sha1
	record.State = structs.TransTranslateSuccess
	record.StateDescribe = structs.TransTranslateSuccess.String()
	record.Error = ""
	err = datamodels.UpdateRecord(record)
	if err != nil {
		return
	}
}

// translateFile 异步翻译，将结果写入到数据库中
func (t *translateService) translateFile(srcLang string, desLang string, recordId int64, userId int64) {
	// 先查找是否存在相同的翻译结果
	record, _ := datamodels.QueryTranslateRecordByIdAndUserId(recordId, userId)
	if record == nil {
		log.Error("查询不到RecordId为", recordId, "的记录")
		return
	}
	srcDir := fmt.Sprintf("%s/%d/%s", structs.UploadDir, record.UserId, record.DirRandId)
	translatedDir := fmt.Sprintf("%s/%d/%s", structs.OutputDir, record.UserId, record.DirRandId)
	srcFilePathName := fmt.Sprintf("%s/%s%s", srcDir, record.FileName, record.FileExt)
	// 计算文件md5
	md5, err := utils.GetFileMd5(srcFilePathName)
	if err != nil {
		return
	}
	// 拼接sha1字符串
	sha1 := utils.Sha1(fmt.Sprintf("%s&%s&%s", md5, srcLang, desLang))
	records, err := datamodels.QueryTranslateRecordsBySha1(sha1)
	if err != nil {
		log.Errorln(err)
		return
	}
	if !utils.PathExists(translatedDir) {
		err := os.MkdirAll(translatedDir, os.ModePerm)
		if err != nil {
			log.Errorln(err)
			return
		}
	}
	for _, r := range records {
		srcFile := fmt.Sprintf("%s/%d/%s/%s%s", structs.OutputDir, r.UserId, r.DirRandId, r.FileName, r.OutFileExt)
		desFile := fmt.Sprintf("%s/%d/%s/%s%s", structs.OutputDir, record.UserId, record.DirRandId, record.FileName, record.OutFileExt)
		all, err := ioutil.ReadFile(srcFile)
		if err != nil {
			log.Errorln(err)
			return
		}
		ioutil.WriteFile(desFile, all, 0666)
		record.SrcLang = srcLang
		record.DesLang = desLang
		record.Sha1 = sha1
		record.State = structs.TransTranslateSuccess
		record.StateDescribe = structs.TransTranslateSuccess.String()
		record.Error = ""
		err = datamodels.UpdateRecord(record)
		if err != nil {
			log.Errorln(err)
			return
		}
		return
	}
	// 没有找到相同的文件和 srclang 、desLang的时候
	ext := filepath.Ext(record.FileExt)
	if strings.ToLower(ext) == ".docx" {
		t.translateDocxFile(srcLang, desLang, record)
	} else {
		t.translateCommonFile(srcLang, desLang, record)
	}
}

func (t *translateService) ocrDetectedImage(filePath string, srcLang string) (string, error) {
	var ocrType string
	for k, v := range constant.LanguageOcrList {
		if k == srcLang {
			ocrType = v
			break
		}
	}
	return rpc.OcrParseFile(filePath, ocrType)
}

func (t *translateService) tikaDetectedText(filePath string) (string, error) {
	return rpc.TikaParseFile(filePath)
}

func (t *translateService) extractContent(TransType int, filePath string, srcLang string) (string, error) {
	if TransType == 1 {
		return t.ocrDetectedImage(filePath, srcLang)
	} else {
		return t.tikaDetectedText(filePath)
	}
}