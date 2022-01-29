package translate_models

import (
	"baliance.com/gooxml/document"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"strings"
	"translate-server/apis"
	"translate-server/constant"
	"translate-server/datamodels"
	"translate-server/structs"
	"translate-server/utils"
)

func translateCommonFile(srcLang string, desLang string, record *structs.Record) {
	srcDir := fmt.Sprintf("%s/%d/%s", structs.UploadDir, record.UserId, record.DirRandId)
	extractDir := fmt.Sprintf("%s/%d/%s", structs.ExtractDir, record.UserId, record.DirRandId)
	translatedDir := fmt.Sprintf("%s/%d/%s", structs.OutputDir, record.UserId, record.DirRandId)
	srcFilePathName := fmt.Sprintf("%s/%s%s", srcDir, record.FileName, record.FileExt)
	// 开始抽取数据
	record.SrcLang = srcLang
	record.DesLang = desLang
	record.State = structs.TransBeginExtract
	record.StateDescribe = structs.TransBeginExtract.String()
	err := datamodels.UpdateRecord(record)
	if err != nil {
		return
	}
	content, err := extractContent(record.TransType, srcFilePathName, srcLang)
	if err != nil {
		record.State = structs.TransExtractFailed
		record.StateDescribe = structs.TransExtractFailed.String()
		record.Error = err.Error()
		datamodels.UpdateRecord(record)
		return
	}
	content = strings.Trim(content, " ")
	// 抽取成功，但是是空数据，那么就退出了
	if len(content) == 0 {
		record.State = structs.TransExtractSuccessContentEmpty
		record.StateDescribe = structs.TransExtractSuccessContentEmpty.String()
		datamodels.UpdateRecord(record)
		return
	}
	// 更新状态
	record.State = structs.TransExtractSuccess
	record.StateDescribe = structs.TransExtractSuccess.String()
	datamodels.UpdateRecord(record)
	if !utils.PathExists(extractDir) {
		err := os.MkdirAll(extractDir, os.ModePerm)
		if err != nil {
			record.State = structs.TransExtractFailed
			record.StateDescribe = structs.TransExtractFailed.String()
			record.Error = err.Error()
			datamodels.UpdateRecord(record)
			return
		}
	}
	desFile := fmt.Sprintf("%s/%s%s", extractDir, record.FileName, record.OutFileExt)
	err = ioutil.WriteFile(desFile, []byte(content), 0666)
	if err != nil {
		record.State = structs.TransExtractFailed
		record.StateDescribe = structs.TransExtractFailed.String()
		record.Error = err.Error()
		datamodels.UpdateRecord(record)
		return
	}
	// 更新为开始翻译状态
	record.State = structs.TransBeginTranslate
	record.StateDescribe = structs.TransBeginTranslate.String()
	err = datamodels.UpdateRecord(record)
	if err != nil {
		return
	}
	transContent, sha1, err := translate(srcLang, desLang, content)
	if err != nil {
		record.State = structs.TransTranslateFailed
		record.StateDescribe = structs.TransTranslateFailed.String()
		record.Error = err.Error()
		err = datamodels.UpdateRecord(record)
		return
	}
	if !utils.PathExists(translatedDir) {
		err := os.MkdirAll(translatedDir, os.ModePerm)
		if err != nil {
			return
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
		return
	}

	if err != nil {
		return
	}
	record.Sha1 = sha1
	record.State = structs.TransTranslateSuccess
	record.StateDescribe = structs.TransTranslateSuccess.String()
	record.Error = ""
	err = datamodels.UpdateRecord(record)
	if err != nil {
		return
	}
}


func ocrDetectedImage(filePath string, srcLang string) (string, error) {
	var ocrType string
	for k, v := range constant.LanguageOcrList {
		if k == srcLang {
			ocrType = v
			break
		}
	}
	return apis.OcrParseFile(filePath, ocrType)
}

func tikaDetectedText(filePath string) (string, error) {
	return apis.TikaParseFile(filePath)
}

func extractContent(TransType int, filePath string, srcLang string) (string, error) {
	if TransType == 1 {
		return ocrDetectedImage(filePath, srcLang)
	} else {
		return tikaDetectedText(filePath)
	}
}
