package logic

import (
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"math/rand"
	"scgptEval/dao/mysql"
	"scgptEval/dao/redis"
	"time"
)

// SubmitQuiz 提交题目的logic层逻辑
// 1.数据库：插入刷题记录record, 并更新用户最后刷题时间user —— 事务操作
// 2.缓存：更新三个Key
func SubmitQuiz(userID, quizID int64, selectOptions string) (err error) {
	if err = mysql.SubmitQuiz(userID, quizID, selectOptions); err != nil {
		if errors.Is(err, mysql.ErrorQuizNotExist) {
			return
		} else if errors.Is(err, mysql.ErrorRecordExist) {
			return
		}
		zap.L().Error("mysql.SubmitQuiz() failed",
			zap.Int64("user_id", userID),
			zap.Int64("quiz_id", quizID), zap.Error(err))
		return
	}

	if err = redis.SubmitQuiz(fmt.Sprint(userID), fmt.Sprint(quizID)); err != nil {
		zap.L().Error("redis.SubmitQuiz() failed",
			zap.Int64("user_id", userID),
			zap.Int64("quiz_id", quizID), zap.Error(err))
		return
	}
	return
}

// GetUntriedQuizzes 获取未刷过的题目
func GetUntriedQuizzes(quizNum int, userID int64) (quizList []map[string]interface{}, err error) {
	// 1.先从record中找出用户刷过的题目id
	doneQuizIds, err := mysql.GetDoneQuizID(userID)
	if err != nil {
		return nil, err
	}

	// 2.从题库中找出未刷过的指定数量题目
	// fixme 暂未设计{剩余题目数}小于{quizNum}的逻辑
	quizzes, err := mysql.GetUntriedQuizzes(doneQuizIds, quizNum)
	if err != nil {
		return nil, err
	}

	for _, quiz := range quizzes {
		quizData := make(map[string]interface{})
		options := make(map[string]string)
		quizData["quiz_id"] = quiz.ID
		quizData["type"] = quiz.Type
		quizData["content"] = quiz.Content
		// 对选项内容作反序列化
		err := json.Unmarshal([]byte(quiz.Options), &options)
		if err != nil {
			continue
		}
		quizData["options"] = options

		quizList = append(quizList, quizData)
	}
	return quizList, nil
}

// GetUntriedQuizzes2 获取未刷过的题目 (利用redis交集运算快速获取未刷过的题目id)
func GetUntriedQuizzes2(quizNum int, userID int64) (quizList []map[string]interface{}, err error) {
	// 1.先从redis取出用户未刷过的题目id列表
	quizIds, err := redis.GetUntriedQuiz(fmt.Sprint(userID))

	// 2.从中随机挑选指定个id
	// fixme 暂未设计{剩余题目数}小于{quizNum}的逻辑
	ShuffleKnuth(quizIds)
	selectedIds := quizIds[:quizNum]

	// 3.查询对应的记录
	quizzes, err := mysql.GetQuizzes(selectedIds)
	if err != nil {
		return nil, err
	}
	var quizData map[string]interface{}
	for _, quiz := range quizzes {
		quizData["quiz_id"] = quiz.ID
		quizData["type"] = quiz.Type
		quizData["content"] = quiz.Content
		quizData["options"] = quiz.Options

		quizList = append(quizList, quizData)
	}
	return quizList, nil
}

// ShuffleKnuth Fisher-Yates-Knuth快速洗牌算法
func ShuffleKnuth(slice []string) {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	n := len(slice)
	for i := n - 1; i > 0; i-- {
		j := random.Intn(i + 1)
		slice[i], slice[j] = slice[j], slice[i]
	}
}
