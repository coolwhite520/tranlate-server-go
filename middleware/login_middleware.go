package middleware

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/jwt"
	"strings"
	"time"
	"translate-server/datamodels"
)

/*
Documentation:
    https://github.com/kataras/jwt#table-of-contents
*/

// Replace with your own key and keep them secret.
// The "signatureSharedKey" is used for the HMAC(HS256) signature algorithm.
var signatureSharedKey = []byte("sercrethatmaycontainch@r32length")

func GenerateToken( user datamodels.User) (string, error){
	// Sign and generate compact form token.
	token, err := jwt.Sign(jwt.HS256, signatureSharedKey, user, jwt.MaxAge(24 * time.Hour))
	if err != nil {
		return "", err
	}
	tokenString := string(token) // or jwt.BytesToString
	return tokenString, nil
}

// CheckLoginMiddleware 检测是否成功登录的中间件
func CheckLoginMiddleware(ctx iris.Context) {
	token := ctx.GetHeader("Authorization")
	split := strings.Split(token, " ")
	if len(split) < 2 {
		ctx.JSON(
			map[string]interface{}{
				"code": datamodels.HttpUserNotLogin,
				"msg": datamodels.HttpUserNotLogin.String(),
			})
		return
	}
	// Verify the token.
	token = split[1]
	verifiedToken, err := jwt.Verify(jwt.HS256, signatureSharedKey, []byte(token))
	if err != nil {
		ctx.JSON(
			map[string]interface{}{
			"code": datamodels.HttpUserExpired,
			"msg": datamodels.HttpUserExpired.String(),
		})
		return
	}

	var user datamodels.User
	verifiedToken.Claims(&user)
	//standardClaims := jwt.GetVerifiedToken(ctx).StandardClaims
	//expiresAtString := standardClaims.ExpiresAt().Format(ctx.Application().ConfigurationReadOnly().GetTimeFormat())
	//timeLeft := standardClaims.Timeleft()
	ctx.Values().Set("User", user)
	ctx.Next()

}
