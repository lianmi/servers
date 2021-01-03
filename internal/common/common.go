/*
公共全局变量
*/
package common

import "time"

//短信校验码一天总量100条
const (
	SMSCOUNT  = uint64(100000)
	SMSEXPIRE = 300
)

//redis
const (
	REDISTRUE  = 1
	REDISFALSE = 0
)

const (
	SecretKey   = "lianimicloud-secret"                           //salt for jwt
	IdentityKey = "userName"                                      //jwt key
	ExpireTime  = 30 * 24 * time.Hour                             //token expire time, one year
	PubAvatar   = "/avatars/4d470ea0fe9f7e4812858f83e0d9daa8.jpg" //默认头像

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

/*
阿里云

// RAM角色  ipfsuploader
// AccessID  = "LTAI4G3o4sECdSBsD7rGLmCs" //tempUploader@1230446857465673.onaliyun.com
// AccessKey = "0XmB9tLOBLhmjIcM6CrBv2PHfnoDa8"
// RoleAcs   = "acs:ram::1230446857465673:role/ipfsuploader"

*/
const (
	Endpoint = "https://oss-cn-hangzhou.aliyuncs.com"

	SuperAccessID        = "LTAI4FzZsweRdNRd3KLsUc2J"       //最高权限
	SuperAccessKeySecret = "W8a576pxtoyiJ7n8g4RHBFz9k5fF3r" //最高权限

	AccessID  = "LTAI4G8bgDLiaU9LfLyGQwgw"
	AccessKey = "uSI3XA0fk5FbLTVwhZ5bJNO1N1UAJA"
	RoleAcs   = "acs:ram::1230446857465673:role/lianmiipfswrite" //可读写

	BucketName = "lianmi-ipfs"
	//阿里云OSS临时token的过期时间, 默认是3600秒
	EXPIRESECONDS        = 3600
	PrivateEXPIRESECONDS = 300

	OSSUploadPicPrefix = "https://lianmi-ipfs.oss-cn-hangzhou.aliyuncs.com/"
	//例子 https://lianmi-ipfs.oss-cn-hangzhou.aliyuncs.com/msg/2020/11/29/id1/EF6B35D42C56273EF6D94B0DFC53C9C8
)

const (
	//所有同步的时间戳数量
	TotalSyncCount = 9

	//离线系统通知的最大同步数量
	OffLineMsgCount = 10
)

const (
	//助记词生成seed的加盐:
	// SeedPassword = "socialhahasky"
	SeedPassword = "" //TODO 暂时不动，等准备上线后再统一改

	//运营方的助记词
	// MnemonicServer = "job gravity goose boring filter lyrics source giant throw dismiss film emotion margin depend ostrich peanut exist version unfold logic cause protect section drama"
	MnemonicServer = "element urban soda endless beach celery scheme wet envelope east glory retire"

	/*

		约定:
			一. 第0号索引派生的负责存储eth，并以此地址给用户派发gas,
				account Address:  0xC50Fe56057B5D6Ab4b714C54d72C8e3018975D5D
				Private key0 of account in hex: 387153a31bf48456fed325e1a5be9e17c1c87e00cd5bac8721db3b0cc79a1d74
				Public key0 of account  in hex: 906abda2050da89224a1d9e13d64f38b14de1f0f46b2043354f7032d2ba1ebdb0b7a88bb40700ce2a0deca6e9e28524f2bff3f63654dc6e94561ed5babedf5eb

			二.  第1号索引派生的负责存储LNMC，当用户充值的时候，以此地址给用户派发LNMC
				account1 Address:  0x1826654168d449004794C1d6F092d5E3F0F8365A
				Private key1 of account in hex: 387153a31bf48456fed325e1a5be9e17c1c87e00cd5bac8721db3b0cc79a1d74
				Public key1 of account  in hex: 906abda2050da89224a1d9e13d64f38b14de1f0f46b2043354f7032d2ba1ebdb0b7a88bb40700ce2a0deca6e9e28524f2bff3f63654dc6e94561ed5babedf5eb

			三、第2号存储提现的LNMC代币

			四、第3号索引派生的负责接收会员费的代币LNMC，当用户购买会员时，以此地址作为接收地址

			五、4-9号索引保留，10-以后 ，用于接收转账


	*/
	ETHINDEX        = 0 // 第0号叶子存储eth
	LNMCINDEX       = 1 //第1号存储LNMC代币
	WITHDRAWINDEX   = 2 //第2号存储提现的LNMC代币
	MEMBERSHIPINDEX = 3 //第3号负责接收会员费的LNMC代币

	//1个eth
	ETHER = 100000000000000000

	//gas最低消耗
	GASLIMIT = 5000000

	// 每签到2次奖励的gas
	AWARDGAS = 10000000

	// 提现后，钱包必须保留的最低余额
	BaseAmountLNMC = 1000

	//抽取佣金费率
	FEERATE float64 = 0.002
)

//会员返佣比例相关， 比例是49%
const (
	VipBusinessUsername    = "id3"  //Vip会员的商户id, 暂定，上线后需要重新设定
	ChargeBusinessUsername = "id10" // 接收服务费的商户id, 暂定，上线后需要重新设定
	CommissionOne          = float64(0.3)
	CommissionTwo          = float64(0.1)
	CommissionThree        = float64(0.09)
)

//订单相关
const (
	//订单完成后，买家发送确认收货
	OrderTransferForDone = int32(1)

	//订单由买家发起撤单申请，商户同意撤单并退款
	OrderTransferForCancel = int32(2)
)

//支付宝
const (
	AlipayAppId = "2021002115683928"

	//应用私钥
	AppPrivateKey = "MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCbzmGCUADocil228GT7edg5F2NaCNSCJx8Jn+LaaT+onzAMal1A4pCg0pUMAMWfdbMYP336an/6Lp0V2K/c0Lzr6o5KGpZJ93pgrHAo5PGjCt/Dw9j/WPGX3kHQMYsGHG9iFZx85v7JwcRo5u5wrO126YYC+JUNoPq7FVv9rcC0Amw61MycYkL95irmivIGI333TJdMdsv52s984yxNeBuTbBEjprn9Zz//joIaP1Tl1HC/NZmFxcCYKMz+D+bl1w9oaP8+jYPwW3EMYWGmAR3mw9EZVAtfwFwta8OYzpo0MFVeRMwpLBFpyDt8+l307MnVaR12dtOvhwLl2vjMRTfAgMBAAECggEBAJR8eZ1xlYvx0OZ/xNqwbkR/H1F2n8K8hjYjkoZQ5nfubynTqoXkG84Lxbi6ERdMUntxLFkqjWNgbuIVrfx7YqFPFtFmXQQe5HR4o+LNgjZEu+dZePd4M7CIqJVq+/JmUW+qEYiD/HG83hXHcM/2aMK2VHKyUL6lPc+T8FDGNeAs223nuat9zSrsdDYph+zyLPHyZYPQuhojnyntXn4BTh4NPfF/6sIyAi2d4tLMkk6LV7MkH1PpUWcAjN+BxWtKAB8vUbvwJg0vbA8y0Phq5Np1t1WNyDgTPgwH73RFVZW/7Lg09zwYRDCMm36v2JuvZvVw3mK5YX2aU52kPIlLETECgYEAy/nwds398Nch88pgJ0Bbm64YPK67o6V2kCZzLl1JLOHCdnzrkrmt2XfOWfJ9T2CoCFyFtXNonUqbChcspijaP4XPx5EoY0uKZnmUQxPz9YeapSzXuG9QFv8EkVb7ZwDGkTHCcM0vDrCwS/XVNVzpXgu5+pl6BSMIO46mcJMxhe0CgYEAw4tOm8AC/an6ftMN364zS1hrx/zag4lNcgjNFlTovdCeW9RJQo4GKnHuNiPodibzCC/lL8CmRcbyFiQBqH6WJMV2dHztpUwwbr3RiM1OTItf9yZHiasaBy3qONdU5fo+j1x5aNAk5ItDZX6DOWo56UY8XkHDGW964fSeC/jFLHsCgYAQ0PNFKChmYaYX7jhNJB4pUIoI/rLThAGpUrIuQVyWCaq5kATv3MT7Z8goXDh+gc54mgAf/HrEdPEhPNXegQG1OPfvUQVOYlzvo9hYS13SgTJ7qZ3DQ9ILg0zCGrSxQjwcnkiUeiYGBQUTzhmcw6MtsLPNeDe6ErBMEK+iGlB75QKBgHHu9CFJkjSMWniUrku66wYmgb4ndIYZdPdRa3VsiaM3L12f5gOSTsNiWIJRD7vv28DUbzwQipCzZxBBcHnlL8RDDU64D5s1Ni8ACFsmDE4LEyIkup/bArJWLVdrF3tcACF1pwPL6wMCpYU4XmsQmqdxlfDxbiSe0MFgzsl47CGLAoGAYPcIiLXuFreNL7qMvGubPmi0vPpquY9BAKCo6aogBUevxikcRiRr2zFxeaVcGYSCdTZGEuNa4/E39jbAys55WryeP1bttU9SiahZ+l1ktKlYS80arFn6ScMwiHMAWXqHNHhD5jdo9Ivuf1q6l9ISh7drRY43rw77652BvlmIHjM="

	//应用公钥
	AppPublicKey = "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAm85hglAA6HIpdtvBk+3nYORdjWgjUgicfCZ/i2mk/qJ8wDGpdQOKQoNKVDADFn3WzGD99+mp/+i6dFdiv3NC86+qOShqWSfd6YKxwKOTxowrfw8PY/1jxl95B0DGLBhxvYhWcfOb+ycHEaObucKztdumGAviVDaD6uxVb/a3AtAJsOtTMnGJC/eYq5oryBiN990yXTHbL+drPfOMsTXgbk2wRI6a5/Wc//46CGj9U5dRwvzWZhcXAmCjM/g/m5dcPaGj/Po2D8FtxDGFhpgEd5sPRGVQLX8BcLWvDmM6aNDBVXkTMKSwRacg7fPpd9OzJ1WkddnbTr4cC5dr4zEU3wIDAQAB"

	//支付宝公钥 - 巨商
	AlipayPublicKey = "MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAikjYvpiSlrLUeI03brTCtFPAzcPA62JSAI6ytmdXxEJmanXqDA2W1POE72wBTjJoQatmVzPT8+WL6DH94PtjRQx0zFnEOiOKWdGLXp95v+YpNVaPA6SlmZ3XlBRWzrpCMcuejcjk5QYy5VMcQybFVepWHwNirpMjK0qJ5CEy5camNZgAD+kFkeXxg/RX3oWE9MP4yUjObJniUYZwGnTRcTTVnBi2FyGrfmHQSGnWEm/F5Q4jk8/vnFDdBnSLRYRySyEZAHRqGJse4U9VLkhOWBAYKfvqFHU8UbbH/Z/fhl31p7xA/eG9hHyoFJagW5JDco6cDC0rP5m00js+B5BJtQIDAQAB"

	ServerDomain = "https://api.lianmi.cloud"
)

//服务手续费费率， VIP用户减半， 非Vip半价
const (
	//VIP用户 免费的金额， 元
	RateFreeAmout = 30
	Rate          = 0.08
)
