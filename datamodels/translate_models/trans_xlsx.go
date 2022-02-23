package translate_models

import (
	"baliance.com/gooxml/spreadsheet"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"translate-server/apis"
	"translate-server/datamodels"
	"translate-server/structs"
	"translate-server/utils"
)

func calculateXlsxTotalProgress(srcFilePathName string) (int, error) {
	sum := 0
	xlsx, err := spreadsheet.Open(srcFilePathName)
	if err != nil {
		return 0, err
	}
	sheets := xlsx.Sheets()
	for _, s := range sheets {
		rows := s.Rows()
		for _, r := range rows {
			for _,c := range r.Cells(){
				if !c.IsEmpty() {
					content := c.GetString()
					content = strings.Trim(content, " ")
					if len(content) > 0 {
						sum ++
					}
				}
			}
		}
	}
	return sum, nil
}

func translateXlsxFile(srcLang string, desLang string, record *structs.Record) error {
	srcDir := fmt.Sprintf("%s/%d/%s", structs.UploadDir, record.UserId, record.DirRandId)
	translatedDir := fmt.Sprintf("%s/%d/%s", structs.OutputDir, record.UserId, record.DirRandId)
	srcFilePathName := fmt.Sprintf("%s/%s%s", srcDir, record.FileName, record.FileExt)
	ext := filepath.Ext(record.FileExt)
	if strings.ToLower(ext) == ".xls" {
		err := apis.PyConvertSpecialFile(srcFilePathName, "x2x")
		if err != nil {
			return err
		}
		srcFilePathName = srcFilePathName + "x"
	}
	totalProgress, err := calculateXlsxTotalProgress(srcFilePathName)
	if err != nil {
		return err
	}
	currentProgress := 0
	percent := 0

	xlsx, err := spreadsheet.Open(srcFilePathName)
	if !utils.PathExists(translatedDir) {
		err = os.MkdirAll(translatedDir, os.ModePerm)
		if err != nil {
			return err
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
					if len(content) > 0 {
						transContent, _, _ := translate(srcLang, desLang, content)
						c.SetString(transContent)
						currentProgress++
						if percent != currentProgress * 100 /totalProgress{
							percent = currentProgress * 100 /totalProgress
							datamodels.UpdateRecordProgress(record.Id, percent)
						}
					}
				}
			}
		}
	}
	desFile := fmt.Sprintf("%s/%s%s", translatedDir, record.FileName, record.OutFileExt)
	err = xlsx.SaveToFile(desFile)
	if err != nil {
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

