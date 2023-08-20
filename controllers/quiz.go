package controllers

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"scgptEval/dao/mysql"
	"scgptEval/dao/redis"
	"scgptEval/logic"
)

// GetUserAmount 获取用户刷题量: GET
func GetUserAmount(c *gin.Context) {
	userID, err := getCurrentUser(c)
	if err != nil {
		zap.L().Error("getCurrentUser() failed", zap.Error(err))
		ResponseError(c, CodeInvalidToken)
		return
	}
	amount, err := redis.GetUserAmount(fmt.Sprint(userID))
	if err != nil {
		zap.L().Error("redis.GetUserAmount() failed", zap.Int64("userID", userID), zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}

	ResponseSuccess(c, gin.H{
		"user_id": fmt.Sprint(userID),
		"amount":  amount,
	})
}

// GetRanking 获取刷题排行信息: GET
func GetRanking(c *gin.Context) {
	// 默认展示刷题量前20名的用户信息
	rank, err := redis.GetRanking(20)
	if err != nil {
		zap.L().Error("redis.GetRanking() failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}

	ResponseSuccess(c, rank)
}

// SubmitQuiz 提交题目: POST
func SubmitQuiz(c *gin.Context) {
	// 1.请求参数获取并校验
	var req struct {
		QuizID        int64  `json:"quiz_id,string" binding:"required"`
		SelectOptions string `json:"select_options" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, CodeInvalidParam)
		return
	}
	// 获取当前请求的用户id
	userID, err := getCurrentUser(c)
	if err != nil {
		zap.L().Error("getCurrentUser() failed", zap.Error(err))
		ResponseError(c, CodeInvalidToken)
		return
	}
	// 2.执行提交逻辑：更新数据库与缓存
	if err = logic.SubmitQuiz(userID, req.QuizID, req.SelectOptions); err != nil {
		if errors.Is(err, mysql.ErrorQuizNotExist) {
			ResponseError(c, CodeQuizNotExist)
			return
		} else if errors.Is(err, mysql.ErrorRecordExist) {
			ResponseError(c, CodeRecordExist)
			return
		}
		zap.L().Error("logic.SubmitQuiz() failed",
			zap.Int64("user_id", userID),
			zap.Int64("quiz_id", req.QuizID), zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, nil)
}

// GetUntriedQuizzes 获取指定数量的题目: POST
func GetUntriedQuizzes(c *gin.Context) {
	var req struct {
		QuizNum int `json:"quiz_num" binding:"oneof=20 50 100"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseErrorWithMsg(c, CodeInvalidParam, "题目数仅限20，50，100")
		return
	}
	userID, err := getCurrentUser(c)
	if err != nil {
		zap.L().Error("getCurrentUser() failed", zap.Error(err))
		ResponseError(c, CodeInvalidToken)
		return
	}
	// 从redis获取用户未刷过的题目id列表，再从mysql中随机挑选目标条数的题目
	quizList, err := logic.GetUntriedQuizzes(req.QuizNum, userID)
	if err != nil {
		zap.L().Error("logic.GetUntriedQuizzes() failed",
			zap.Int64("user_id", userID),
			zap.Int("Need Quiz Num", req.QuizNum), zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, quizList)
}
