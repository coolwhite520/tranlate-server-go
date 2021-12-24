package server

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"net"
	"translate-server/controller"
	"translate-server/services"
)

func StartMainServer(listener net.Listener) {
	app := iris.New()
	mvc.Configure(app.Party("/api"), activationMVC, userMVC, adminMVC, translateMVC)
	app.Run(iris.Listener(listener))
}
// 激活的
func activationMVC(app *mvc.Application)  {
	app.Router.Use(func(ctx iris.Context) {
		//log.Infof("Path: %s", ctx.Path())
		ctx.Next()
	})
	party := app.Party("/activation")
	party.Handle(new(controller.ActivationController))
}
// 用户的
func userMVC(app *mvc.Application) {
	app.Router.Use(func(ctx iris.Context) {
		//log.Infof("Path: %s", ctx.Path())
		ctx.Next()
	})
	party := app.Party("/user")
	service := services.NewUserService()
	party.Register(service)
	party.Handle(new(controller.UserController))
}

// 管理
func adminMVC(app *mvc.Application) {
	app.Router.Use(func(ctx iris.Context) {
		//log.Infof("Path: %s", ctx.Path())
		ctx.Next()
	})
	party := app.Party("/admin")
	service := services.NewUserService()
	party.Register(service)
	party.Handle(new(controller.AdminController))
}
// 翻译
func translateMVC(app *mvc.Application) {
	app.Router.Use(func(ctx iris.Context) {
		//log.Infof("Path: %s", ctx.Path())
		ctx.Next()
	})
	service := services.NewTranslateService()
	activationService := services.NewActivationService()
	party := app.Party("/translate")
	party.Register(service, activationService)
	party.Handle(new(controller.TranslateController))
}


