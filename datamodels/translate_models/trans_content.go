package translate_models

import (
	"time"
	"translate-server/datamodels"
	"translate-server/structs"
)

// TranslateContent 同步翻译，用户界面卡住，直接返回翻译结果
func TranslateContent(srcLang string, desLang string, content string, userId int64) (string, error) {
	transContent, sha1, err := translate(srcLang, desLang, content)
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
	record.Progress = 100
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
