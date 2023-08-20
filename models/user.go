package models

import "time"

// User 数据库user表结构体
type User struct {
	ID       int64  `gorm:"primaryKey;autoIncrement" json:"user_id,string"`
	Username string `gorm:"unique" json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`

	Amount      int64     `json:"amount"`
	FinalQATime time.Time `gorm:"column:final_qa_time" json:"final_qa_time"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}
