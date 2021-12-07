package services

import (
	"fmt"
	"github.com/kataras/iris/v12"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
	"time"
	"translate-server/datamodels"
	"translate-server/utils"
)

const UploadDir = "./uploads"
const OutputDir = "./output"

type TranslateService interface {
	ReceiveFiles(Ctx iris.Context) ([]datamodels.Record, error)
	TranslateContent(srcLang, desLang, content string) (string, error)
	TranslateFile(srcLang, desLang, filePath string)
	DeleteTranslateRecordById(int64, bool) error
	QueryTranslateRecordById(int64) ([]datamodels.Record, error)
}

func NewTranslateService() TranslateService {
	return &translateService{}
}

type translateService struct {
}

func (t *translateService) ReceiveFiles(Ctx iris.Context) ([]datamodels.Record,error) {
	var records []datamodels.Record
	u := Ctx.Values().Get("User")
	user, _ := (u).(datamodels.User)
	// 创建用户的子目录
	nowUnixMicro := time.Now().UnixMicro()
	userUploadDir := fmt.Sprintf("%s/%d/%d", UploadDir, user.ID, nowUnixMicro)
	if !utils.PathExists(userUploadDir) {
		os.MkdirAll(userUploadDir, 0777)
	}
	userOutputDir := fmt.Sprintf("%s/%d/%d", OutputDir, user.ID, nowUnixMicro)

	files, _, err := Ctx.UploadFormFiles(userUploadDir)
	if err != nil {
		return records, err
	}

	srcLang := Ctx.FormValue("src_lang")
	desLang := Ctx.FormValue("des_lang")

	for _, v := range files {
		filePath := fmt.Sprintf("%s/%s", userUploadDir, v.Filename)
		outputFilePath := fmt.Sprintf("%s/%s.txt", userOutputDir, v.Filename)
		contentType, _ := utils.GetFileContentType(filePath)
		md5, _ := utils.GetFileMd5(filePath)
		record := datamodels.Record{
			ContentType:    contentType,
			Md5:            md5,
			Content:        "",
			OutputContent:  "",
			SrcLang:        srcLang,
			DesLang:        desLang,
			FilePath:       filePath,
			OutputFilePath: outputFilePath,
			CreateAt:       time.Now().Format("2006-01-02 15:04:05"),
			State:          datamodels.TransNoRun,
			StateDescribe:  "",
			UserId:         user.ID,
		}
		//err := t.insertTranslateRecord(record)
		//if err == nil {
		//	records = append(records, record)
		//}
		records = append(records, record)
	}
	return records, nil
}

// TranslateContent 同步翻译，用户界面卡住，直接返回翻译结果
func (t *translateService) TranslateContent(srcLang, desLang, content string) (string, error) {

	return "", nil
}

// TranslateFile 异步翻译，将结果写入到数据库中
func (t *translateService) TranslateFile(srcLang, desLang, filePath string)  {
	content, err := t.extractContent(filePath)
	if err != nil {
		return
	}
	_, err = t.TranslateContent(srcLang, desLang, content)
	if err != nil {
		return
	}
}

func (t *translateService) DeleteTranslateRecordById(id int64, bDelFile bool) error {
	return nil
}
func (t *translateService) QueryTranslateRecordById(int64) ([]datamodels.Record, error) {
	return nil, nil
}

func (t *translateService) insertTranslateRecord(record datamodels.Record) error {
	tx, _ := db.Begin()
	sql := fmt.Sprintf("INSERT INTO tbl_user('Username', 'HashedPassword', 'IsSuper') VALUES(?,?,?);")
	stmt, err := tx.Prepare(sql)
	if err != nil {
		log.Error(err)
		return err
	}
	_, err = stmt.Exec()
	tx.Commit()
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (t translateService) ocrDetectedImage(filePath string) (string, error) {
	return "", nil
}

func (t translateService) tikaDetectedText(filePath string) (string, error) {
	return "", nil
}

func (t *translateService) extractContent(filePath string) (string, error) {
	contentType, err := utils.GetFileContentType(filePath)
	if err != nil {
		return "", err
	}
	if strings.Contains(contentType, "image/") {
		return t.ocrDetectedImage(filePath)
	} else {
		return t.tikaDetectedText(filePath)
	}
}
