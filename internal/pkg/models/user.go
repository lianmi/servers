package models

import (
	"time"

	"github.com/jinzhu/gorm"

	pb "github.com/lianmi/servers/api/proto/user"
)

/*
	ID         int       `gorm:"primary_key" json:"id"`
	CreatedAt  time.Time `json:"created_on"`
	UpdatedAt time.Time `json:"modified_on"`
	Username   string    `json:"username" validate:"required"`
	Password   string    `json:"password" validate:"required"`
	Avatar     string    `json:"avatar"`
	UserType   int       `json:"user_type"`
	Deleted    int       `json:"deteled"`
	State      int       `json:"state"`
	CreatedBy  string    `json:"created_by"`
	ModifiedBy string    `json:"modified_by"
*/
//定义用户的数据结构
type User struct {
	ID                uint64      `gorm:"primary_key" form:"id" json:"id,omitempty"`                         //自动递增id
	CreatedAt         time.Time   `form:"created_at" json:"created_at,omitempty"`                            //创建时刻
	UpdatedAt         time.Time   `form:"updated_at" json:"updated_at,omitempty"`                            //更新时刻
	Mobile            string      `form:"mobile" json:"mobile" binding:"required"`                           //注册手机
	Username          string      `json:"username" validate:"required"`                                      //用户注册号，自动生成，字母 + 数字
	Password          string      `json:"password" validate:"required"`                                      //用户密码，md5加密
	Gender            pb.Gender   `form:"gender" json:"gender" binding:"required"`                           //性别
	Avatar            string      `form:"avatar" json:"avatar,omitempty"`                                    //头像url
	Label             string      `form:"label" json:"label,omitempty" `                                     //签名标签
	Email             string      `form:"email" json:"email,omitempty" `                                     //密保邮件，需要发送校验邮件确认
	UserType          pb.UserType `form:"user_type" json:"user_type" binding:"required"`                     //用户类型
	Deleted           int         `form:"deteled" json:"deteled"`                                            //软删除开关
	State             int         `form:"state" json:"state"`                                                //状态 0-正常 1-禁用
	Extend            string      `form:"extend" json:"extend,omitempty" `                                   //扩展字段
	ContactPerson     string      `form:"contact_person" json:"contact_person" binding:"required"`           //联系人
	Introductory      string      `gorm:"type:text;null" form:"introductory" json:"introductory,omitempty" ` // Text文本类型
	Province          string      `form:"province" json:"province,omitempty" `                               //省份, 如广东省
	City              string      `form:"city" json:"city,omitempty" `                                       //城市，如广州市
	County            string      `form:"county" json:"county,omitempty" `                                   //区，如天河区
	Street            string      `form:"street" json:"street,omitempty" `                                   //街道
	Address           string      `form:"address" json:"address,omitempty" `                                 //地址
	BranchesName      string      `form:"branches_name" json:"branches,omitempty" `                          //网点名称
	LegalPerson       string      `form:"legal_person" json:"legal_person,omitempty" `                       //法人姓名
	LegalIdentityCard string      `form:"legal_identity_card" json:"legal_identity_card,omitempty" `         //法人身份证
	CreatedBy         string      `form:"created_by" json:"created_by,omitempty"`                            //由谁创建，分为注册或后台添加
	ModifiedBy        string      `form:"modified_by" json:"modified_by,omitempty"`                          //最后由哪个操作员修改
}

//BeforeCreate CreatedAt赋值
func (user *User) BeforeCreate(scope *gorm.Scope) error {
	scope.SetColumn("CreatedAt", time.Now())
	return nil
}

//BeforeUpdate UpdatedAt赋值
func (user *User) BeforeUpdate(scope *gorm.Scope) error {
	scope.SetColumn("UpdatedAt", time.Now())
	return nil
}

// UserRole 用户身份结构体  用Account代替用户名
type UserRole struct {
	UserName  string
	UserRoles []*Role
}
