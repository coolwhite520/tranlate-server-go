package translate_models

import (
	"baliance.com/gooxml/document"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strings"
	"translate-server/apis"
	"translate-server/datamodels"
	"translate-server/structs"
	"translate-server/utils"
)

func translateCommonFile(srcLang string, desLang string, record *structs.Record) error {
	srcDir := fmt.Sprintf("%s/%d/%s", structs.UploadDir, record.UserId, record.DirRandId)
	extractDir := fmt.Sprintf("%s/%d/%s", structs.ExtractDir, record.UserId, record.DirRandId)
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
	// 更新状态
	record.State = structs.TransExtractSuccess
	record.StateDescribe = structs.TransExtractSuccess.String()
	datamodels.UpdateRecord(record)
	if !utils.PathExists(extractDir) {
		err := os.MkdirAll(extractDir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	desFile := fmt.Sprintf("%s/%s%s", extractDir, record.FileName, record.OutFileExt)
	err = ioutil.WriteFile(desFile, []byte(content), 0666)
	if err != nil {
		return err
	}
	// 更新为开始翻译状态
	record.State = structs.TransBeginTranslate
	record.StateDescribe = structs.TransBeginTranslate.String()
	datamodels.UpdateRecord(record)

	transContent, sha1, err := translate(srcLang, desLang, content)
	if err != nil {
		return err
	}
	if !utils.PathExists(translatedDir) {
		err := os.MkdirAll(translatedDir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	desFile = fmt.Sprintf("%s/%s%s", translatedDir, record.FileName, record.OutFileExt)
	doc := document.New()
	paragraph := doc.AddParagraph()
	run := paragraph.AddRun()
	run.AddText(transContent)
	err = doc.SaveToFile(desFile)
	if err != nil {
		log.Errorln(err)
		return err
	}

	if err != nil {
		return err
	}
	record.Sha1 = sha1
	datamodels.UpdateRecord(record)
	return nil
}


