package http

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"translate-server/controller"
)

func StartIntServer() {
	app := iris.New()
	//users, _ := repositories.QueryAllUsers()
	//m := make(map[string]string)
	//for _, v := range users {
	//	m[v.Username] = v.Password
	//}
	//authConfig := basicauth.Options{
	//	Allow: basicauth.AllowUsers(m),
	//}
	//authentication := basicauth.New(authConfig)
	//app.Use(authentication)
	mvc.Configure(app.Party("/api"), fileMVC, textMVC, userMVC)
	app.Run(iris.Addr(":8080"))
}

func userMVC(app *mvc.Application) {
	app.Router.Use(func(ctx iris.Context) {
		ctx.Application().Logger().Infof("Path: %s", ctx.Path())
		ctx.Next()
	})
	app.Party("/user").Handle(new(controller.UserController))
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

