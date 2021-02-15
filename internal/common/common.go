/*
公共全局变量
*/
package common

import "time"

const (
	IsUseCa = false //mqtt服务器使用ca, 开发阶段不加密
)

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
	SecretKey   = "lianimicloud-secret"                          //salt for jwt
	IdentityKey = "userName"                                     //jwt key
	ExpireTime  = 30 * 24 * time.Hour                            //token expire time, one year
	PubAvatar   = "avatars/4d470ea0fe9f7e4812858f83e0d9daa8.jpg" //默认头像
	// update users set avatar = "avatars/4d470ea0fe9f7e4812858f83e0d9daa8.jpg" where avatar="";
	//  select username,avatar from users ;
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
	EXPIRESECONDS = 3600
	//上链图片下载过期时间,  秒
	PrivateEXPIRESECONDS = 300

	OSSUploadPicPrefix = "https://lianmi-ipfs.oss-cn-hangzhou.aliyuncs.com/"
	//例子 https://lianmi-ipfs.oss-cn-hangzhou.aliyuncs.com/msg/2020/11/29/id1/EF6B35D42C56273EF6D94B0DFC53C9C8
)

const (

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

	// 每签到2次奖励的gas
	AWARDGAS = 10000000

	// 提现后，钱包必须保留的最低余额
	BaseAmountLNMC = 1000

	//抽取佣金费率
	FEERATE float64 = 0.002
)

//eth相关
const (
	//使用Geth管理nonce
	UsingGethPendingNonceAt = true
)

//会员返佣比例相关， 比例是49%
const (
	VipBusinessUsername    = "id3"                                        //Vip会员的商户id, 暂定，上线后需要重新设定
	ChargeBusinessUsername = "id10"                                       // 接收服务费的商户id, 暂定，上线后需要重新设定
	ChargeReveiveWallet    = "0xc5a60be98722fef4266b08ac3dec3465dcf99fb5" // 接收服务费的商户钱包 , 暂定，上线后需要重新设定
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

//微信支付, 注意，要更换为连米的
const (
	WXAppID          = "wxf470593a8d5e0e12"
	WXApiKey         = "acf4990004d84488bd6cff67c0e15ade"
	WXMchID          = "1604757586"
	WXCertDateBase64 = "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"
)

//服务手续费费率， VIP用户减半， 非Vip半价
const (
	//VIP用户 免费的金额， 元
	RateFreeAmout = 30
	Rate          = 0.08
)
