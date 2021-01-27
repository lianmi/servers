package models

//Role 身份信息结构体
type Role struct {
	ID       int    `gorm:"primarykey" json:"id"`
	UserID   uint   `json:"user_id"`
	UserName string `json:"user_name"`
	Value    string `json:"value"`
	// Mobile   string `json:"mobile"`
	// Smscode  string `json:"smscode"`
}
