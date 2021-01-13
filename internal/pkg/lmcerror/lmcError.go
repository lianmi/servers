package lmcerror

const (
	InternalServerError    = 500   //系统错误
	ProtobufUnmarshalError = 10001 //协议解包错误
	RedisError             = 10002 //Redis错误
	DataBaseError          = 10003 //系统数据库错误
	DecodingHexError       = 10004 // 反hex错误
	ParseAttachError       = 10005 // 解释Attach错误
	Base64DecodingError    = 10006 // base64解码错误
	AttachBodyTypeError    = 10007 // Attach里的类型错误
	UnkownOrderTypeError   = 10008 // 订单状态未定义
	ParamError             = 10009 // 参数错误

	NormalUsernameIsEmptyError   = 10010 //普通用户账号不能为空
	BusinessUsernameIsEmptyError = 10011 //商户用户账号不能为空

	UserIsBlockedError               = 10012 //此用户已经被封号
	BusinessUserIsBlockedError       = 10013 //此商户已经被封号
	TargetUserIsNotBusinessTypeError = 10014 // 目标用户不是商户类型
	GenerateSignatureUrlError        = 10015 // 生成临时凭证错误
	UserNotExistsError               = 10016 // 用户不存在或未注册

	UserModUpdateProfileError = 2010201 //修改用户资料出错
	UserModUpdateStoreError   = 2010202 //修改商户店铺资料出错

	IsNotFriendError       = 2020101 //对方用户不是当前用户的好友
	IsBlackUserError       = 2020102 //已被对方拉黑， 不能加好友
	AddFriendError         = 2020103 //加好友出错
	AddFriendExpireError   = 2020104 //好友请求已超过有效期
	FollowIsBlackUserError = 2020105 //用户已被对方拉黑,不能关注

	TeamStatusError                        = 2040101 //群组状态错误
	TeamMembersLimitError                  = 2040102 //用户拥有的群的总数量是否已经达到上限
	TeamIsNotExistsError                   = 2040103 //群组不存在或未审核通过
	TeamOneDayInviteLimitError             = 204014  //一天最多只能邀请50人入群
	TeamOperatorNotOwnerError              = 204015  //当前操作者不是群主或管理员
	TeamOperatorNotRightDeleteMemberrError = 204016  //当前操作者无权删除群成员
	TeamAlreadyMemberError                 = 204017  //用户已是群成员
	InviteTeamMembersError                 = 204018  //校验用户是否曾经被人拉入群
	AddTeamUserError                       = 204019  //增加群成员错误
	UserIsAlreadyTeammemberError           = 204020  //用户已经是群成员
	TeamVerifyTypePrivateErroe             = 204021  // 此群仅限邀请加入
	TeamUserIsNotExists                    = 204022  // 此用户不是群成员
	OwnerLeaveTeamError                    = 204023  // 管理员或群主不能退群，必须由群主删除
	DeleteTeamUserError                    = 204024  //移除群成员错误
	ManagerNotRightMuteAllTeamUsersError   = 204025  //管理员无权设置全体禁言
	NorlamNotRightMuteTeamUsersError       = 204026  //其它成员无权设置禁言
	NorlamNotRightMuteTimeTeamUsersError   = 204027  //其它成员无权设置禁言时长
	NorlamNotRightSetProfileTeamUsersError = 204028  //其它成员无权设置群成员资料
	NorlamNotRightAuditError               = 204029  //其它成员无权审核用户入群申请

	UserModMarkTagParamError          = 2010501 //参数错误, 不能给自己打标
	UserModMarkTagParamNorFriendError = 2010502 //参数错误, 不能给非好友打标
	UserModMarkTagAddError            = 2010503 //添加打标出错
	UserModMarkTagRemoveError         = 2010504 //移除打标出错

	AuthModNotRight = 2020501 //从设备无权踢主设备

	OrderModProductIDNotEmpty        = 2070101 //新的上架商品id必须是空的
	OrderModProductTypeError         = 2070102 //新的上架商品所属类型不正确
	OrderModProductExpireError       = 2070103 //过期时间小于当前时间戳
	OrderModAddProductUserTypeError  = 2070104 //用户不是商户类型，不能上架商品
	ProductIDIsEmptError             = 2070105 //上架商品id不能为空
	OrderModAddProductNotOnSellError = 2070106 //此商品没有上架过
	BuyUserIsEmptyError              = 2070107 //买家账号为空
	BusinessUserIsEmptyError         = 2070108 //商家账号为空
	OPKEmptyError                    = 2070109 //商家OPK为空

	RegisterPreKeysArrayEmptyError      = 2090101 //一次性公钥数量为零
	RegisterPreKeysNotBusinessTypeError = 2090102 //用户不是商户类型，不能上传OPK

	GetPreKeyOrderIDEmptyProductIDError = 2090201 //商品id不能为空

	ProductExpireError = 2090202 // 商品有效期过期

	OrderMsgTypeError       = 2090301 //消息类型是非订单类型
	OrderIDIsEmptyError     = 2090302 //订单id不能为空
	QueryVipPriceError      = 2090303 //获取VIP价格错误
	OrderTotalAmountError   = 2090304 //订单金额错误
	OrderIDIsNotExistsError = 2090305 //订单ID不存在
	OrderIDNotBelongToError = 2090306 //此订单id不属于此商户

	OrderStatusConfirmIsDoneError        = 2090501 //此订单已经确认收货,不能再更改其状态
	OrderStatusIsCancelError             = 2090502 //此订单已撤单,不能再更改其状态
	OrderStatusIsPayingError             = 2090503 //此订单当前状态为支付中, 不能更改状态
	OrderStatusIsRefusedError            = 2090504 //此订单已拒单,不能再更改其状态
	OrderStatusIsUrgedError              = 2090505 //此订单当前状态为买家催单, 只能催一次
	OrderStatusIsBuyerError              = 2090506 //买家不能接单
	OrderStatusPayedError                = 2090507 //完成支付之后不能修改订单内容及金额
	OrderStatusBusinessChangeError       = 2090508 //当前状态处于完成订单状态, 不能更改为其它
	OrderStatusChangeConfirmError        = 2090509 //当前状态处于完成订单状态, 只能选择确认
	OrderStatusNotPayError               = 2090510 //买家确认收货, 但是未完成支付
	OrderStatusCannotChangetoPayingError = 2090511 //此状态不能由用户设置为支付中
	OrderStatusOnceUrgedError            = 2090512 //买家催单, 只能催一次
	OrderStatusVipExpeditedError         = 2090513 //VIP用户才允许加急

	PreKeyGetCountError = 2090601 // 只有商户才能查询OPK存量

	WalletTranferError = 2100101 //钱包转账错误
)

