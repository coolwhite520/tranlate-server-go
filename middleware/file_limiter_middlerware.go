package middleware

import (
	"fmt"
	"github.com/kataras/iris/v12"
)

const MaxNumber = 8
const maxSize = MaxNumber * iris.MB

// FileLimiterMiddleware 上传文件大小检测的中间件
func FileLimiterMiddleware(ctx iris.Context) {
	if ctx.GetContentLength() > maxSize {
		ctx.JSON(map[string]interface{}{
			"code": -100,
			"msg":  fmt.Sprintf("文件大小超过了 %d MB", MaxNumber),
		})
		return
	}
	ctx.Next()
}
