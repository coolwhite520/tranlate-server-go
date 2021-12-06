package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"translate-server/activation"
	"translate-server/jwt"
)

const maxSize = 5 << 20 // 5MB

type FileController struct {
	Ctx iris.Context
}


func (f *FileController) BeforeActivation(b mvc.BeforeActivation) {
	b.Router().Use(activation.CheckActivationMiddleware)
	b.Router().Use(jwt.CheckLoginMiddleware)
}

func (f *FileController) PostUpload() mvc.Result {
	f.Ctx.UploadFormFiles("./uploads")
	return mvc.Response{}
}
