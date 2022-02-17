package translate_models

import (
	"fmt"
	"translate-server/apis"
	"translate-server/datamodels"
	"translate-server/structs"
	"translate-server/utils"
)

func translateEmlFile(srcLang string, desLang string, record *structs.Record) {
	srcDir := fmt.Sprintf("%s/%d/%s", structs.UploadDir, record.UserId, record.DirRandId)
	translatedDir := fmt.Sprintf("%s/%d/%s", structs.OutputDir, record.UserId, record.DirRandId)
	srcFilePathName := fmt.Sprintf("%s/%s%s", srcDir, record.FileName, record.FileExt)
	record.SrcLang = srcLang
	record.DesLang = desLang
	record.State = structs.TransExtractSuccess
	record.StateDescribe = structs.TransExtractSuccess.String()
	datamodels.UpdateRecord(record)
	desFile := fmt.Sprintf("%s/%s%s", translatedDir, record.FileName, record.OutFileExt)
	err := apis.PyTransSpecialFile(record.Id, srcFilePathName, desFile, srcLang, desLang)
	if err != nil {
		record.State = structs.TransTranslateFailed
		record.StateDescribe = structs.TransExtractFailed.String()
		record.Error = err.Error()
		datamodels.UpdateRecord(record)
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
