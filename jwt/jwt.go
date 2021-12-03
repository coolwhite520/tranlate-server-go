package jwt

import (
	"fmt"
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

func GenerateToken(ctx iris.Context, user datamodels.User) {
	// Sign and generate compact form token.
	token, err := jwt.Sign(jwt.HS256, signatureSharedKey, user, jwt.MaxAge(10*time.Minute))
	if err != nil {
		ctx.StopWithStatus(iris.StatusInternalServerError)
		return
	}
	tokenString := string(token) // or jwt.BytesToString
	ctx.Header("Authorization", fmt.Sprintf("Bearer %s", tokenString))
}

func ParseToken(ctx iris.Context) {
	// Extract the token, e.g. cookie, Authorization: Bearer $token
	// or URL query.
	token := ctx.GetHeader("Authorization")
	split := strings.Split(token, " ")
	if len(split) < 2 {
		ctx.JSON(
			map[string]interface{}{
				"code": -100,
				"err": "账户未登录",
			})
		return
	}
	// Verify the token.
	verifiedToken, err := jwt.Verify(jwt.HS256, signatureSharedKey, []byte(token))
	if err != nil {
		ctx.JSON(
			map[string]interface{}{
			"code": -100,
			"err": "账户未登录",
		})
		return
	}
	// Decode the custom claims.
	var claims datamodels.User
	verifiedToken.Claims(&claims)
	//// Just an example on how you can retrieve all the standard claims (set by jwt.MaxAge, "exp").
	//standardClaims := jwt.GetVerifiedToken(ctx).StandardClaims
	//expiresAtString := standardClaims.ExpiresAt().Format(ctx.Application().ConfigurationReadOnly().GetTimeFormat())
	//timeLeft := standardClaims.Timeleft()
	ctx.Next()

}
