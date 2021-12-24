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
	//ws := mvc.New(app.Party("/api/ws/upload"))
	//ws.HandleWebsocket(&controller.WebsocketController{ Namespace: "default"} )
	//websocketServer := neffos.New(websocket.DefaultGorillaUpgrader, ws)
	//ws.Router.Get("/", websocket.Handler(websocketServer))

	mvc.Configure(app.Party("/api"), activationMVC, userMVC, adminMVC, translateMVC)
	app.Run(iris.Listener(listener))
}

func activationMVC(app *mvc.Application)  {
	party := app.Party("/activation")
	party.Handle(new(controller.ActivationController))
}

func userMVC(app *mvc.Application) {
	party := app.Party("/user")
	service := services.NewUserService()
	party.Register(service)
	party.Handle(new(controller.UserController))
}


func adminMVC(app *mvc.Application) {
	party := app.Party("/admin")
	service := services.NewUserService()
	party.Register(service)
	party.Handle(new(controller.AdminController))
}

func translateMVC(app *mvc.Application) {
	service := services.NewTranslateService()
	activationService := services.NewActivationService()
	party := app.Party("/translate")
	party.Register(service, activationService)
	party.Handle(new(controller.TranslateController))
}


