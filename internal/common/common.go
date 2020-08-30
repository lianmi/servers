/*
公共全局变量
*/
package common

import "time"

const (
	SecretKey   = "lianimicloud-secret"                                         //salt for jwt
	IdentityKey = "userName"                                                    //jwt key
	ExpireTime  = 30 * 24 * time.Hour                                           //token expire time, one year
	PubAvatar   = "https://zbj-bucket1.oss-cn-shenzhen.aliyuncs.com/avatar.JPG" //默认头像 TODO 要换为自己的OSS

)

const (

	//允许任何人添加好友
	AllowAny int = 1

	//拒绝任何人添加好友
	DenyAny int = 2

	//添加好友需要验证,默认值
	NeedConfirm int = 3
)

const (
	UserBlocked = 2 //用户被封禁
)

const (
	//群成员上限为600人
	PerTeamMembersLimit int = 600

	//每个网点用户只允许最多建群数量
	MaxTeamLimit int = 50

	//一天最多拉多少人入群
	OnedayInvitedLimit = 50
)

const (
	//阿里云OSS临时token的过期时间, 默认是3600秒
	EXPIRESECONDS = 3600
)

const (
	//所有同步的时间戳数量
	TotalSyncCount = 6
)
