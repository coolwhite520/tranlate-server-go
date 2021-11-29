package server

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/basicauth"
)

func StartIntServer() {
	app := iris.New()
	authConfig := basicauth.Options{
		Allow: basicauth.AllowUsers(map[string]string{"panda": "000000"}),
	}
	authentication := basicauth.New(authConfig)
	app.Use(authentication)
	app.Post("/uploads", uploadFile)
	app.Run(iris.Addr(fmt.Sprintf("%s:%d", "127.0.0.1", 3333)))
}
