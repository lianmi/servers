package models

//用户对客服评分
type Grade struct {
	ID                      uint64 `gorm:"primary_key" form:"id" json:"id,omitempty"` //自动递增id
	CreatedAt               int64  `form:"created_at" json:"created_at,omitempty"`    //创建时刻,毫秒
	UpdatedAt               int64  `form:"updated_at" json:"updated_at,omitempty"`
	Title                   string `json:"title" `                                 //本次app用户求助的标题，约定： consult + _+ 日期字符串(20201025) + _ + 编号（自增）
	AppUsername             string `json:"app_username" `                          //APP用户的注册账号id
	CustomerServiceUsername string `json:"customer_service_username" `             //客服或技术人员的注册账号id
	JobNumber               string `json:"jpb_number" `                            //客服或技术人员的工号
	Type                    int    `json:"type" `                                  //客服或技术人员的类型， 1-客服，2-技术
	Evaluation              string `json:"evaluation" `                            //职称, 技术工程师，技术员等
	NickName                string `json:"nick_name" `                             //呢称
	Catalog                 string `json:"catalog" `                               //问题类型
	Desc                    string `json:"desc" gorm:"column:desc;type:longtext" ` //问题描述
	GradeNum                int    `json:"grade_num" `                             //评分, 0-3 4-6 7-10
}
