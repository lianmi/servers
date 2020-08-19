/*
公共全局变量
*/
package common

import "time"

const (
	SecretKey   = "lianimicloud-secret"                                         //salt for jwt
	IdentityKey = "userName"                                                    //jwt key
	ExpireTime  = 30 * 24 * time.Hour                                           //token expire time, one year
	PubAvatar   = "https://zbj-bucket1.oss-cn-shenzhen.aliyuncs.com/avatar.JPG" //默认头像

)

const (

	//允许任何人添加好友
	AllowAny int = 1
	//拒绝任何人添加好友
	DenyAny int = 2
	//添加好友需要验证,默认值
	NeedConfirm int = 3
)
