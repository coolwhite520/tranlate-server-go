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
	"translate-server/rpc"
	"translate-server/utils"
)

const UploadDir = "./uploads"
const ExtractDir = "./extracts"
const OutputDir = "./outputs"

var RecordTableFieldList = []string{
	"Md5",
	"Content",
	"ContentType",
	"TransType",
	"OutputContent",
	"SrcLang",
	"DesLang",
	"FileName",
	"DirRandId",
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
	QueryTranslateRecordsByUserIdAndType(userId int64, transType int) ([]datamodels.Record, error)
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
	DirRandId := fmt.Sprintf("%d", nowUnixMicro)
	userUploadDir := fmt.Sprintf("%s/%d/%s", UploadDir, user.Id, DirRandId)
	if !utils.PathExists(userUploadDir) {
		err := os.MkdirAll(userUploadDir, 0777)
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
		var TransType int
		if strings.Contains(contentType, "image/") {
			TransType = 1
		} else {
			TransType = 2
		}
		md5, _ := utils.GetFileMd5(filePath)
		record := datamodels.Record{
			ContentType:   contentType,
			TransType:     TransType,
			Md5:           md5,
			FileName:      v.Filename,
			DirRandId:     DirRandId,
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
	record.TransType = 0
	record.State = datamodels.TransTranslateSuccess
	record.StateDescribe = datamodels.TransTranslateSuccess.String()
	record.Md5 = utils.Md5V(record.Content)
	record.OutputContent = outputContent
	record.UserId = userId
	// 记录到数据库中
	err := t.InsertRecord(&record)
	if err != nil {
		return outputContent, err
	}
	return outputContent, nil
}

// TranslateFile 异步翻译，将结果写入到数据库中
func (t *translateService) TranslateFile(srcLang string, desLang string, recordId int64, userId int64) {
	record, _ := t.QueryTranslateRecordById(recordId, userId)
	if  record == nil {
		log.Error("查询不到RecordId为", recordId, "的记录")
		return
	}
	srcDir := fmt.Sprintf("%s/%d/%s", UploadDir, userId, record.DirRandId)
	extractDir := fmt.Sprintf("%s/%d/%s", ExtractDir, userId, record.DirRandId)
	translatedDir := fmt.Sprintf("%s/%d/%s", OutputDir,userId, record.DirRandId)
	srcFilePathName := path.Join(srcDir, record.FileName)

	record.State = datamodels.TransBeginExtract
	record.StateDescribe = datamodels.TransBeginExtract.String()
	t.UpdateRecord(record)
	content, err := t.extractContent(record.TransType, srcFilePathName)
	if err != nil {
		record.State = datamodels.TransExtractFailed
		record.StateDescribe = datamodels.TransExtractFailed.String()
		record.Error = err.Error()
		t.UpdateRecord(record)
		return
	}
	record.State = datamodels.TransExtractSuccess
	record.StateDescribe = datamodels.TransExtractSuccess.String()
	t.UpdateRecord(record)
	if !utils.PathExists(extractDir) {
		os.MkdirAll(extractDir, 0777)
	}
	desFile := fmt.Sprintf("%s/%s.txt", extractDir, record.FileName)
	ioutil.WriteFile(desFile, []byte(content), 0777)

	record.State = datamodels.TransBeginTranslate
	record.StateDescribe = datamodels.TransBeginTranslate.String()
	t.UpdateRecord(record)
	transContent, err := t.translate(srcLang, desLang, content)
	if err != nil {
		record.State = datamodels.TransTranslateFailed
		record.StateDescribe = datamodels.TransTranslateFailed.String()
		record.Error = err.Error()
		t.UpdateRecord(record)
		return
	}
	if !utils.PathExists(translatedDir) {
		os.MkdirAll(translatedDir, 0777)
	}
	desFile = fmt.Sprintf("%s/%s.txt", translatedDir, record.FileName)
	ioutil.WriteFile(desFile, []byte(transContent), 0777)
	record.State = datamodels.TransTranslateSuccess
	record.StateDescribe = datamodels.TransTranslateSuccess.String()
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
		srcDir := fmt.Sprintf("%s/%d/%s", UploadDir, userId, byId.DirRandId)
		extractDir := fmt.Sprintf("%s/%d/%s", ExtractDir, userId, byId.DirRandId)
		translatedDir := fmt.Sprintf("%s/%d/%s", OutputDir, userId, byId.DirRandId)
		srcFilePathName := path.Join(srcDir, byId.FileName)
		middleFilePathName := path.Join(extractDir, byId.FileName)
		desFilePathName := path.Join(translatedDir, byId.FileName)
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
		&record.TransType,
		&record.OutputContent,
		&record.SrcLang,
		&record.DesLang,
		&record.FileName,
		&record.DirRandId,
		&record.State,
		&record.StateDescribe,
		&record.Error,
		&record.UserId,
		&tt)
	if err != nil {
		return nil, err
	}
	record.CreateAt = tt.Format("2006-01-02 15:04:05")
	return record, nil
}
func (t *translateService) QueryTranslateRecordsByUserIdAndType(userId int64, transType int) ([]datamodels.Record, error) {
	sql := fmt.Sprintf("SELECT * FROM tbl_record where UserId=? and TransType=? order by CreateAt DESC")
	rows, err := db.Query(sql, userId, transType)
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
			&record.TransType,
			&record.OutputContent,
			&record.SrcLang,
			&record.DesLang,
			&record.FileName,
			&record.DirRandId,
			&record.State,
			&record.StateDescribe,
			&record.Error,
			&record.UserId,
			&tt)
		if err != nil {
			return nil, err
		}
		record.CreateAt = tt.Format("2006-01-02 15:04:05")
		records = append(records, record)
	}
	return records, nil
}
func (t *translateService) QueryTranslateRecordsByUserId(userId int64) ([]datamodels.Record, error) {
	sql := fmt.Sprintf("SELECT * FROM tbl_record where UserId=? order by CreateAt DESC")
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
			&record.TransType,
			&record.OutputContent,
			&record.SrcLang,
			&record.DesLang,
			&record.FileName,
			&record.DirRandId,
			&record.State,
			&record.StateDescribe,
			&record.Error,
			&record.UserId,
			&tt)
		if err != nil {
			return nil, err
		}
		record.CreateAt = tt.Format("2006-01-02 15:04:05")
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
		record.TransType,
		record.OutputContent,
		record.SrcLang,
		record.DesLang,
		record.FileName,
		record.DirRandId,
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
		record.TransType,
		record.OutputContent,
		record.SrcLang,
		record.DesLang,
		record.FileName,
		record.DirRandId,
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
	return rpc.PyTranslate(srcLang, desLang, content)
}

func (t translateService) ocrDetectedImage(filePath string) (string, error) {
	return rpc.OcrParseFile(filePath)
}

func (t translateService) tikaDetectedText(filePath string) (string, error) {
	return rpc.TikaParseFile(filePath)
}

func (t *translateService) extractContent(TransType int, filePath string) (string, error) {
	if TransType == 1 {
		return t.ocrDetectedImage(filePath)
	} else {
		return t.tikaDetectedText(filePath)
	}
}
