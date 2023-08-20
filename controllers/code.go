package controllers

type ResCode int64

// 错误码及其映射内容
const (
	CodeSuccess ResCode = 1000 + iota
	CodeInvalidParam
	CodeUserExist
	CodeUserNotExist
	CodeInvalidPassword
	CodeRecordExist
	CodeQuizNotExist
	CodeServerBusy

	CodeInvalidToken
	CodeNeedLogin
	CodeInvalidAuthFormat
)

var codeMsgMap = map[ResCode]string{
	CodeSuccess:         "success",
	CodeInvalidParam:    "请求参数错误",
	CodeUserExist:       "用户名已存在",
	CodeUserNotExist:    "用户名不存在",
	CodeInvalidPassword: "用户名或密码错误",
	CodeRecordExist:     "提交记录已存在",
	CodeQuizNotExist:    "题目不存在",
	CodeServerBusy:      "服务繁忙",

	CodeNeedLogin:         "未登录",
	CodeInvalidToken:      "无效Token",
	CodeInvalidAuthFormat: "认证格式有误",
}

func (c ResCode) Msg() string {
	msg, ok := codeMsgMap[c]
	if !ok {
		msg = codeMsgMap[CodeServerBusy]
	}
	return msg
}
