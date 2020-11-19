package models

import (
	"time"

	"github.com/jinzhu/gorm"

	PbUser "github.com/lianmi/servers/api/proto/user"
)

/*
AllowAny(1)<允许任何人添加好友>

DenyAny(2)<拒绝任何人添加好友>

NeedConfirm(3)<添加好友需要验证,默认值>
*/
//定义用户的数据结构
type User struct {
	ID                 uint64 `gorm:"primary_key" form:"id" json:"id,omitempty"`                   //自动递增id
	CreatedAt          int64  `form:"created_at" json:"created_at,omitempty"`                      //创建时刻,毫秒
	UpdatedAt          int64  `form:"updated_at" json:"updated_at,omitempty"`                      //更新时刻,毫秒
	Username           string `json:"username" `                                                   //用户注册号，自动生成，字母 + 数字
	Password           string `json:"password" validate:"required"`                                //用户密码，md5加密
	SmsCode            string `json:"smscode" validate:"required"`                                 //校验码
	Nick               string `json:"nick" validate:"required"`                                    //用户呢称，必填
	Gender             int    `form:"gender" json:"gender" binding:"required"`                     //性别
	Avatar             string `form:"avatar" json:"avatar,omitempty"`                              //头像url
	Label              string `form:"label" json:"label,omitempty" `                               //签名标签
	Mobile             string `form:"mobile" json:"mobile" binding:"required"`                     //注册手机
	Email              string `form:"email" json:"email,omitempty" `                               //密保邮件，需要发送校验邮件确认
	AllowType          int    `form:"allow_type" json:"allow_type"`                                //用户加好友枚举，默认是3
	UserType           int    `form:"user_type" json:"user_type" binding:"required"`               //用户类型 1-普通，2-商户
	BankCard           string `form:"bank_card" json:"bank_card,omitempty" `                       //银行卡账号
	Bank               string `form:"bank" json:"bank,omitempty" `                                 //开户银行
	TrueName           string `form:"true_name" json:"true_name,omitempty" `                       //用户实名
	Deleted            int    `form:"deteled" json:"deteled"`                                      //软删除开关
	State              int    `form:"state" json:"state"`                                          //状态 0-普通用户，非VIP 1-付费用户(购买会员) 2-封号
	Extend             string `form:"extend" json:"extend,omitempty" `                             //扩展字段
	ContactPerson      string `form:"contact_person" json:"contact_person" binding:"required"`     //联系人
	ReferrerUsername   string `form:"referrer_username" json:"referrer_username,omitempty" `       //推荐人，上线；介绍人, 账号的数字部分，app的推荐码就是用户id的数字
	BelongBusinessUser string `form:"belong_business_user" json:"belong_business_user,omitempty" ` //归属哪个商户，如果种子用户是商户的话，则一直都是这个商户
	CreatedBy          string `form:"created_by" json:"created_by,omitempty"`                      //由谁创建， 分为注册或后台添加
	ModifiedBy         string `form:"modified_by" json:"modified_by,omitempty"`                    //最后由哪个操作员修改
}

//BeforeCreate CreatedAt赋值
func (user *User) BeforeCreate(scope *gorm.Scope) error {
	scope.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (user *User) BeforeUpdate(scope *gorm.Scope) error {
	scope.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
	return nil
}

// UserRole 用户身份结构体
type UserRole struct {
	UserName  string
	DeviceID  string
	UserRoles []*Role
}

func (user *User) GetGender() PbUser.Gender {
	return PbUser.Gender(user.Gender)
}

func (user *User) GetAllowType() PbUser.AllowType {
	return PbUser.AllowType(user.AllowType)
}

func (user *User) GetUserType() PbUser.UserType {
	return PbUser.UserType(user.UserType)
}
