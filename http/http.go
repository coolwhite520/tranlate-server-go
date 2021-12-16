package http

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"translate-server/controller"
	"translate-server/datamodels"
	"translate-server/services"
)

func StartMainServer() {
	app := iris.New()
	mvc.Configure(app.Party("/api"), activationMVC, userMVC, usersMVC, translateMVC)
	app.Run(iris.Addr(":7777"))
}

func activationMVC(app *mvc.Application)  {
	party := app.Party("/activation")
	party.Handle(new(controller.ActivationController))
}

func userMVC(app *mvc.Application) {
	party := app.Party("/user")
	service := services.NewUserService()
	users, _ := service.QueryAdminUsers()
	if users == nil {
		password, _ := datamodels.GeneratePassword("admin")
		service.InsertUser(datamodels.User{
			Username:     fmt.Sprintf("admin"),
			HashedPassword: password,
			IsSuper:  true,
		})
	}
	party.Register(service)
	party.Handle(new(controller.UserController))
}


func usersMVC(app *mvc.Application) {
	party := app.Party("/users")
	service := services.NewUserService()
	party.Register(service)
	party.Handle(new(controller.UsersController))
}

func translateMVC(app *mvc.Application) {
	service := services.NewTranslateService()
	activationService := services.NewActivationService()
	party := app.Party("/translate")
	party.Register(service, activationService)
	party.Handle(new(controller.TranslateController))
}


