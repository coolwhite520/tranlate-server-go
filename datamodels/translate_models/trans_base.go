package translate_models

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
	"translate-server/apis"
	"translate-server/datamodels"
	"translate-server/utils"
)

func translate(srcLang string, desLang string, content string) (string, string, error) {
	sha1 := utils.Sha1(fmt.Sprintf("%s&%s&%s", content, srcLang, desLang))
	transContent := datamodels.GetRedisString(sha1)
	var err error
	if len(transContent) > 0 {
		return transContent, sha1, nil
	}
	if len(transContent) == 0 {
		transContent, err = apis.PyTranslate(srcLang, desLang, content)
		if err != nil {
			log.Errorln(err)
			return "", "", err
		}
		datamodels.SetRedisString(sha1, transContent, time.Hour * 24)
	}
	return transContent, sha1, nil
}

