package mysql

import (
	"errors"
	"github.com/jinzhu/gorm"
	"math/rand"
	"scgptEval/models"
	"time"
)

var (
	ErrorQuizNotExist = errors.New("题目不存在")
)

// CheckQuizExist 检查提交的题目是否合法(存在)
func CheckQuizExist(quizID int64) (err error) {
	if res := db.Where("id = ?", quizID).First(&models.Quiz{}); res.RowsAffected == 0 {
		// 题目不存在
		return ErrorQuizNotExist
	}
	return nil
}

// SubmitQuiz 提交题目
func SubmitQuiz(userID, quizID int64, selectOptions string) (err error) {
	if err = CheckQuizExist(quizID); err != nil {
		// 检查题目是否真实存在，防止非法插入
		return
	}
	// fixme 如果有重复提交的需求可改下面的逻辑
	if err = CheckRecordExist(userID, quizID); err != nil {
		return
	}
	record := &models.Record{
		QuizID:       quizID,
		UserID:       userID,
		SelectOption: selectOptions,
		CreatedAt:    time.Now(),
	}

	// 启动事务
	tx := db.Begin()
	defer func() {
		// 使用defer延迟函数来确保在函数返回之前若出现panic始终执行回滚事务的操作
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Record表插入新纪录
	if err = tx.Create(record).Error; err != nil {
		tx.Rollback()
		return
	}
	// 更新user的最后刷题时间
	if err = tx.Model(&models.User{}).
		Where("id = ?", userID).
		Update("final_qa_time", time.Now()).Error; err != nil {
		tx.Rollback()
		return
	}
	if err = tx.Commit().Error; err != nil { // 提交事务
		tx.Rollback() // 事务提交失败也回滚
		return
	}
	return
}

// GetQuizzes 根据id获取题目记录
func GetQuizzes(ids []string) (quizzes []models.Quiz, err error) {
	// 由于ids是quiz表主键切片，所以可直接这样查询
	if err = db.Where(ids).Find(&quizzes).Error; err != nil {
		return nil, err
	}
	return quizzes, nil
}

// GetUntriedQuizzes 随机获取指定数量的未刷过题目记录
func GetUntriedQuizzes(doneQuizIds []int64, num int) (quizzes []models.Quiz, err error) {
	rand.Seed(time.Now().UnixNano())

	queryDB := db
	if len(doneQuizIds) > 0 {
		queryDB = db.Where("id NOT IN (?)", doneQuizIds)
	}
	err = queryDB.Order(gorm.Expr("RAND()")).
		Limit(num).
		Find(&quizzes).Error
	if err != nil {
		return nil, err
	}
	return quizzes, nil
}
