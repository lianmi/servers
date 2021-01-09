package lmcerror

const (
	ProtobufUnmarshalError = 10001 //协议解包错误
	RedisError             = 10002 //Redis错误
	DataBaseError          = 10003 //系统数据库错误

	UserModUpdateProfileError = 2010201 //修改用户资料出错
	UserModUpdateStoreError   = 2010202 //修改商户店铺资料出错

	UserModMarkTagParamError          = 2010501 //参数错误, 不能给自己打标
	UserModMarkTagParamNorFriendError = 2010502 //参数错误, 不能给非好友打标
	UserModMarkTagAddError            = 2010503 //添加打标出错
	UserModMarkTagRemoveError         = 2010504 //移除打标出错

)

var LmcErrors = map[int]string{
	ProtobufUnmarshalError: "协议解包错误",

	RedisError: "缓存错误",

	DataBaseError: "系统数据库错误",

	//用户模块
	UserModUpdateProfileError: "修改用户资料出错",
	UserModUpdateStoreError:   "修改商户店铺资料出错",

	UserModMarkTagParamError:          "参数错误, 不能给自己打标",
	UserModMarkTagParamNorFriendError: "参数错误, 不能给非好友打标",
	UserModMarkTagAddError:            "添加打标出错",
	UserModMarkTagRemoveError:         "移除打标出错",
}

func ErrorMsg(errorCode int) string {
	if msg, ok := LmcErrors[errorCode]; ok {
		return msg
	} else {
		return "未知错误"
	}
}
