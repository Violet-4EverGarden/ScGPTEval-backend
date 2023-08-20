package models

import "time"

type Record struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"record_id"`
	QuizID       int64     `gorm:"column:quiz_id" json:"quiz_id"`
	UserID       int64     `gorm:"column:user_id" json:"user_id"`
	SelectOption string    `gorm:"column:selected_option"`
	CreatedAt    time.Time `gorm:"column:created_at"`
}
