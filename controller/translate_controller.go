package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"translate-server/middleware"
	"translate-server/services"
)

type TranslateController struct {
	Ctx iris.Context
	TranslateService services.TranslateService
	ActivationService services.ActivationService
}

func (t *TranslateController) BeforeActivation(b mvc.BeforeActivation) {
	b.Router().Use(middleware.CheckActivationMiddleware,
		middleware.CheckLoginMiddleware,
		middleware.FileLimiterMiddleware)
	b.Handle("POST", "/upload", "PostUpload",middleware.SupportLangMiddleware)
	b.Handle("GET", "/lang", "GetLangList")
}

// GetLangList 获取支持的语言
func (t *TranslateController) GetLangList() mvc.Result {
	file, state := t.ActivationService.ParseKeystoreFile()
	if state != services.Success {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": -100,
				"msg": state.String(),
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": 200,
			"msg": "success",
			"data":file.SupportLangList,
		},
	}
}


func (t *TranslateController) PostUpload() mvc.Result {
	list, err := t.TranslateService.ReceiveFiles(t.Ctx)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": -100,
				"msg": err.Error(),
			},
		}
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": 200,
			"msg": "success",
			"data": list,
		},
	}
}
