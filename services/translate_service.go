package services

import (
	"fmt"
	"github.com/kataras/iris/v12"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
	"translate-server/datamodels"
	"translate-server/utils"
)

const UploadDir = "./uploads"
const OutputDir = "./outputs"

var RecordTableFieldList = []string{
	"Md5",
	"Content",
	"ContentType",
	"OutputContent",
	"SrcLang",
	"DesLang",
	"FileName",
	"FileSrcDir",
	"FileDesDir",
	"State",
	"StateDescribe",
	"Error",
	"UserId",
}

type TranslateService interface {
	ReceiveFiles(Ctx iris.Context) ([]datamodels.Record, error)
	TranslateContent(srcLang string, desLang string, content string, userId int64) (string, error)
	TranslateFile(srcLang string, desLang string, recordId int64, userId int64)
	DeleteTranslateRecordById(id int64, userId int64, bDel bool) error
	QueryTranslateRecordById(id int64, userId int64) (*datamodels.Record, error) // user自己只能看见自己的文件
	QueryTranslateRecordsByUserId(userId int64) ([]datamodels.Record, error)
}

func NewTranslateService() TranslateService {
	return &translateService{}
}

type translateService struct {
}

func (t *translateService) ReceiveFiles(Ctx iris.Context) ([]datamodels.Record, error) {
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
		err = t.InsertRecord(&record)
		if err != nil {
			continue
		}
		records = append(records, record)
	}
	return records, nil
}

// TranslateContent 同步翻译，用户界面卡住，直接返回翻译结果
func (t *translateService) TranslateContent(srcLang string, desLang string, content string, userId int64) (string, error) {
	outputContent, _ := t.translate(srcLang, desLang, content)
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
	err := t.InsertRecord(&record)
	if err != nil {
		return outputContent, err
	}
	return "", nil
}

// TranslateFile 异步翻译，将结果写入到数据库中
func (t *translateService) TranslateFile(srcLang string, desLang string, recordId int64, userId int64) {
	record, _ := t.QueryTranslateRecordById(recordId, userId)
	if  record == nil {
		log.Error("查询不到RecordId为", recordId, "的记录")
		return
	}
	srcFilePathName := path.Join(record.FileSrcDir, record.FileName)
	content, err := t.extractContent(srcFilePathName)
	if err != nil {
		record.State = datamodels.TransError
		record.StateDescribe = datamodels.TransError.String()
		record.Error = err.Error()
		t.UpdateRecord(record)
		return
	}
	transContent, err := t.translate(srcLang, desLang, content)
	if err != nil {
		record.State = datamodels.TransError
		record.StateDescribe = datamodels.TransError.String()
		record.Error = err.Error()
		t.UpdateRecord(record)
		return
	}
	if !utils.PathExists(record.FileDesDir) {
		os.MkdirAll(record.FileDesDir, 0777)
	}
	desFile := fmt.Sprintf("%s/%s.txt", record.FileDesDir, record.FileName)
	ioutil.WriteFile(desFile, []byte(transContent), 0777)
	record.State = datamodels.TransSuccess
	record.StateDescribe = datamodels.TransSuccess.String()
	record.Error = ""
	t.UpdateRecord(record)
}

