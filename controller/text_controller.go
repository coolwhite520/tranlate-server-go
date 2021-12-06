package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	log "github.com/sirupsen/logrus"
	"translate-server/activation"
	"translate-server/datamodels"
	"translate-server/jwt"
)

type TextController struct {
	Ctx iris.Context
}

func (t *TextController) BeforeActivation(b mvc.BeforeActivation) {
	b.Router().Use(activation.CheckActivationMiddleware)
	b.Router().Use(jwt.CheckLoginMiddleware)
}

func (t *TextController) Get() mvc.Result {
	// 通过中间件对token解析，然后传递出来，可以在这里进行获取并使用
	middleUser := t.Ctx.Values().Get("User")
	// 强制类型转换一下
	user, ok := (middleUser).(datamodels.User)
	if ok {
		log.Info(user)
	}
	return mvc.Response{}
}