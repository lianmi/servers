/*
公共全局变量
*/
package common

const (
	
	SecretKey  = "lianimicloud-secret" //salt for jwt
	IdentityKey = "userName"
	ExpireTime = 365 * 24 * 3600       //token expire time, one year

)