package routes

import (
	"github.com/gin-gonic/gin"
	"scgptEval/controllers"
	"scgptEval/middlewares"
)

func InitQuiz(quiz *gin.RouterGroup) {
	quiz.Use(middlewares.JWTAuthMiddleware())
	quiz.GET("/get_amount", controllers.GetUserAmount)
	quiz.GET("/get_ranking", controllers.GetRanking)

	quiz.POST("/submit_quiz", controllers.SubmitQuiz)
	quiz.POST("/get_quizzes", controllers.GetUntriedQuizzes)
}
