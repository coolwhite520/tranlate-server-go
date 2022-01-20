package datamodels

import (
	"baliance.com/gooxml/document"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
	"translate-server/rpc"
	"translate-server/structs"
	"translate-server/utils"
)

var transInstance *translateModel

func NewTranslateModel() *translateModel {
	once.Do(func() {
		transInstance = new(translateModel)
	})
	return transInstance
}

type translateModel struct {
}

func (t *translateModel) Translate(srcLang string, desLang string, content string) (string, string, error) {
	sha1 := utils.Sha1(fmt.Sprintf("%s&%s&%s", content, srcLang, desLang))
	records, err := QueryTranslateRecordsBySha1(sha1)
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
func (t *translateModel) TranslateContent(srcLang string, desLang string, content string, userId int64) (string, error) {
	transContent, sha1, err := t.Translate(srcLang, desLang, content)
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
	records, err := QueryTranslateRecordsBySha1(sha1)
	if err != nil {
		return "", err
	}
	// 如果是自己之前的记录，那么更新一下时间就好
	for _, v := range records {
		if v.UserId == userId && v.TransType == 0 {
			record.Id = v.Id
			record.CreateAt = time.Now().Format("2006-01-02 15:04:05")
			UpdateRecord(&record)
			return record.OutputContent, nil
		}
	}
	InsertRecord(&record)
	return record.OutputContent, nil
}

func (t *translateModel) translateDocxFile(srcLang string, desLang string, record *structs.Record) {
	srcDir := fmt.Sprintf("%s/%d/%s", structs.UploadDir, record.UserId, record.DirRandId)
	translatedDir := fmt.Sprintf("%s/%d/%s", structs.OutputDir, record.UserId, record.DirRandId)
	srcFilePathName := fmt.Sprintf("%s/%s%s", srcDir, record.FileName, record.FileExt)
	// 开始抽取数据
	record.SrcLang = srcLang
	record.DesLang = desLang
	record.State = structs.TransBeginExtract
	record.StateDescribe = structs.TransBeginExtract.String()
	err := UpdateRecord(record)
	if err != nil {
		return
	}
	content, err := t.extractContent(record.TransType, srcFilePathName, srcLang)
	if err != nil {
		record.State = structs.TransExtractFailed
		record.StateDescribe = structs.TransExtractFailed.String()
		record.Error = err.Error()
		UpdateRecord(record)
		return
	}
	content = strings.Trim(content, " ")
	// 抽取成功，但是是空数据，那么就退出了
	if len(content) == 0 {
		record.State = structs.TransExtractSuccessContentEmpty
		record.StateDescribe = structs.TransExtractSuccessContentEmpty.String()
		UpdateRecord(record)
		return
	}
	// 更新状态
	record.State = structs.TransExtractSuccess
	record.StateDescribe = structs.TransExtractSuccess.String()
	UpdateRecord(record)

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
	UpdateRecord(record)
}

func (t *translateModel) translateCommonFile(srcLang string, desLang string, record *structs.Record) {
	srcDir := fmt.Sprintf("%s/%d/%s", structs.UploadDir, record.UserId, record.DirRandId)
	extractDir := fmt.Sprintf("%s/%d/%s", structs.ExtractDir, record.UserId, record.DirRandId)
	translatedDir := fmt.Sprintf("%s/%d/%s", structs.OutputDir, record.UserId, record.DirRandId)

	srcFilePathName := fmt.Sprintf("%s/%s%s", srcDir, record.FileName, record.FileExt)
	// 开始抽取数据
	record.SrcLang = srcLang
	record.DesLang = desLang
	record.State = structs.TransBeginExtract
	record.StateDescribe = structs.TransBeginExtract.String()
	err := UpdateRecord(record)
	if err != nil {
		return
	}
	content, err := t.extractContent(record.TransType, srcFilePathName, srcLang)
	if err != nil {
		record.State = structs.TransExtractFailed
		record.StateDescribe = structs.TransExtractFailed.String()
		record.Error = err.Error()
		UpdateRecord(record)
		return
	}
	content = strings.Trim(content, " ")
	// 抽取成功，但是是空数据，那么就退出了
	if len(content) == 0 {
		record.State = structs.TransExtractSuccessContentEmpty
		record.StateDescribe = structs.TransExtractSuccessContentEmpty.String()
		UpdateRecord(record)
		return
	}
	// 更新状态
	record.State = structs.TransExtractSuccess
	record.StateDescribe = structs.TransExtractSuccess.String()
	UpdateRecord(record)
	if !utils.PathExists(extractDir) {
		err := os.MkdirAll(extractDir, os.ModePerm)
		if err != nil {
			record.State = structs.TransExtractFailed
			record.StateDescribe = structs.TransExtractFailed.String()
			record.Error = err.Error()
			UpdateRecord(record)
			return
		}
	}
	desFile := fmt.Sprintf("%s/%s%s", extractDir, record.FileName, record.OutFileExt)
	err = ioutil.WriteFile(desFile, []byte(content), 0666)
	if err != nil {
		record.State = structs.TransExtractFailed
		record.StateDescribe = structs.TransExtractFailed.String()
		record.Error = err.Error()
		UpdateRecord(record)
		return
	}
	// 更新为开始翻译状态
	record.State = structs.TransBeginTranslate
	record.StateDescribe = structs.TransBeginTranslate.String()
	err = UpdateRecord(record)
	if err != nil {
		return
	}
	transContent, sha1, err := t.Translate(srcLang, desLang, content)
	if err != nil {
		record.State = structs.TransTranslateFailed
		record.StateDescribe = structs.TransTranslateFailed.String()
		record.Error = err.Error()
		err = UpdateRecord(record)
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
	record.State = structs.TransTranslateSuccess
	record.StateDescribe = structs.TransTranslateSuccess.String()
	record.Error = ""
	err = UpdateRecord(record)
	if err != nil {
		return
	}
}

// TranslateFile 异步翻译，将结果写入到数据库中
func (t *translateModel) TranslateFile(srcLang string, desLang string, recordId int64, userId int64) {
	record, _ := QueryTranslateRecordByIdAndUserId(recordId, userId)
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
	records, err := QueryTranslateRecordsBySha1(sha1)
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
		record.Sha1 = sha1
		record.State = structs.TransTranslateSuccess
		record.StateDescribe = structs.TransTranslateSuccess.String()
		record.Error = ""
		err = UpdateRecord(record)
		if err != nil {
			log.Errorln(err)
			return
		}
		return
	}

	ext := filepath.Ext(record.FileExt)
	if strings.ToLower(ext) == ".docx" {
		t.translateDocxFile(srcLang, desLang, record)
	} else {
		t.translateCommonFile(srcLang, desLang, record)
	}
}

func (t *translateModel) DeleteTranslateRecordById(id int64, userId int64, bDelFile bool) error {
	byId, err2 := QueryTranslateRecordByIdAndUserId(id, userId)
	if err2 != nil {
		return err2
	}
	if bDelFile && byId.ContentType != "" {
		srcDir := fmt.Sprintf("%s/%d/%s", structs.UploadDir, userId, byId.DirRandId)
		extractDir := fmt.Sprintf("%s/%d/%s", structs.ExtractDir, userId, byId.DirRandId)
		translatedDir := fmt.Sprintf("%s/%d/%s", structs.OutputDir, userId, byId.DirRandId)
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
	return DeleteTranslateRecordById(id, userId)
}

func (t *translateModel) ocrDetectedImage(filePath string, srcLang string) (string, error) {
	var ocrType string
	for k, v := range structs.LanguageOcrList {
		if k == srcLang {
			ocrType = v
			break
		}
	}
	return rpc.OcrParseFile(filePath, ocrType)
}

func (t *translateModel) tikaDetectedText(filePath string) (string, error) {
	return rpc.TikaParseFile(filePath)
}

func (t *translateModel) extractContent(TransType int, filePath string, srcLang string) (string, error) {
	if TransType == 1 {
		return t.ocrDetectedImage(filePath, srcLang)
	} else {
		return t.tikaDetectedText(filePath)
	}
}
