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
	app.Use(middleware.IpAccessMiddleware)
	mvc.Configure(app.Party("/api"), activationMVC, userMVC, adminMVC, translateMVC)
	app.Run(iris.Listener(listener))
}
// 激活
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
// 管理员
func adminMVC(app *mvc.Application) {
	party := app.Party("/admin")
	service := services.NewUserService()
	tableService := services.NewIpTableService()
	party.Register(service, tableService)
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
	activationService, err := services.NewActivationService()
	if err != nil {
		log.Errorln(err)
		panic(err)
	}
	party := app.Party("/translate")
	party.Register(service, activationService)
	party.Handle(new(controller.TranslateController))
}