func (t *translateService) DeleteTranslateRecordById(id int64, userId int64, bDelFile bool) error {
	tx, _ := db.Begin()
	byId, err2 := t.QueryTranslateRecordById(id, userId)
	if err2 != nil {
		return err2
	}

	if bDelFile && byId.ContentType != ""{
		srcFilePathName := path.Join(byId.FileSrcDir, byId.FileName)
		desFilePathName := path.Join(byId.FileDesDir, byId.FileName)
		os.Remove(srcFilePathName)
		os.Remove(desFilePathName)
	}

	sql := fmt.Sprintf("DELETE FROM tbl_record where Id=? and UserId=?")
	stmt, err := tx.Prepare(sql)
	if err != nil {
		log.Error(err)
		return err
	}
	_, err = stmt.Exec(id, userId)
	tx.Commit()
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (t *translateService) QueryTranslateRecordById(id int64, userId int64) (*datamodels.Record, error) {
	sql := fmt.Sprintf("SELECT * FROM tbl_record where Id=? and UserId=?;")
	row:= db.QueryRow(sql, id, userId)
	record := new(datamodels.Record)
	var tt time.Time
	err := row.Scan(
		&record.Id,
		&record.Md5,
		&record.Content,
		&record.ContentType,
		&record.OutputContent,
		&record.SrcLang,
		&record.DesLang,
		&record.FileName,
		&record.FileSrcDir,
		&record.FileDesDir,
		&record.State,
		&record.StateDescribe,
		&record.Error,
		&record.UserId,
		&tt)
	if err != nil {
		return nil, err
	}
	record.CreateAt = tt.Local().Format("2006-01-02 15:04:05")
	return record, nil
}

func (t *translateService) QueryTranslateRecordsByUserId(userId int64) ([]datamodels.Record, error) {
	sql := fmt.Sprintf("SELECT * FROM tbl_record where UserId=?")
	rows, err := db.Query(sql, userId)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	var records []datamodels.Record
	for rows.Next() {
		record := datamodels.Record{}
		var tt time.Time
		err = rows.Scan(
			&record.Id,
			&record.Md5,
			&record.Content,
			&record.ContentType,
			&record.OutputContent,
			&record.SrcLang,
			&record.DesLang,
			&record.FileName,
			&record.FileSrcDir,
			&record.FileDesDir,
			&record.State,
			&record.StateDescribe,
			&record.Error,
			&record.UserId,
			&tt)
		if err != nil {
			return nil, err
		}
		record.CreateAt = tt.Local().Format("2006-01-02 15:04:05")
		records = append(records, record)
	}
	return records, nil
}

func (t *translateService) UpdateRecord(record *datamodels.Record) error {
	var q []string
	for _, _ = range RecordTableFieldList {
		q = append(q, "?")
	}
	allFields := strings.Join(RecordTableFieldList, ",")
	allQs := strings.Join(q, ",")
	tx, _ := db.Begin()
	sql := fmt.Sprintf("REPLACE INTO tbl_record(Id,%s) VALUES(?,%s);", allFields, allQs)
	stmt, err := tx.Prepare(sql)
	if err != nil {
		log.Error(err)
		return err
	}
	_, err = stmt.Exec(
		record.Id,
		record.Md5,
		record.Content,
		record.ContentType,
		record.OutputContent,
		record.SrcLang,
		record.DesLang,
		record.FileName,
		record.FileSrcDir,
		record.FileDesDir,
		record.State,
		record.StateDescribe,
		record.Error,
		record.UserId,
	)
	tx.Commit()
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (t *translateService) InsertRecord(record *datamodels.Record) error {
	var q []string
	for _, _ = range RecordTableFieldList {
		q = append(q, "?")
	}
	allFields := strings.Join(RecordTableFieldList, ",")
	allQs := strings.Join(q, ",")
	tx, _ := db.Begin()
	sql := fmt.Sprintf("INSERT INTO tbl_record(%s) VALUES(%s);", allFields, allQs)
	stmt, err := tx.Prepare(sql)
	if err != nil {
		log.Error(err)
		return err
	}
	result, err := stmt.Exec(
		record.Md5,
		record.Content,
		record.ContentType,
		record.OutputContent,
		record.SrcLang,
		record.DesLang,
		record.FileName,
		record.FileSrcDir,
		record.FileDesDir,
		record.State,
		record.StateDescribe,
		record.Error,
		record.UserId,
	)
	tx.Commit()
	if err != nil {
		log.Error(err)
		return err
	}
	record.Id, _ = result.LastInsertId()
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
