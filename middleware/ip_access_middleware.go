package middleware

import (
	"github.com/kataras/iris/v12"
	"github.com/thinkeridea/go-extend/exnet"
)

func IpAccessMiddleware(ctx iris.Context) {
	_ = exnet.ClientIP(ctx.Request())
	ctx.Next()
}
