package controller

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	log "github.com/sirupsen/logrus"
)

type TextController struct {
	Ctx iris.Context
}

func (t *TextController) BeforeActivation(b mvc.BeforeActivation) {
	b.Handle("GET", "/{id:int32}", "MyById")
}

func (t *TextController) MyById(Id int32) {
	log.Info(Id)
	t.Ctx.JSON(iris.Map{"name": "panda", "id": Id})
}