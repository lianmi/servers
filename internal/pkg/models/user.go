package models

/*
主键一定是唯一性索引，唯一性索引并不一定就是主键。

所谓主键就是能够唯一标识表中某一行的属性或属性组，一个表只能有一个主键，但可以有多个候选索引。因为主键可以唯一标识某一行记录，所以可以确保执行数据更新、删除的时候不会出现张冠李戴的错误。主键除了上述作用外，常常与外键构成参照完整性约束，防止出现数据不一致。数据库在设计时，主键起到了很重要的作用。

1.主键可以保证记录的唯一和主键域非空，数据库管理系统对于主键自动生成唯一索引，所以主键也是一个特殊的索引。
2. 一个表中可以有多个唯一性索引，但只能有一个主键。
3. 主键列不允许空值，而唯一性索引列允许空值。
4. 索引可以提高查询的速度。

其实主键和索引都是键，不过主键是逻辑键，索引是物理键，意思就是主键不实际存在，而索引实际存在在数据库中，主键一般都要建，主要是用来避免一张表中有相同的记录，索引一般可以不建，但如果需要对该表进行查询操作，则最好建，这样可以加快检索的速度。
*/

import (
	"time"

	"github.com/lianmi/servers/internal/pkg/models/global"
	"gorm.io/gorm"

	PbUser "github.com/lianmi/servers/api/proto/user"
)

//定义用户的数据结构, id+ Username 构成复合主键 , Mobile 是唯一索引
type UserBase struct {
	Username  string `gorm:"primarykey" json:"username" `                                //用户注册号，自动生成，字母 + 数字
	Password  string `json:"password" validate:"required"`                               //用户密码，md5加密
	Nick      string `json:"nick" validate:"required"`                                   //用户呢称，必填
	Gender    int    `form:"gender" json:"gender" binding:"required"`                    //性别
	Avatar    string `form:"avatar" json:"avatar,omitempty"`                             //头像url
	Label     string `form:"label" json:"label,omitempty" `                              //签名标签
	Mobile    string `gorm:"uniqueIndex" form:"mobile" json:"mobile" binding:"required"` //注册手机
	Email     string `form:"email" json:"email,omitempty" `                              //密保邮件，需要发送校验邮件确认
	Extend    string `form:"extend" json:"extend,omitempty" `                            //扩展字段
	AllowType int    `form:"allow_type" json:"allow_type"`                               //用户加好友枚举，默认是3
	UserType  int    `form:"user_type" json:"user_type" binding:"required"`              //用户类型 1-普通，2-商户

	//状态 当用户类型为普通用户时: 0-普通用户，非VIP 1-付费用户(购买会员) 2-封号
	//    当用户类型为商户时： 0-预注册 1-已审核 2-被封号
	State              int    `form:"state" json:"state"`
	TrueName           string `form:"true_name" json:"true_name,omitempty" `                       //用户实名
	IdentityCard       string `form:"identity_card" json:"identity_card" binding:"required"`       //身份证
	Province           string `form:"province" json:"province,omitempty" `                         //省份, 如广东省
	City               string `form:"city" json:"city,omitempty" `                                 //城市，如广州市
	County             string `form:"county" json:"county,omitempty" `                             //区，如天河区
	Street             string `form:"street" json:"street,omitempty" `                             //街道
	Address            string `form:"address" json:"address,omitempty" `                           //地址
	ReferrerUsername   string `form:"referrer_username" json:"referrer_username,omitempty" `       //推荐人，上线；介绍人, 账号的数字部分，app的推荐码就是用户id的数字
	BelongBusinessUser string `form:"belong_business_user" json:"belong_business_user,omitempty" ` //归属哪个商户，如果种子用户是商户的话，则一直都是这个商户
	VipEndDate         int64  `form:"vip_end_date" json:"vip_end_date,omitempty"`                  //VIP用户到期时间
	ECouponCardUsed    bool   `form:"e_coupon_card_used" json:"e_coupon_card_used,omitempty"`      //VIP7天体验卡
}

type User struct {
	global.LMC_Model

	UserBase
}

//BeforeCreate CreatedAt赋值
func (user *User) BeforeCreate(tx *gorm.DB) error {
	tx.Statement.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (user *User) BeforeUpdate(tx *gorm.DB) error {
	tx.Statement.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

// UserRole 用户身份结构体
type UserRole struct {
	UserName  string
	DeviceID  string
	UserRoles []*Role
}

func (user *UserBase) GetGender() PbUser.Gender {
	return PbUser.Gender(user.Gender)
}

func (user *UserBase) GetAllowType() PbUser.AllowType {
	return PbUser.AllowType(user.AllowType)
}

func (user *UserBase) GetUserType() PbUser.UserType {
	return PbUser.UserType(user.UserType)
}
