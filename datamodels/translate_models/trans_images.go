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

func translateImagesFile(srcLang string, desLang string, record *structs.Record) error {
	srcDir := fmt.Sprintf("%s/%d/%s", structs.UploadDir, record.UserId, record.DirRandId)
	extractDir := fmt.Sprintf("%s/%d/%s", structs.ExtractDir, record.UserId, record.DirRandId)
	translatedDir := fmt.Sprintf("%s/%d/%s", structs.OutputDir, record.UserId, record.DirRandId)
	srcFilePathName := fmt.Sprintf("%s/%s%s", srcDir, record.FileName, record.FileExt)

	content, err := ocrDetectedImage(srcFilePathName, srcLang)
	if err != nil {
		return err
	}



	content = strings.Trim(content, " ")
	// 抽取成功，但是是空数据，那么就退出了
	if len(content) == 0 {
		return err
	}
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

