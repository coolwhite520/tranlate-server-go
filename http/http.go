package http

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/basicauth"
	"translate-server/models"
)

func StartIntServer() {
	app := iris.New()
	users, _ := models.QueryAllUsers()
	m := make(map[string]string)
	for _, v := range users {
		m[v.Name] = v.Password
	}
	authConfig := basicauth.Options{
		Allow: basicauth.AllowUsers(m),
	}
	authentication := basicauth.New(authConfig)
	app.Use(authentication)
	url := fmt.Sprintf("%s:%d", "127.0.0.1", 3333)
	app.Run(iris.Addr(url))
}
