package translate_models

import (
	"fmt"
	"translate-server/datamodels"
	"translate-server/rpc"
	"translate-server/utils"
)

func translate(srcLang string, desLang string, content string) (string, string, error) {
	sha1 := utils.Sha1(fmt.Sprintf("%s&%s&%s", content, srcLang, desLang))
	records, err := datamodels.QueryTranslateRecordsBySha1(sha1)
	if err != nil {
		return "", "", err
	}
	var transContent string
	for _, v := range records {
		if v.SrcLang == srcLang && v.DesLang == desLang {
			if v.TransType == 0 {
				transContent = v.OutputContent
				break
			}
		}
	}
	if len(transContent) == 0 {
		transContent, err = rpc.PyTranslate(srcLang, desLang, content)
		if err != nil {
			return "", "", err
		}
	}
	return transContent, sha1, nil
}

