package translate_models

import (
	log "github.com/sirupsen/logrus"
	"path/filepath"
	"translate-server/apis"
	"translate-server/datamodels"
)

// TranslateFile 异步翻译，将翻译推送到file模块进行处理
func TranslateFile(srcLang string, desLang string, recordId int64, userId int64) {
	// 先查找是否存在相同的翻译结果
	record, _ := datamodels.QueryTranslateRecordByIdAndUserId(recordId, userId)
	if record == nil {
		log.Error("查询不到RecordId为", recordId, "的记录")
		return
	}
	// 更新语言
	record.SrcLang = srcLang
	record.DesLang = desLang
	datamodels.UpdateRecord(record)
	dataAbsPath, err := filepath.Abs("./data")
	if err != nil {
		log.Error(err)
		return
	}
	// 让file容器接管并全权负责
	err = apis.RpcTransFile(recordId, dataAbsPath)
	if err != nil {
		log.Error(err)
		return
	}
}
