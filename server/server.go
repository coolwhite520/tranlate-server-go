package server

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"net"
	"translate-server/controller"
	"translate-server/middleware"
	"translate-server/services"
)

func StartMainServer(listener net.Listener) {
	app := iris.New()
	app.Use(middleware.IpAccessMiddleware)
	mvc.Configure(app.Party("/api"), activationMVC, userMVC, adminMVC, translateMVC)
	app.Run(iris.Listener(listener))
}
// 激活
func activationMVC(app *mvc.Application)  {
	party := app.Party("/activation")
	newActivation:= services.NewActivationService()
	party.Register(newActivation)
	party.Handle(new(controller.ActivationController))
}
// 管理员
func adminMVC(app *mvc.Application) {
	party := app.Party("/admin")
	service := services.NewAdminService()
	party.Register(service)
	party.Handle(new(controller.AdminController))
}

// 用户登录修改密码等操作
func userMVC(app *mvc.Application) {
	party := app.Party("/user")
	service := services.NewUserService()
	party.Register(service)
	party.Handle(new(controller.UserController))
}

// 翻译功能
func translateMVC(app *mvc.Application) {
	service := services.NewTranslateService()
	party := app.Party("/translate")
	party.Register(service)
	party.Handle(new(controller.TranslateController))
}


