package server

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	log "github.com/sirupsen/logrus"
	"net"
	"translate-server/controller"
	"translate-server/middleware"
	"translate-server/services"
)

func StartMainServer(listener net.Listener) {
	app := iris.New()
	mvc.Configure(app.Party("/api"), activationMVC, userMVC, adminMVC, translateMVC)
	app.Run(iris.Listener(listener))
}
// 激活的
func activationMVC(app *mvc.Application)  {

	party := app.Party("/activation")
	newActivation, err := services.NewActivationService()
	if err != nil {
		log.Errorln(err)
		panic(err)
	}
	party.Register(newActivation)
	party.Handle(new(controller.ActivationController))
}
// 用户的
func userMVC(app *mvc.Application) {
	app.Router.Use(middleware.IpAccessMiddleware)
	party := app.Party("/user")
	service := services.NewUserService()
	party.Register(service)
	party.Handle(new(controller.UserController))
}

// 管理
func adminMVC(app *mvc.Application) {
	party := app.Party("/admin")
	service := services.NewUserService()
	party.Register(service)
	party.Handle(new(controller.AdminController))
}
// 翻译
func translateMVC(app *mvc.Application) {
	app.Router.Use(middleware.IpAccessMiddleware)
	service := services.NewTranslateService()
	activationService, err := services.NewActivationService()
	if err != nil {
		log.Errorln(err)
		panic(err)
	}
	party := app.Party("/translate")
	party.Register(service, activationService)
	party.Handle(new(controller.TranslateController))
}


