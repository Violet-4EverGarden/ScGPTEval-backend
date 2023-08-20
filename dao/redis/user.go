package redis

import (
	"fmt"
	"time"
)

// CreateUserInfo 为指定id的用户创建对应的KeyUserInfoHash
func CreateUserInfo(userID int64, username string) (err error) {
	now := time.Now().Unix()
	userKey := KeyUserInfoHashPrefix + fmt.Sprint(userID)
	userInfo := map[string]interface{}{
		"username":      username,
		"final_qa_time": now,
	}
	if err = rdb.HMSet(userKey, userInfo).Err(); err != nil {
		return err
	}
	return nil
}

// UpdateUserName 更新缓存中用户信息的用户名
func UpdateUserName(userID int64, newName string) (err error) {
	if err = rdb.HSet(KeyUserInfoHashPrefix+fmt.Sprint(userID), "username", newName).Err(); err != nil {
		return err
	}
	return nil
}
