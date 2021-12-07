package middleware

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"translate-server/datamodels"
	"translate-server/services"
)

func isIn(target string, strArray []datamodels.SupportLang) bool {
	for _, element := range strArray {
		if target == element.EnName {
			return true
		}
	}
	return false
}


// SupportLangMiddleware 检测支持的语言列表的中间件
func SupportLangMiddleware(ctx iris.Context) {
	newActivation := services.NewActivationService()
	file, state := newActivation.ParseKeystoreFile()
	if state != services.Success {
		ctx.JSON(
			map[string]interface{}{
				"code":      -100,
				"msg":       state.String(),
			})
		return
	}
	srcLang := ctx.FormValue("src_lang")
	desLang := ctx.FormValue("des_lang")
	if !isIn(srcLang, file.SupportLangList) {
		ctx.JSON(
			map[string]interface{}{
				"code":      -100,
				"msg":       fmt.Sprintf("不支持的源语言,您当前的版本仅支持%v", file.SupportLangList),
			})
		return
	}
	if !isIn(desLang, file.SupportLangList) {
		ctx.JSON(
			map[string]interface{}{
				"code":      -100,
				"msg":       fmt.Sprintf("不支持的目标语言,您当前的版本仅支持%v", file.SupportLangList),
			})
		return
	}
	ctx.Next()
}
