package controllers

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"net/http"
	"scgptEval/dao/mysql"
	"scgptEval/dao/redis"
	"scgptEval/models"
	"scgptEval/pkg/jwt"
	"strings"
)

// SignUp 用户注册
func SignUp(c *gin.Context) {
	var req struct {
		Username   string `json:"username" binding:"required"`
		Password   string `json:"password" binding:"required"`
		RePassword string `json:"re_password" binding:"required,eqfield=Password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		zap.L().Error("SignUp with invalid param", zap.Error(err))
		// 判断err是否为validator.ValidationErrors类型
		errs, ok := err.(validator.ValidationErrors)
		if !ok { // 若不是，则为ShouldBindJSON校验的 格式 或 数据类型 错误
			ResponseError(c, CodeInvalidParam)
			return
		}
		// 若是validator.ValidationErrors类型错误则进行翻译并响应错误
		ResponseErrorWithMsg(c, CodeInvalidParam, removeTopStruct(errs.Translate(trans)))
		return
	}
	// 2.业务处理: 在mysql插入新用户，同时在redis创建新key(存储用户信息的hash)
	userID, err := mysql.SignUp(req.Username, req.Password)
	if err != nil {
		if errors.Is(err, mysql.ErrorUserExist) {
			// 用户已存在错误
			ResponseError(c, CodeUserExist)
			return
		}
		// 其它注册失败错误
		zap.L().Error("mysql.SignUp failed",
			zap.String("username", req.Username),
			zap.String("password", req.Password), zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	if err := redis.CreateUserInfo(userID, req.Username); err != nil {
		// 已经过了mysql的重名验证，如果redis创建key失败需要打log记录一下
		zap.L().Error("redis.UpdateUserName() failed",
			zap.Int64("user_id", userID),
			zap.String("newName", req.Username), zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	// 3.返回响应
	ResponseSuccess(c, nil)
}

// Login 用户登录
func LogIn(c *gin.Context) {
	// 1.获取请求参数并校验
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		zap.L().Error("LogIn with invalid param", zap.Error(err))
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			ResponseError(c, CodeInvalidParam)
			return
		}
		ResponseErrorWithMsg(c, CodeInvalidParam, removeTopStruct(errs.Translate(trans)))
		return
	}
	// 2.登录并响应
	aToken, rToken, err := mysql.LogIn(&user)
	if err != nil {
		if errors.Is(err, mysql.ErrorUserNotExist) {
			// 用户不存在
			ResponseError(c, CodeUserNotExist)
			return
		} else if errors.Is(err, mysql.ErrorInvalidPassword) {
			// 密码错误
			ResponseError(c, CodeInvalidPassword)
			return
		}
		// 其它登录错误，打log；返回服务繁忙
		zap.L().Error("mysql.LogIn() failed", zap.String("username", user.Username), zap.Error(err))
		ResponseError(c, CodeServerBusy)
	}

	ResponseSuccess(c, gin.H{
		"accessToken":  aToken,
		"refreshToken": rToken,
		"user_id":      fmt.Sprint(user.ID), // 转string，避免前端数字失真
		"username":     user.Username,
	})
}

// ChangeName 用户改名
func ChangeName(c *gin.Context) {
	// 1.请求参数并校验
	var req struct {
		NewName string `json:"new_name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, CodeInvalidParam)
		return
	}
	// 获取当前请求的user_id
	userID, err := getCurrentUser(c)
	if err != nil {
		zap.L().Error("getCurrentUser() failed", zap.Error(err))
		ResponseError(c, CodeInvalidToken)
		return
	}
	// 2.更新数据库与缓存的用户信息
	if err = mysql.UpdateUserName(userID, req.NewName); err != nil {
		if errors.Is(err, mysql.ErrorUserExist) {
			// 用户名已存在
			ResponseError(c, CodeUserExist)
			return
		}
		// 其它错误，打log
		zap.L().Error("mysql.UpdateUserName() failed",
			zap.Int64("user_id", userID),
			zap.String("newName", req.NewName), zap.Error(err))
		ResponseError(c, CodeServerBusy)
	}
	if err = redis.UpdateUserName(userID, req.NewName); err != nil {
		// 已经过了mysql的重名验证，如果redis更新失败需要打log记录一下
		zap.L().Error("redis.UpdateUserName() failed",
			zap.Int64("user_id", userID),
			zap.String("newName", req.NewName), zap.Error(err))
		ResponseError(c, CodeServerBusy)
	}

	ResponseSuccess(c, gin.H{
		"username": req.NewName,
	})
}

// RefreshTokenHandler 刷新Access Token
func RefreshTokenHandler(c *gin.Context) {
	rt := c.Query("refresh_token") // 获取GET请求中携带的refresh_token参数
	// 默认使用Bearer Token
	authHeader := c.Request.Header.Get("Authorization")
	if authHeader == "" {
		ResponseErrorWithMsg(c, CodeInvalidToken, "请求头缺少Auth Token")
		c.Abort()
		return
	}
	// 按空格分割
	parts := strings.SplitN(authHeader, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		ResponseErrorWithMsg(c, CodeInvalidToken, "Token格式有误")
		c.Abort()
		return
	}
	aToken, rToken, err := jwt.RefreshToken(parts[1], rt)
	fmt.Println(err)
	c.JSON(http.StatusOK, gin.H{
		"access_token":  aToken,
		"refresh_token": rToken,
	})
}
