package middlewares

import (
	"scgptEval/controllers"
	"scgptEval/pkg/jwt"
	"strings"

	"github.com/gin-gonic/gin"
)

// JWTAuthMiddleware 基于JWT的认证中间件
func JWTAuthMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		// 客户端携带Token有三种方式 1.放在请求头 2.放在请求体 3.放在URI
		// 这里假设Token放在Header的Authorization中，并使用Bearer开头，即 Authorization: Bearer xxxxx.xxxxx.xxxxx
		// 这里的具体实现方式要依据你的实际业务情况决定 (例如携带Token的方式也有可能是 X-Token: xxxxx.xxxxx.xxxxx)
		authHeader := c.Request.Header.Get("Authorization")

		// 请求头中auth为空，未携带相关Token，表示用户还未登录
		if authHeader == "" {
			controllers.ResponseError(c, controllers.CodeNeedLogin)
			c.Abort()
			return
		}

		// 按空格分割，并且判断是否为Bearer开头
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			controllers.ResponseError(c, controllers.CodeNeedLogin)
			c.Abort()
			return
		}

		// parts[1]是获取到的tokenString，我们使用之前定义好的解析JWT的函数来解析它
		mc, err := jwt.ParseToken(parts[1])
		if err != nil {
			controllers.ResponseError(c, controllers.CodeInvalidToken) // 无效Token
			c.Abort()
			return
		}
		// 将当前请求的username信息保存到请求的上下文c上
		c.Set(controllers.ContextUserIDKey, mc.UserID)
		c.Next() // 在后续的处理请求的函数中 可以通过c.Get(ContextUserIDKey)来获取当前请求的用户信息（见request.go中）
	}
}
