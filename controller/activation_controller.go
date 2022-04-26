package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"translate-server/middleware"
	"translate-server/services"
)

type ActivationController struct {
	Ctx iris.Context
	NewActivation services.ActivationService
}

func (a *ActivationController) BeforeActivation(b mvc.BeforeActivation) {
	b.Router().Use(middleware.CheckLoginMiddleware, middleware.CheckSuperMiddleware)
	b.Handle("POST", "/ban", "PostAddBan")
	b.Handle("POST", "/proof", "PostActivationProof")
}
//Post 激活码上传校验
func (a *ActivationController) Post() mvc.Result {
	return a.NewActivation.Activation(a.Ctx)
}

//PostAddBan 失效当前授权
func (a *ActivationController) PostAddBan() mvc.Result {
	return a.NewActivation.AddBan()
}

// PostActivationProof 获取授权状态凭证
func (a *ActivationController) PostActivationProof() {
	a.NewActivation.PostActivationProof(a.Ctx)
}