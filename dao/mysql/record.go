package mysql

import (
	"errors"
	"github.com/jinzhu/gorm"
	"scgptEval/models"
)

var (
	ErrorRecordExist    = errors.New("记录已存在")
	ErrorRecordNotExist = errors.New("记录不存在")
)

// CheckRecordExist 检查是否已有做题记录
func CheckRecordExist(userId, quizID int64) (err error) {
	if res := db.Where("user_id = ? AND quiz_id = ?", userId, quizID).
		First(&models.Record{}); res.RowsAffected > 0 {
		// 用户已存在
		return ErrorRecordExist
	}
	return nil
}

func GetDoneQuizID(userID int64) (quizIds []int64, err error) {
	var records []models.Record
	if err = db.Where("user_id = ?", userID).Find(&records).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []int64{}, nil
		}
		return nil, err
	}
	for _, record := range records {
		quizIds = append(quizIds, record.QuizID)
	}
	return quizIds, nil
}
