package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"translate-server/services"
)

type ActivationController struct {
	Ctx iris.Context
	NewActivation services.ActivationService
}

func (a *ActivationController) BeforeActivation(b mvc.BeforeActivation) {

}
//Post 激活码上传校验
func (a *ActivationController) Post() mvc.Result {
	return a.NewActivation.Activation(a.Ctx)
}


