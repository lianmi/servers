/*
用户标签表
*/
package models

//定义标签表的数据结构
type Tag struct {
	// global.LMC_Model  //不能用，因为需要用userName做主键

	Username       string `gorm:"primarykey" json:"username" ` //用户注册号
	TargetUsername string `json:"target_username" `            //目标用户注册号
	TagType        int    `form:"tag_type" json:"tag_type"`    //为目标用户打的标签类型
}

// //BeforeCreate CreatedAt赋值
// func (t *Tag) BeforeCreate(tx *gorm.DB) error {
// 	tx.Statement.SetColumn("CreatedAt", time.Now().UnixNano()/1e6)
// 	return nil
// }

// //BeforeUpdate UpdatedAt赋值
// func (t *Tag) BeforeUpdate(tx *gorm.DB) error {
// 	tx.Statement.SetColumn("UpdatedAt", time.Now().UnixNano()/1e6)
// 	return nil
// }
