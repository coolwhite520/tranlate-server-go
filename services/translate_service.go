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
const OutputDir = "./outputs"

type TranslateService interface {
	ReceiveFiles(Ctx iris.Context) ([]datamodels.Record, error)
	TranslateContent(srcLang string, desLang string, content string, userId int64) (string, error)
	TranslateFile(srcLang string, desLang string, recordId int64, userId int64)
	DeleteTranslateRecordById(int64, bool) error
	QueryTranslateRecordById(int64) (*datamodels.Record, error)
	QueryTranslateRecordsById(int64) ([]datamodels.Record, error)
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
	userUploadDir := fmt.Sprintf("%s/%d/%d", UploadDir, user.Id, nowUnixMicro)
	if !utils.PathExists(userUploadDir) {
		os.MkdirAll(userUploadDir, 0777)
	}
	userOutputDir := fmt.Sprintf("%s/%d/%d", OutputDir, user.Id, nowUnixMicro)

	files, _, err := Ctx.UploadFormFiles(userUploadDir)
	if err != nil {
		return records, err
	}
	for _, v := range files {
		filePath := fmt.Sprintf("%s/%s", userUploadDir, v.Filename)
		contentType, _ := utils.GetFileContentType(filePath)
		md5, _ := utils.GetFileMd5(filePath)
		record := datamodels.Record{
			ContentType:   contentType,
			Md5:           md5,
			FileName:      v.Filename,
			FileSrcDir:    userUploadDir,
			FileDesDir:    userOutputDir,
			CreateAt:      time.Now().Format("2006-01-02 15:04:05"),
			State:         datamodels.TransNoRun,
			StateDescribe: datamodels.TransNoRun.String(),
			UserId:        user.Id,
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
func (t *translateService) TranslateContent(srcLang string, desLang string, content string, userId int64) (string, error) {
	outputContent, _ :=  t.translate(srcLang, desLang, content)
	var record datamodels.Record
	record.SrcLang = srcLang
	record.DesLang = desLang
	record.Content = content
	record.ContentType = ""
	record.State = datamodels.TransSuccess
	record.StateDescribe = datamodels.TransSuccess.String()
	record.Md5 = utils.Md5V(record.Content)
	record.OutputContent = outputContent
	record.UserId = userId
	// 记录到数据库中
	err := t.InsertTranslateRecord(&record)
	if err != nil {
		return outputContent, err
	}
	return "", nil
}

// TranslateFile 异步翻译，将结果写入到数据库中
func (t *translateService) TranslateFile(srcLang string, desLang string, recordId int64, userId int64)  {
	record, _ := t.QueryTranslateRecordById(recordId)
	content, err := t.extractContent(record.FileSrcDir)
	if err != nil {
		return
	}
	_, err = t.translate(srcLang, desLang, content)
	if err != nil {
		return
	}
	if !utils.PathExists(record.FileDesDir) {
		os.MkdirAll(record.FileDesDir, 0777)
	}

}

func (t *translateService) DeleteTranslateRecordById(id int64, bDelFile bool) error {
	return nil
}

func (t *translateService) QueryTranslateRecordById(int64) (*datamodels.Record, error) {
	return nil, nil
}

func (t *translateService) QueryTranslateRecordsById(int64) ([]datamodels.Record, error) {

	return nil, nil
}

func (t *translateService) UpdateTranslateRecord(record * datamodels.Record) error {
	return nil
}

func (t *translateService) InsertTranslateRecord(record *datamodels.Record) error {
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

func (t *translateService) translate(srcLang string, desLang string, content string) (string, error) {
	return "", nil
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
