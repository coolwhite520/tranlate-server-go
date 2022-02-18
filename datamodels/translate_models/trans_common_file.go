package translate_models

import (
	"baliance.com/gooxml/document"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
	"translate-server/apis"
	"translate-server/datamodels"
	"translate-server/structs"
	"translate-server/utils"
)

func translateCommonFile(srcLang string, desLang string, record *structs.Record) error {
	srcDir := fmt.Sprintf("%s/%d/%s", structs.UploadDir, record.UserId, record.DirRandId)
	translatedDir := fmt.Sprintf("%s/%d/%s", structs.OutputDir, record.UserId, record.DirRandId)
	srcFilePathName := fmt.Sprintf("%s/%s%s", srcDir, record.FileName, record.FileExt)
	// 开始抽取数据
	record.SrcLang = srcLang
	record.DesLang = desLang
	record.State = structs.TransBeginExtract
	record.StateDescribe = structs.TransBeginExtract.String()
	datamodels.UpdateRecord(record)

	content, err := apis.TikaParseFile(srcFilePathName)
	if err != nil {
		return err
	}
	content = strings.Trim(content, " ")
	// 抽取成功，但是是空数据，那么就退出了
	if len(content) == 0 {
		return errors.New("content empty.")
	}
	tokenize, err := apis.PyTokenize(srcLang, content)
	if err != nil {
		return err
	}
	totalProgress := len(tokenize)
	currentProgress := 0
	percent := 0

	doc := document.New()
	paragraph := doc.AddParagraph()
	for _, v := range tokenize {
		transContent, _, _ := translate(srcLang, desLang, v)
		run := paragraph.AddRun()
		run.AddText(transContent)
		currentProgress++
		if percent != currentProgress * 100 /totalProgress{
			percent = currentProgress * 100 /totalProgress
			datamodels.UpdateRecordProgress(record.Id, percent)
		}
	}

	if !utils.PathExists(translatedDir) {
		err := os.MkdirAll(translatedDir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	desFile := fmt.Sprintf("%s/%s%s", translatedDir, record.FileName, record.OutFileExt)
	err = doc.SaveToFile(desFile)
	if err != nil {
		log.Errorln(err)
		return err
	}
	// 计算文件md5
	md5, err := utils.GetFileMd5(srcFilePathName)
	if err != nil {
		return err
	}
	// 拼接sha1字符串
	sha1 := utils.Sha1(fmt.Sprintf("%s&%s&%s", md5, srcLang, desLang))
	record.Sha1 = sha1
	datamodels.UpdateRecord(record)
	return nil
}


