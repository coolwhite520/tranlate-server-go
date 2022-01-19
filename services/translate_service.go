package services

import (
	"baliance.com/gooxml/document"
	"fmt"
	"github.com/kataras/iris/v12"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
	"translate-server/datamodels"
	"translate-server/rpc"
	"translate-server/utils"
)

const UploadDir = "./data/uploads"
const ExtractDir = "./data/extracts"
const OutputDir = "./data/outputs"

var RecordTableFieldList = []string{
	"Sha1",
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
	"OutFileExt",
}

type TranslateService interface {
	ReceiveFiles(Ctx iris.Context) ([]datamodels.Record, error)
	TranslateContent(srcLang string, desLang string, content string, userId int64) (string, error)
	TranslateFile(srcLang string, desLang string, recordId int64, userId int64)
	DeleteTranslateRecordById(id int64, userId int64, bDel bool) error
	QueryTranslateRecordById(id int64, userId int64) (*datamodels.Record, error) // user自己只能看见自己的文件
	QueryTranslateRecordsByUserId(userId int64) ([]datamodels.Record, error)
	QueryTranslateRecordsByUserIdAndType(userId int64, transType int, offset int, count int) (int, []datamodels.Record, error)
	QueryTranslateFileRecordsByUserId(userId int64, offset int, count int) (int, []datamodels.Record, error)
	// QueryTranslateRecords 只有管理员有这个权限
	QueryTranslateRecords(offset int, count int) (int, []datamodels.RecordEx, error)
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
		var TransType int
		var OutFileExt string
		if strings.Contains(contentType, "image/") {
			TransType = 1
			OutFileExt = ".docx"
		} else {
			TransType = 2
			OutFileExt = filepath.Ext(v.Filename)
		}

		record := datamodels.Record{
			ContentType:   contentType,
			TransType:     TransType,
			FileName:      v.Filename,
			DirRandId:     DirRandId,
			CreateAt:      time.Now().Format("2006-01-02 15:04:05"),
			State:         datamodels.TransNoRun,
			StateDescribe: datamodels.TransNoRun.String(),
			UserId:        user.Id,
			OutFileExt:    OutFileExt,
		}
		err = t.InsertRecord(&record)
		if err != nil {
			continue
		}
		records = append(records, record)
	}
	return records, nil
}

