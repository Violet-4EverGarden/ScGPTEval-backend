package routes

import (
	"github.com/gin-gonic/gin"
	"scgptEval/controllers"
	"scgptEval/middlewares"
)

func InitUser(user *gin.RouterGroup) {
	user.POST("/signup", controllers.SignUp)
	user.POST("/login", controllers.LogIn)
	// access token过期时刷新，需要携带过期token
	user.GET("/refresh_token", controllers.RefreshTokenHandler)
	user.Use(middlewares.JWTAuthMiddleware())
	user.POST("/change_name", controllers.ChangeName)
}
