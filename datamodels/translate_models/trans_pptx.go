package translate_models

import (
	"baliance.com/gooxml/presentation"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"translate-server/datamodels"
	"translate-server/structs"
	"translate-server/utils"
)

func translatePptxFile(srcLang string, desLang string, record *structs.Record) {
	srcDir := fmt.Sprintf("%s/%d/%s", structs.UploadDir, record.UserId, record.DirRandId)
	translatedDir := fmt.Sprintf("%s/%d/%s", structs.OutputDir, record.UserId, record.DirRandId)
	srcFilePathName := fmt.Sprintf("%s/%s%s", srcDir, record.FileName, record.FileExt)
	record.SrcLang = srcLang
	record.DesLang = desLang
	record.State = structs.TransExtractSuccess
	record.StateDescribe = structs.TransExtractSuccess.String()
	datamodels.UpdateRecord(record)
	ppt, err := presentation.Open(srcFilePathName)
	if err != nil {
		log.Errorln(err)
		return
	}

	slides := ppt.Slides()

	for _, l := range slides {
		holders := l.PlaceHolders()
		for _, h := range holders {
			paragraphs := h.Paragraphs()
			for _, p := range paragraphs {
				fmt.Println(p)
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
	err = ppt.SaveToFile(desFile)
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
	datamodels.UpdateRecord(record)
}
