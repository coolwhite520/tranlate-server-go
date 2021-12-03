package activation

import "github.com/kataras/iris/v12"

func CheckSerialMiddleware(ctx iris.Context)  {
	ctx.Next()
}
