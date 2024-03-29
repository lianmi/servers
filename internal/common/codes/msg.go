package codes

//MsgFlags 错误信息
var MsgFlags = map[int]string{
	SUCCESS:                  "ok",
	ERROR:                    "fail",
	InvalidParams:            "请求参数错误",
	ErrExistTag:              "已存在该标签名称",
	ErrNotExistTag:           "该标签不存在",
	ErrNotExistArticle:       "该文章不存在",
	ErrAuthCheckTokenFail:    "Token鉴权失败",
	ErrAuthCheckTokenTimeout: "Token已超时",
	ErrAuthToken:             "Token生成失败",
	ErrAuth:                  "Token错误",
	PageNotFound:             "Page not found",
	ErrNotDigital:            "非数字",
	ErrWrongSmsCode:          "短信校验码错误",
	ErrExistUser:             "用户名已存在",
	ErrExistMobile:           "手机号已存在",
	ErrNotRegisterMobile:     "此手机号没注册",
	ErrNotFoundInviter:       "邀请码对应的用户不存在",
}

//GetMsg 获取错误信息
func GetMsg(code int) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}

	return MsgFlags[ERROR]
}
