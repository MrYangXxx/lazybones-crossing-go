package middleware

import (
	"github.com/dgrijalva/jwt-go"
	"hidevops.io/hiboot/pkg/app"
	"hidevops.io/hiboot/pkg/app/web/context"
	"hidevops.io/hiboot/pkg/at"
	hibootjwt "hidevops.io/hiboot/pkg/starter/jwt"
)

type CasBinMiddleware struct {
	at.Middleware
	token hibootjwt.Token
}

func init() {
	app.Register(newCasBinMiddleware)
}

func newCasBinMiddleware(token hibootjwt.Token) *CasBinMiddleware {
	return &CasBinMiddleware{
		token: token,
	}
}

//
func (m *CasBinMiddleware) Auth(at struct {
	at.MiddlewareHandler `value:"/" `
}, ctx context.Context) error {

	//可以忽略检查的请求路径
	canIgnore := ctx.Request().URL.Path == "/user/login" || ctx.Request().URL.Path == "/user/info" || ctx.Request().URL.Path == "/user/registry" || ctx.Request().URL.Path == "/captcha"

	if canIgnore {
		ctx.Next()
		return nil
	}

	//其他请求验证token和访问权限
	//token, _ := request.ParseFromRequest(ctx.Request(), request.AuthorizationHeaderExtractor, func(token *jwt.Token) (interface{}, error) {
	//	return m.token.VerifyKey(), nil
	//})
	token, _ := jwt.Parse(ctx.GetCookie("lazybones_token"), func(*jwt.Token) (interface{}, error) {
		return m.token.VerifyKey(), nil
	})

	//token无效
	if token == nil || !token.Valid {
		ctx.ResponseWriter().WriteHeader(401)
		return nil
	}

	ctx.Next()

	return nil
}
