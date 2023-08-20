package models

import "time"

// Quiz 题库表结构体
type Quiz struct {
	ID   int64  `gorm:"primaryKey;autoIncrement" json:"quiz_id"`
	Type string `gorm:"type:ENUM('1','2','3')" json:"quiz_type"`

	Content   string    `json:"content"`
	Options   string    `gorm:"type:text" json:"options"` // Golang map <--> mysql text
	CreatedAt time.Time `gorm:"column:created_at"`
}
