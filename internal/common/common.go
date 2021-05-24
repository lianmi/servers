/*
公共全局变量
*/
package common

import "time"

const (
	IsUseCa = true //mqtt服务器使用ca, 开发阶段不加密
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
	ETH   = 1
	ETHER = 1000000000000000000

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

	//服务端的订单附件加密上链协商私钥
	ServerPrivateKey = "b3e7a1b8fa4d7e958eecbc72f8d95e667889b787a83e8be576523aefb82ba507"

	// 服务端的订单附件加密上链协商公钥 在UI写死
	ServerPublicKey = "36c02735d5500646e48a10da640713dcc3382347ab7ee2fc15244bbe38270178"
)

//服务手续费费率， VIP用户减半， 非Vip半价
const (
	//VIP用户 免费的金额， 元
	RateFreeAmout = 30
	Rate          = 0.08
)

// 微信支付相关
const (
	WechatPay_appID    = "wx9dff85e1c3a3b342"                       // 服务商appid
	WechatPay_mchId    = "1608460662"                               // 服务商商户id
	WechatPay_serierNo = "7D44E512E73027719552E38F0DE879D1A76C2B87" // 服务商证书序列号，更换证书后需要修改
	WechatPay_apiV3Key = "LianmiLianmiLianmiLianmicloud508"         // 服务商 apikey
	Wechat_pkContent   = `-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQDMoaqW4aXuuRcR
24TSzoUCbzGXf2DP/aI8dFoXH3/kNF7H/GejfZ/TuvM3R5oOWsN0BXdmY1hdO4q3
sdiLtpSus6SCGi45iw/v+3JJa/u2pDqQbN4ZTNZlOSZjfrlAmUqG1G47cg7J/4p6
/RZEcEb7WtC2ETv/EE0Ge7lqUzfycJ/5EQTpTUg8mqcYcF3XC3GpR+uEaATc15zs
elJwPito7Th+fdrC2CaNQYvsxqTzjD6zaKnfJGTp6OccloGqn15bzoCpSsMSYFL/
NsjNM6Gfh9ANKmOe4MqY2o+6hreSUnSrjPllF9bIzR/yr9LraZdh/EumBjpF/Cb1
3MfdH3MpAgMBAAECggEBAIU/ZGi5aKZpSfdr3TK0HfJ223EOFcl6HBGHpj5WWZ4M
6AcLeaUBIXjqzIMbkdp1Cb7b7GL0n86d/fcdzKc1bd3QxnedeqonvmoDbukWcqL8
j9IJwhnxac4iB7hUBWdmKhxf6aO14qFwUAlEEiLghagY+70CvfGZ+L4XBKaSp+Sq
fG56dYpPC/Gch5BYf3pCStW9G/V9e6wFR5DGRNC52Svw2pMJ4pQcdHqfzKmlJYXR
cL8V8cXnxiTIiFDuYWiNAdEOeausc2UXJEg3cIfy9YcrQi2mT+twsVswoIYBmegT
P70XdjlVSPwZBGPIcEoTxPBPU+inRFz273l+pa5OKg0CgYEA+1Wl3fqRCpfcbQTR
ngPD8gXKJ2TXIuqlwEz1RKWNu/3tNpqQ7b/Hsn6Q0dEbRuxDX2R9O8jcZUHLVnQP
O/Q4tmHcuxWC1PdMavuI1U/TwrBebxT4W/Fduhc3aF30Vk8N5Y0nXRpccdus4vce
0qa0GdFmbPhaibXcxB/HAMKPEdMCgYEA0G4VaSSc/1RqeeYf57gO2FfXI3J0gdX4
d642PNvMxK/DsrTZlPeZrdNcOcvl670ubRNPpz6pfcdac9x0wgVIeU4ZqN2xGqlt
Pp+RFI2Ob9dV5/ANVO0aFIhFqeqTR8Bus7KSoTx2t75pvSq2VNGb7GL8mr3lfQH1
6FTkwRfdjZMCgYBl1LDMfG3xpc/IV/B6Hjpwv8nFJkVIP1wCyuuA8ba4WUyYGA3q
Vg6aEk+owxlTJfyyFKvs4hfx6rNxBrr5ZpznwETHhBKrKLtMiTdKffplYkIQraVm
0ydPc4KehZqusX8G56bwQPL9qqyklM1nOeW0pDPkqMc+DnIxAFMHysxewwKBgDOd
HxYzZ+FeoSNglkQGcz6lufPgMvO37diNPochkvqd3+NQH5VhHyBJd8wkLuKKrYV7
Q71Rqh0okcChNhSZxFGtwnLruyC0FgZs8ztYto4BkBdofZSrRksRV9b07NXW1FMR
hHgDBg8ISxz6B77HTUpjVNRo8/xZ0PBgnWknpMibAoGBAJUmMxGRFdO3WOGo80Od
kKVN98lq4ZUdA2zPYUMxkRDHS2u3aLshEA2vnseqKHabV8M9UXyvUK9uH8KT+8rn
3jE1PsaOyc+RPMC+jobPG8FJOZRYV5lDDAlLt/g8QKWUBx+jNaFQijVidWcjPjND
2G4qLkNFV/7SB+31YvVqwB7w
-----END PRIVATE KEY-----
`
	// app 的绑定的appid
	WechatPay_SUBAppid_LM = "wx239c33b9be7cd047"
)

//彩票中心相关
const (
	Fucai_Topic = "lianmi/cloud/lottety-center-guangdong" //福彩中心
	Tiyu_Topic  = "lianmi/cloud/lottety-center-guangdong" //体育彩票中心
)

const (
	WechatLoginAppid  = "wx239c33b9be7cd047"               // 微信登陆 appid
	WechatLoginSecret = "431e4bec654f13edb7be0f018214244b" // 微信密钥
)
