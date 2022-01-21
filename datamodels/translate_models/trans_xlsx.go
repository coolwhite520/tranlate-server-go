package translate_models

import (
	"baliance.com/gooxml/spreadsheet"
	"fmt"
	"os"
	"strings"
	"translate-server/datamodels"
	"translate-server/structs"
	"translate-server/utils"
)

func translateXlsxFile(srcLang string, desLang string, record *structs.Record) {
	srcDir := fmt.Sprintf("%s/%d/%s", structs.UploadDir, record.UserId, record.DirRandId)
	translatedDir := fmt.Sprintf("%s/%d/%s", structs.OutputDir, record.UserId, record.DirRandId)
	srcFilePathName := fmt.Sprintf("%s/%s%s", srcDir, record.FileName, record.FileExt)
	record.SrcLang = srcLang
	record.DesLang = desLang
	record.State = structs.TransExtractSuccess
	record.StateDescribe = structs.TransExtractSuccess.String()
	datamodels.UpdateRecord(record)
	xlsx, err := spreadsheet.Open(srcFilePathName)
	if !utils.PathExists(translatedDir) {
		err := os.MkdirAll(translatedDir, os.ModePerm)
		if err != nil {
			return
		}
	}
	sheets := xlsx.Sheets()
	for _, s := range sheets {
		rows := s.Rows()
		for _, r := range rows {
			for _,c := range r.Cells(){
				if !c.IsEmpty() {
					content := c.GetString()
					content = strings.Trim(content, " ")
					transContent, _, _ := translate(srcLang, desLang, content)
					c.SetString(transContent)
				}
			}
		}
	}
	desFile := fmt.Sprintf("%s/%s%s", translatedDir, record.FileName, record.OutFileExt)
	err = xlsx.SaveToFile(desFile)
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

