package http

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"translate-server/activation"
	"translate-server/controller"
	"translate-server/datamodels"
	"translate-server/services"
)

func StartIntServer() {
	app := iris.New()
	//// Load the template files.
	//tmpl := iris.HTML("./web/views", ".html").
	//	Layout("shared/layout.html").
	//	Reload(true)
	//app.RegisterView(tmpl)
	//app.HandleDir("/public", iris.Dir("./web/public"))
	party := app.Party("/api")
	party.Use(activation.CheckSerialMiddleware)
	mvc.Configure(party, userMVC, usersMVC, fileMVC, textMVC)
	app.Run(iris.Addr(":8080"))
}


func userMVC(app *mvc.Application) {
	app.Router.Use(func(ctx iris.Context) {
		ctx.Application().Logger().Infof("Path: %s", ctx.Path())
		ctx.Next()
	})
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
	app.Router.Use(func(ctx iris.Context) {
		ctx.Application().Logger().Infof("Path: %s", ctx.Path())
		ctx.Next()
	})
	party := app.Party("/users")
	service := services.NewUserService()
	party.Register(service)
	party.Handle(new(controller.UsersController))
}

func fileMVC(app *mvc.Application) {
	app.Router.Use(func(ctx iris.Context) {
		ctx.Application().Logger().Infof("Path: %s", ctx.Path())
		ctx.Next()
	})
	app.Party("/file").Handle(new(controller.FileController))
}

func textMVC(app *mvc.Application) {
	app.Router.Use(func(ctx iris.Context) {
		ctx.Application().Logger().Infof("Path: %s", ctx.Path())
		ctx.Next()
	})
	app.Party("/text").Handle(new(controller.TextController))
}

