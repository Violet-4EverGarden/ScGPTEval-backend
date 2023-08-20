package routes

import (
	"net/http"
	"scgptEval/logger"

	"github.com/gin-gonic/gin"
)

func SetUp() *gin.Engine {
	r := gin.New()
	r.Use(logger.GinLogger(), logger.GinRecovery(true))

	user := r.Group("/user")
	InitUser(user)

	quiz := r.Group("/quiz")
	InitQuiz(quiz)

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	return r
}
