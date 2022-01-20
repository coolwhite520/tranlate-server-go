package translate_models

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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
		translateDocxFile(srcLang, desLang, record)
	} else {
		translateCommonFile(srcLang, desLang, record)
	}
}