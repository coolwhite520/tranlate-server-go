package http

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	log "github.com/sirupsen/logrus"
	"translate-server/controllers"
)

func StartIntServer() {
	log.WithFields(log.Fields{
		"ServerResponse": true,
		"ReqUrl":         "baidu.com2",
	}).Info("gou dong xi 23")
	app := iris.Default()
	//authConfig := basicauth.Options{
	//	Allow: basicauth.AllowUsers(map[string]string{"panda": "000000"}),
	//}
	//authentication := basicauth.New(authConfig)
	//app.Use(authentication)
	mvc.Configure(app.Party("/api"), userInterfaceMvc, fileInterfaceMvc, textInterfaceMvc)
	app.Run(iris.Addr(fmt.Sprintf("%s:%d", "127.0.0.1", 3333)))


}

func userInterfaceMvc(app *mvc.Application) {
	app.Router.Use(func(ctx iris.Context) {
		ctx.Application().Logger().Infof("Path: %s", ctx.Path())
		ctx.Next()
	})
	app.Party("/user").Handle(new(controllers.UserController))
}

func textInterfaceMvc(app *mvc.Application) {
	app.Router.Use(func(ctx iris.Context) {
		ctx.Application().Logger().Infof("Path: %s", ctx.Path())
		ctx.Next()
	})
	app.Party("/text").Handle(new(controllers.TextController))
}

func fileInterfaceMvc(app *mvc.Application) {
	app.Router.Use(func(ctx iris.Context) {
		ctx.Application().Logger().Infof("Path: %s", ctx.Path())
		ctx.Next()
	})
	app.Party("/file").Handle(new(controllers.FileController))
}

