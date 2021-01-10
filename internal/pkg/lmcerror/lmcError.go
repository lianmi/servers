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

	AuthModNotRight = 2020501 //从设备无权踢主设备

	OrderModProductIDNotEmpty             = 2070101 //新的上架商品id必须是空的
	OrderModProductTypeError              = 2070102 //新的上架商品所属类型不正确
	OrderModProductExpireError            = 2070103 //过期时间小于当前时间戳
	OrderModAddProductUserTypeError       = 2070104 //用户不是商户类型，不能上架商品
	OrderModAddProductEmptyProductIDError = 2070105 //上架商品id不能为空
	OrderModAddProductNotOnSellError      = 2070106 //此商品没有上架过
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

	//订单模块
	OrderModProductIDNotEmpty:             "新的上架商品id必须是空的",
	OrderModProductTypeError:              "新的上架商品所属类型不正确",
	OrderModProductExpireError:            "过期时间小于当前时间戳",
	OrderModAddProductUserTypeError:       "过期时间小于当前时间戳",
	OrderModAddProductEmptyProductIDError: "上架商品id不能为空",
}

func ErrorMsg(errorCode int) string {
	if msg, ok := LmcErrors[errorCode]; ok {
		return msg
	} else {
		return "未知错误"
	}
}
