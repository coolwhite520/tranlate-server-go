package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"translate-server/activation"
	"translate-server/jwt"
)

type FileController struct {
	Ctx iris.Context
}


func (f *FileController) BeforeActivation(b mvc.BeforeActivation) {
	b.Router().Use(activation.CheckActivationMiddleware)
	b.Router().Use(jwt.CheckTokenMiddleware)
}
