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
