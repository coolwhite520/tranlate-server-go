package controller

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"translate-server/middleware"
	"translate-server/services"
)
const MaxNumber = 8
const maxSize = MaxNumber * iris.MB

const UploadDir = "./uploads"

type TranslateController struct {
	Ctx iris.Context
	TranslateService services.TranslateService
}

func FileLimiterMiddleware(ctx iris.Context) {
	if ctx.GetContentLength() > maxSize {
		ctx.JSON(map[string]interface{}{
			"code": -100,
			"msg":  fmt.Sprintf("文件大小超过了 %d MB", MaxNumber),
		})
		return
	}
	ctx.Next()
}


func (t *TranslateController) BeforeActivation(b mvc.BeforeActivation) {
	b.Router().Use(middleware.CheckActivationMiddleware, middleware.CheckLoginMiddleware, FileLimiterMiddleware)
}


func (t *TranslateController) PostUpload() mvc.Result {
	files, n, err := t.Ctx.UploadFormFiles(UploadDir)
	if err != nil {
		return mvc.Response{
			Object: map[string]interface{}{
				"code": -100,
				"msg":  err.Error(),
			},
		}
	}
	for _, v := range files {
		//println(v.Filename)
		filePath := fmt.Sprintf("%s/%s", UploadDir, v.Filename)
		println(filePath)
		go func() {
			_, err := t.TranslateService.TranslateFile("", "", filePath)
			if err != nil {
				// 发现错误，异步写入到个人翻译记录中
			}
		}()
	}
	return mvc.Response{
		Object: map[string]interface{}{
			"code": 200,
			"msg":  n,
		},
	}
}