var LmcErrors = map[int]string{
	ProtobufUnmarshalError: "协议解包错误",

	RedisError:           "缓存错误",
	DataBaseError:        "系统数据库错误",
	DecodingHexError:     "Hex解码错误",
	ParseAttachError:     "解释Attach错误",
	Base64DecodingError:  "Base64解码错误",
	AttachBodyTypeError:  "Attach里的类型错误",
	UnkownOrderTypeError: "订单状态未定义",
	ParamError:           "参数错误",
	UserNotExistsError:   "用户不存在或未注册",

	NormalUsernameIsEmptyError:   "普通用户账号不能为空",
	BusinessUsernameIsEmptyError: "商户用户账号不能为空",

	UserIsBlockedError:               "此用户已经被封号",
	TargetUserIsNotBusinessTypeError: "目标用户不是商户类型",

	GenerateSignatureUrlError: "生成临时凭证错误",

	//用户模块
	UserModUpdateProfileError: "修改用户资料出错",
	UserModUpdateStoreError:   "修改商户店铺资料出错",

	UserModMarkTagParamError:          "参数错误, 不能给自己打标",
	UserModMarkTagParamNorFriendError: "参数错误, 不能给非好友打标",
	UserModMarkTagAddError:            "添加打标出错",
	UserModMarkTagRemoveError:         "移除打标出错",

	IsNotFriendError:       "对方用户不是当前用户的好友",
	IsBlackUserError:       "已被对方拉黑， 不能加好友",
	AddFriendError:         "加好友出错",
	AddFriendExpireError:   "好友请求已超过有效期",
	FollowIsBlackUserError: "用户已被对方拉黑,不能关注",

	TeamStatusError:                        "群组状态错误",
	TeamMembersLimitError:                  "用户拥有的群的总数量是否已经达到上限",
	TeamIsNotExistsError:                   "群组不存在或未审核通过",
	TeamOneDayInviteLimitError:             "一天最多只能邀请50人入群",
	TeamOperatorNotOwnerError:              "当前操作者不是群主或管理员",
	TeamOperatorNotRightDeleteMemberrError: "当前操作者无权删除群成员",
	TeamAlreadyMemberError:                 "用户已是群成员",
	InviteTeamMembersError:                 "校验用户是否曾经被人拉入群",
	AddTeamUserError:                       "增加群成员错误",
	UserIsAlreadyTeammemberError:           "用户已经是群成员",
	TeamVerifyTypePrivateErroe:             "此群仅限邀请加入",
	TeamUserIsNotExists:                    "此用户不是群成员",
	OwnerLeaveTeamError:                    "管理员或群主不能退群，必须由群主删除",
	DeleteTeamUserError:                    "移除群成员错误",
	ManagerNotRightMuteAllTeamUsersError:   "管理员无权设置全体禁言",
	NorlamNotRightMuteTeamUsersError:       "其它成员无权设置禁言",
	NorlamNotRightMuteTimeTeamUsersError:   "其它成员无权设置禁言时长",
	NorlamNotRightSetProfileTeamUsersError: "其它成员无权设置群成员资料",
	NorlamNotRightAuditError:               "其它成员无权审核用户入群申请",

	//订单模块
	OrderModProductIDNotEmpty:        "新的上架商品id必须是空的",
	OrderModProductTypeError:         "新的上架商品所属类型不正确",
	OrderModProductExpireError:       "过期时间小于当前时间戳",
	OrderModAddProductUserTypeError:  "过期时间小于当前时间戳",
	ProductIDIsEmptError:             "商品id不能为空",
	OrderModAddProductNotOnSellError: "此商品没有上架过",
	BuyUserIsEmptyError:              "买家账号为空",
	BusinessUserIsEmptyError:         "商家账号为空",

	RegisterPreKeysArrayEmptyError:      "一次性公钥数量为零",
	RegisterPreKeysNotBusinessTypeError: "用户不是商户类型，不能上传OPK",

	GetPreKeyOrderIDEmptyProductIDError: "商品id不能为空",
	ProductExpireError:                  "商品有效期过期",

	OrderMsgTypeError:       "消息类型是非订单类型",
	OrderIDIsEmptyError:     "订单id不能为空",
	QueryVipPriceError:      "获取VIP价格错误",
	OrderTotalAmountError:   "订单金额错误",
	OrderIDIsNotExistsError: "订单ID不存在",
	OrderIDNotBelongToError: "此订单id不属于此商户",

	OrderStatusConfirmIsDoneError:        "此订单已经确认收货,不能再更改其状态",
	OrderStatusIsCancelError:             "此订单已撤单,不能再更改其状态",
	OrderStatusIsPayingError:             "此订单当前状态为支付中, 不能更改状态",
	OrderStatusIsRefusedError:            "此订单已拒单,不能再更改其状态",
	OrderStatusIsUrgedError:              "此订单当前状态为买家催单中, 只能催一次",
	OrderStatusIsBuyerError:              "买家不能接单",
	OrderStatusPayedError:                "完成支付之后不能修改订单内容及金额",
	OrderStatusBusinessChangeError:       "当前状态处于完成订单状态, 不能更改为其它",
	OrderStatusChangeConfirmError:        "当前状态处于完成订单状态, 只能选择确认",
	OrderStatusNotPayError:               "买家确认收货, 但是未完成支付",
	OrderStatusCannotChangetoPayingError: "此状态不能由用户设置为支付中",
	OrderStatusOnceUrgedError:            "买家催单, 只能催一次",
	OrderStatusVipExpeditedError:         "VIP用户才允许加急",

	PreKeyGetCountError: "只有商户才能查询OPK存量",
}

func ErrorMsg(errorCode int) string {
	if msg, ok := LmcErrors[errorCode]; ok {
		return msg
	} else {
		return "未知错误"
	}
}
