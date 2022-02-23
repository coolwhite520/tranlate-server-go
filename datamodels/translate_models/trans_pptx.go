package translate_models

import (
	"fmt"
	"path/filepath"
	"strings"
	"translate-server/apis"
	"translate-server/datamodels"
	"translate-server/structs"
	"translate-server/utils"
)

func translatePptxFile(srcLang string, desLang string, record *structs.Record) error {
	srcDir := fmt.Sprintf("%s/%d/%s", structs.UploadDir, record.UserId, record.DirRandId)
	translatedDir := fmt.Sprintf("%s/%d/%s", structs.OutputDir, record.UserId, record.DirRandId)
	srcFilePathName := fmt.Sprintf("%s/%s%s", srcDir, record.FileName, record.FileExt)
	desFile := fmt.Sprintf("%s/%s%s", translatedDir, record.FileName, record.OutFileExt)
	ext := filepath.Ext(record.FileExt)
	if strings.ToLower(ext) == ".ppt" {
		err := apis.PyConvertSpecialFile(srcFilePathName, "p2p")
		if err != nil {
			return err
		}
		srcFilePathName = srcFilePathName + "x"
	}
	err := apis.PyTransSpecialFile(record.Id, srcFilePathName, desFile, srcLang, desLang)
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
