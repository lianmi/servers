package models

/*
客服技术人员表
*/
type CustomerServiceInfo struct {
	ID         uint64 `gorm:"primary_key" form:"id" json:"id,omitempty"` //自动递增id
	CreatedAt  int64  `form:"created_at" json:"created_at,omitempty"`    //创建时刻,毫秒
	UpdatedAt  int64  `form:"updated_at" json:"updated_at,omitempty"`
	Username   string `json:"username" `   //客服或技术人员的注册账号id
	JobNumber  string `json:"job_number" ` //客服或技术人员的工号
	Type       int    `json:"type" `       //客服或技术人员的类型， 1-客服，2-技术
	Evaluation string `json:"evaluation" ` //职称, 技术工程师，技术员等
	NickName   string `json:"nick_name" `  //呢称
}
