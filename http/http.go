package http

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"translate-server/controller"
	"translate-server/datamodels"
	"translate-server/services"
)

func StartIntServer() {
	app := iris.New()
	mvc.Configure(app.Party("/api"), activationMVC, userMVC, usersMVC, translateMVC)
	app.Run(iris.Addr(":8080"))
}

func activationMVC(app *mvc.Application)  {
	party := app.Party("/middleware")
	party.Handle(new(controller.ActivationController))
}

func userMVC(app *mvc.Application) {
	party := app.Party("/user")
	service := services.NewUserService()
	users, _ := service.QueryAllUsers()
	if users == nil {
		password, _ := datamodels.GeneratePassword("admin")
		service.InsertUser(datamodels.User{
			Username:     "admin",
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
	app.Party("/translate").Handle(new(controller.TranslateController))
}