func (t *translateService) Translate(srcLang string, desLang string, content string) (string, string, error) {
	sha1 := utils.Sha1(fmt.Sprintf("%s&%s&%s", content, srcLang, desLang))
	records, err := t.queryTranslateRecordsBySha1(sha1)
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
func (t *translateService) TranslateContent(srcLang string, desLang string, content string, userId int64) (string, error) {
	transContent, sha1, err := t.Translate(srcLang, desLang, content)
	if err != nil {
		return "", err
	}
	var record datamodels.Record
	record.SrcLang = srcLang
	record.DesLang = desLang
	record.Content = content
	record.ContentType = ""
	record.TransType = 0
	record.UserId = userId
	record.Sha1 = sha1
	record.OutputContent = transContent
	record.State = datamodels.TransTranslateSuccess
	record.StateDescribe = datamodels.TransTranslateSuccess.String()
	// 记录到数据库中
	records, err := t.queryTranslateRecordsBySha1(sha1)
	if err != nil {
		return "", err
	}
	// 如果是自己之前的记录，那么更新一下时间就好
	for _, v := range records {
		if v.UserId == userId && v.TransType == 0 {
			record.Id = v.Id
			record.CreateAt = time.Now().Format("2006-01-02 15:04:05")
			t.UpdateRecord(&record)
			return record.OutputContent, nil
		}
	}
	t.InsertRecord(&record)
	return record.OutputContent, nil
}

func (t *translateService) translateDocxFile(srcLang string, desLang string, record *datamodels.Record)  {
	srcDir := fmt.Sprintf("%s/%d/%s", UploadDir, record.UserId, record.DirRandId)
	translatedDir := fmt.Sprintf("%s/%d/%s", OutputDir, record.UserId, record.DirRandId)
	srcFilePathName := path.Join(srcDir, record.FileName)
	// 开始抽取数据
	record.SrcLang = srcLang
	record.DesLang = desLang
	record.State = datamodels.TransBeginExtract
	record.StateDescribe = datamodels.TransBeginExtract.String()
	err := t.UpdateRecord(record)
	if err != nil {
		return
	}
	content, err := t.extractContent(record.TransType, srcFilePathName, srcLang)
	if err != nil {
		record.State = datamodels.TransExtractFailed
		record.StateDescribe = datamodels.TransExtractFailed.String()
		record.Error = err.Error()
		t.UpdateRecord(record)
		return
	}
	content = strings.Trim(content, " ")
	// 抽取成功，但是是空数据，那么就退出了
	if len(content) == 0 {
		record.State = datamodels.TransExtractSuccessContentEmpty
		record.StateDescribe = datamodels.TransExtractSuccessContentEmpty.String()
		t.UpdateRecord(record)
		return
	}

	// 更新状态
	record.State = datamodels.TransExtractSuccess
	record.StateDescribe = datamodels.TransExtractSuccess.String()
	t.UpdateRecord(record)

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
			transContent, _, _ := t.Translate(srcLang, desLang, content)
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
				transContent, _, _ := t.Translate(srcLang, desLang, content)
				run.AddText(transContent)
			}
		}
	}
	tables := doc.Tables()
	for _, tal := range tables {
		for _, r := range tal.Rows() {
			for _,c := range r.Cells() {
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
						transContent, _, _ := t.Translate(srcLang, desLang, content)
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
				transContent, _, _ := t.Translate(srcLang, desLang, content)
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
	desFile := fmt.Sprintf("%s/%s", translatedDir, record.FileName)
	err = doc.SaveToFile(desFile)
	if err != nil {
		return
	}
	record.Sha1 = ""
	record.State = datamodels.TransTranslateSuccess
	record.StateDescribe = datamodels.TransTranslateSuccess.String()
	record.Error = ""
	t.UpdateRecord(record)
}

func (t translateService) translateCommonFile(srcLang string, desLang string, record *datamodels.Record)  {
	srcDir := fmt.Sprintf("%s/%d/%s", UploadDir, record.UserId, record.DirRandId)
	extractDir := fmt.Sprintf("%s/%d/%s", ExtractDir, record.UserId, record.DirRandId)
	translatedDir := fmt.Sprintf("%s/%d/%s", OutputDir, record.UserId, record.DirRandId)
	srcFilePathName := path.Join(srcDir, record.FileName)
	// 开始抽取数据
	record.SrcLang = srcLang
	record.DesLang = desLang
	record.State = datamodels.TransBeginExtract
	record.StateDescribe = datamodels.TransBeginExtract.String()
	err := t.UpdateRecord(record)
	if err != nil {
		return
	}
	content, err := t.extractContent(record.TransType, srcFilePathName, srcLang)
	if err != nil {
		record.State = datamodels.TransExtractFailed
		record.StateDescribe = datamodels.TransExtractFailed.String()
		record.Error = err.Error()
		t.UpdateRecord(record)
		return
	}
	content = strings.Trim(content, " ")
	// 抽取成功，但是是空数据，那么就退出了
	if len(content) == 0 {
		record.State = datamodels.TransExtractSuccessContentEmpty
		record.StateDescribe = datamodels.TransExtractSuccessContentEmpty.String()
		t.UpdateRecord(record)
		return
	}
	// 更新状态
	record.State = datamodels.TransExtractSuccess
	record.StateDescribe = datamodels.TransExtractSuccess.String()
	t.UpdateRecord(record)
	if !utils.PathExists(extractDir) {
		err := os.MkdirAll(extractDir, os.ModePerm)
		if err != nil {
			record.State = datamodels.TransExtractFailed
			record.StateDescribe = datamodels.TransExtractFailed.String()
			record.Error = err.Error()
			t.UpdateRecord(record)
			return
		}
	}
	desFile := fmt.Sprintf("%s/%s%s", extractDir, record.FileName, record.OutFileExt)
	err = ioutil.WriteFile(desFile, []byte(content), 0666)
	if err != nil {
		record.State = datamodels.TransExtractFailed
		record.StateDescribe = datamodels.TransExtractFailed.String()
		record.Error = err.Error()
		t.UpdateRecord(record)
		return
	}
	// 更新为开始翻译状态
	record.State = datamodels.TransBeginTranslate
	record.StateDescribe = datamodels.TransBeginTranslate.String()
	err = t.UpdateRecord(record)
	if err != nil {
		return
	}
	transContent, sha1, err := t.Translate(srcLang, desLang, content)
	if err != nil {
		record.State = datamodels.TransTranslateFailed
		record.StateDescribe = datamodels.TransTranslateFailed.String()
		record.Error = err.Error()
		err = t.UpdateRecord(record)
		return
	}
	if !utils.PathExists(translatedDir) {
		err := os.MkdirAll(translatedDir, os.ModePerm)
		if err != nil {
			return
		}
	}
	desFile = fmt.Sprintf("%s/%s%s", translatedDir, record.FileName, record.OutFileExt)
	err = ioutil.WriteFile(desFile, []byte(transContent), 0666)
	if err != nil {
		return
	}
	record.Sha1 = sha1
	record.State = datamodels.TransTranslateSuccess
	record.StateDescribe = datamodels.TransTranslateSuccess.String()
	record.Error = ""
	err = t.UpdateRecord(record)
	if err != nil {
		return
	}
}

// TranslateFile 异步翻译，将结果写入到数据库中
func (t *translateService) TranslateFile(srcLang string, desLang string, recordId int64, userId int64) {
	record, _ := t.QueryTranslateRecordById(recordId, userId)
	if record == nil {
		log.Error("查询不到RecordId为", recordId, "的记录")
		return
	}
	srcDir := fmt.Sprintf("%s/%d/%s", UploadDir, record.UserId, record.DirRandId)
	translatedDir := fmt.Sprintf("%s/%d/%s", OutputDir, record.UserId, record.DirRandId)
	srcFilePathName := path.Join(srcDir, record.FileName)
	// 计算文件md5
	md5, err := utils.GetFileMd5(srcFilePathName)
	if err != nil {
		return
	}
	// 拼接sha1字符串
	sha1 := utils.Sha1(fmt.Sprintf("%s&%s&%s", md5, srcLang, desLang))
	records, err := t.queryTranslateRecordsBySha1(sha1)
	if err != nil {
		return
	}
	if !utils.PathExists(translatedDir) {
		err := os.MkdirAll(translatedDir, os.ModePerm)
		if err != nil {
			return
		}
	}
	for _, r := range records {
		fmt.Sprintf("%s/%d/%s/%s", OutputDir, r.UserId, r.DirRandId, r.FileName)
	}
	ext := filepath.Ext(record.FileName)
	if strings.ToLower(ext) == ".docx" {
		t.translateDocxFile(srcLang, desLang, record)
	} else {
		t.translateCommonFile(srcLang, desLang, record)
	}
}

func (t *translateService) DeleteTranslateRecordById(id int64, userId int64, bDelFile bool) error {
	tx, _ := db.Begin()
	byId, err2 := t.QueryTranslateRecordById(id, userId)
	if err2 != nil {
		return err2
	}

	if bDelFile && byId.ContentType != "" {
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

// queryTranslateRecordBySha1 根据sha1字符串查找数据
func (t *translateService) queryTranslateRecordsBySha1(sha1str string) ([]datamodels.Record, error) {
	sql := fmt.Sprintf("SELECT * FROM tbl_record where Sha1=?;")
	rows, err := db.Query(sql, sha1str)
	if err != nil {
		return nil, err
	}
	var records []datamodels.Record
	for rows.Next() {
		var record datamodels.Record
		var tt time.Time
		err := rows.Scan(
			&record.Id,
			&record.Sha1,
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
			&record.OutFileExt,
			&tt)
		if err != nil {
			return nil, err
		}
		record.CreateAt = tt.Local().Format("2006-01-02 15:04:05")
		records = append(records, record)
	}
	return records, nil
}
func (t *translateService) QueryTranslateRecordById(id int64, userId int64) (*datamodels.Record, error) {
	sql := fmt.Sprintf("SELECT * FROM tbl_record where Id=? and UserId=?;")
	row := db.QueryRow(sql, id, userId)
	record := new(datamodels.Record)
	var tt time.Time
	err := row.Scan(
		&record.Id,
		&record.Sha1,
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
		&record.OutFileExt,
		&tt)
	if err != nil {
		return nil, err
	}
	record.CreateAt = tt.Local().Format("2006-01-02 15:04:05")
	return record, nil
}
func (t *translateService) QueryTranslateRecordsByUserIdAndType(userId int64,
	transType int, offset int, count int) (int, []datamodels.Record, error) {
	sqlCount := fmt.Sprintf("SELECT count(1) FROM tbl_record where UserId=? and TransType=?")
	ret := db.QueryRow(sqlCount, userId, transType)
	var total int
	err := ret.Scan(&total)
	if err != nil {
		log.Error(err)
		return 0, nil, err
	}

	sql := fmt.Sprintf("SELECT * FROM tbl_record where UserId=? and TransType=? order by CreateAt DESC limit %d,%d", offset, count)
	rows, err := db.Query(sql, userId, transType)
	if err != nil {
		log.Error(err)
		return 0, nil, err
	}
	var records []datamodels.Record
	for rows.Next() {
		record := datamodels.Record{}
		var tt time.Time
		err = rows.Scan(
			&record.Id,
			&record.Sha1,
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
			&record.OutFileExt,
			&tt)
		if err != nil {
			return 0, nil, err
		}
		record.CreateAt = tt.Local().Format("2006-01-02 15:04:05")
		records = append(records, record)
	}
	return total, records, nil
}

func (t *translateService) QueryTranslateFileRecordsByUserId(userId int64, offset int, count int) (int, []datamodels.Record, error) {
	sqlCount := fmt.Sprintf("SELECT count(1) FROM tbl_record where UserId=? and TransType != 0")
	ret := db.QueryRow(sqlCount, userId)
	var total int
	err := ret.Scan(&total)
	if err != nil {
		log.Error(err)
		return 0, nil, err
	}

	sql := fmt.Sprintf("SELECT * FROM tbl_record where UserId=? and TransType != 0 order by CreateAt DESC limit %d,%d", offset, count)
	rows, err := db.Query(sql, userId)
	if err != nil {
		log.Error(err)
		return 0, nil, err
	}
	var records []datamodels.Record
	for rows.Next() {
		record := datamodels.Record{}
		var tt time.Time
		err = rows.Scan(
			&record.Id,
			&record.Sha1,
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
			&record.OutFileExt,
			&tt)
		if err != nil {
			return 0, nil, err
		}
		record.CreateAt = tt.Local().Format("2006-01-02 15:04:05")
		records = append(records, record)
	}
	return total, records, nil
}


func (t *translateService) QueryTranslateRecords(offset int, count int) (int, []datamodels.RecordEx, error) {
	sqlCount := fmt.Sprintf("SELECT count(1)  FROM tbl_record a INNER JOIN tbl_user b ON a.UserId = b.Id;")
	ret := db.QueryRow(sqlCount)
	var total int
	err := ret.Scan(&total)
	if err != nil {
		log.Error(err)
		return 0, nil, err
	}
	sql := fmt.Sprintf("SELECT a.*, b.Username as Username  FROM tbl_record a INNER JOIN tbl_user b ON a.UserId = b.Id order by CreateAt DESC limit %d,%d", offset, count)
	rows, err := db.Query(sql)
	if err != nil {
		log.Error(err)
		return 0, nil, err
	}
	var records []datamodels.RecordEx
	for rows.Next() {
		record := datamodels.RecordEx{}
		var tt time.Time
		err = rows.Scan(
			&record.Id,
			&record.Sha1,
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
			&record.OutFileExt,
			&tt,
			&record.UserName,
			)
		if err != nil {
			return 0, nil, err
		}
		record.CreateAt = tt.Local().Format("2006-01-02 15:04:05")

		records = append(records, record)
	}
	return total, records, nil
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
			&record.Sha1,
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
			&record.OutFileExt,
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
		record.Sha1,
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
		record.OutFileExt,
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
		record.Sha1,
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
		record.OutFileExt,
	)
	tx.Commit()
	if err != nil {
		log.Error(err)
		return err
	}
	record.Id, _ = result.LastInsertId()
	return nil
}

func (t translateService) ocrDetectedImage(filePath string, srcLang string) (string, error) {
	var ocrType string
	for k, v := range datamodels.LanguageOcrList {
		if k == srcLang {
			ocrType = v
			break
		}
	}
	return rpc.OcrParseFile(filePath, ocrType)
}

func (t translateService) tikaDetectedText(filePath string) (string, error) {
	return rpc.TikaParseFile(filePath)
}

func (t *translateService) extractContent(TransType int, filePath string, srcLang string) (string, error) {
	if TransType == 1 {
		return t.ocrDetectedImage(filePath, srcLang)
	} else {
		return t.tikaDetectedText(filePath)
	}
}
