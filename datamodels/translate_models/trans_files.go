package translate_models

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"time"
	"translate-server/apis"
	"translate-server/datamodels"
	"translate-server/structs"
	"translate-server/utils"
)

// TranslateFile 异步翻译，将结果写入到数据库中
func TranslateFile(srcLang string, desLang string, recordId int64, userId int64) {
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
		record.Progress = 100
		record.StartAt = time.Now().Format("2006-01-02 15:04:05")
		record.EndAt = time.Now().Format("2006-01-02 15:04:05")
		datamodels.UpdateRecord(record)
		return
	}
	// 更新初始状态
	record.State = structs.TransBeginTranslate
	record.StateDescribe = structs.TransBeginTranslate.String()
	record.SrcLang = srcLang
	record.DesLang = desLang
	record.Progress = 0
	record.StartAt = time.Now().Format("2006-01-02 15:04:05")
	datamodels.UpdateRecord(record)

	// 让file容器接管并处理
	desFile := fmt.Sprintf("%s/%s%s", translatedDir, record.FileName, record.OutFileExt)
	err = apis.PyTransFile(recordId, srcFilePathName, desFile, srcLang, desLang)

	// 更新结束时间
	record.EndAt = time.Now().Format("2006-01-02 15:04:05")
	datamodels.UpdateRecord(record)

	// 统一处理err
	if err != nil {
		log.Errorln(err)
		record.Error = err.Error()
		record.State = structs.TransTranslateFailed
		record.StateDescribe = structs.TransTranslateFailed.String()
		datamodels.UpdateRecord(record)
	} else {
		// 翻译成功的时候写入sha1，便于后续翻译的命中
		record.Sha1 = sha1
		record.State = structs.TransTranslateSuccess
		record.StateDescribe = structs.TransTranslateSuccess.String()
		record.Progress = 100
		datamodels.UpdateRecord(record)
	}

}
