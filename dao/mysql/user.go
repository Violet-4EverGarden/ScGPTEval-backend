package mysql

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
	"scgptEval/models"
	"scgptEval/pkg/jwt"
	"scgptEval/pkg/snowflake"
	"time"
)

const secret = "moker_reader"

var (
	ErrorUserExist       = errors.New("用户已存在")
	ErrorUserNotExist    = errors.New("用户不存在")
	ErrorInvalidPassword = errors.New("密码错误")
)

// encryptPassword 密码加密
func encryptPassword(oPassword string) string {
	h := md5.New()
	h.Write([]byte(secret))
	return hex.EncodeToString(h.Sum([]byte(oPassword)))
}

// CheckUserExist 检查指定用户名的用户是否已存在
func CheckUserExist(username string) (err error) {
	if res := db.Where("username = ?", username).First(&models.User{}); res.RowsAffected > 0 {
		// 用户已存在
		return ErrorUserExist
	}
	return nil
}

func SignUp(username, password string) (userID int64, err error) {
	if err = CheckUserExist(username); err != nil {
		return
	}
	// 雪花算法生成uid
	userID = snowflake.GenID()
	// 构造一个user实例并插入
	user := &models.User{
		ID:          userID,
		Username:    username,
		Password:    encryptPassword(password), // 数据库不存储明文密码
		Amount:      0,
		FinalQATime: time.Now(), // 最后刷题时间可以不从当前时间为起始
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err = db.Create(user).Error; err != nil {
		return
	}
	return
}

// LogIn 用户登录
func LogIn(user *models.User) (aToken, rToken string, err error) {
	oPassword := user.Password
	if err = db.Where("username = ?", user.Username).First(user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = ErrorUserNotExist
		}
		return "", "", err
	}
	// 判断密码是否正确
	password := encryptPassword(oPassword)
	if password != user.Password {
		return "", "", ErrorInvalidPassword
	}
	// 通过登录验证，生成JWT token
	if aToken, rToken, err = jwt.GenToken(user.ID, user.Username); err != nil {
		// 过了登录验证但token生成失败，打log
		zap.L().Error("jwt.GenToken() failed",
			zap.Int64("userID", user.ID),
			zap.String("username", user.Username),
			zap.Error(err))
		return "", "", err
	}

	return aToken, rToken, nil
}

// UpdateUserName 更新用户名
func UpdateUserName(userID int64, newName string) (err error) {
	if err = CheckUserExist(newName); err != nil {
		return err
	}
	if err = db.Model(&models.User{}).
		Where("id = ?", userID).
		Update("username", newName).Error; err != nil {
		return err
	}
	return nil
}
