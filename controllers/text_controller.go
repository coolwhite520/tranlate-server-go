package controllers

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
)

type TextController struct {
	Ctx iris.Context
}

func (c *TextController) Post() mvc.Result {
	return mvc.Response{}
}