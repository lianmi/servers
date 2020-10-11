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

//阿里云
const (
	Endpoint   = "https://oss-cn-hangzhou.aliyuncs.com"
	AccessID   = "LTAI4G3o4sECdSBsD7rGLmCs"
	AccessKey  = "0XmB9tLOBLhmjIcM6CrBv2PHfnoDa8"
	RoleAcs    = "acs:ram::1230446857465673:role/ipfsuploader"
	BucketName = "lianmi-ipfs"
	//阿里云OSS临时token的过期时间, 默认是3600秒
	EXPIRESECONDS = 3600
)

const (
	//所有同步的时间戳数量
	TotalSyncCount = 9
)

const (

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

			三 、2-9号索引保留，10-以后 ，用于接收转账


	*/
	ETHINDEX      = 0 // 第0号叶子存储eth
	LNMCINDEX     = 1 //第1号存储LNMC代币
	WITHDRAWINDEX = 2 //第2号存储提现的LNMC代币

	//gas最低消耗
	GASLIMIT = 5000000

	// 每签到2次奖励的gas
	AWARDGAS = 10000000

	// 提现后，钱包必须保留的最低余额
	BaseAmountLNMC = 1000

	//抽取佣金费率
	FEERATE float64 = 0.002
)
