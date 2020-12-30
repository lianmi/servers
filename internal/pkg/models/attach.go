package models

import (
	"encoding/json"
	// "gorm.io/gorm"
	// "time"
)

// 订单协议的 OrderProductBody 里 Attach的最外层结构
//约定， type=99是会员购买
//约定， type=100是订单服务费
type AttachBase struct {
	BodyType int    `json:"body_type" validate:"required"` //UI与服务端约定好， 区分Body关联哪一个反JSON数据结构
	Body     string `json:"body"`                          //自定义
}

func (a *AttachBase) ToJson() (string, error) {
	jsonBytes, err := json.Marshal(a)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func AttachBaseFromJson(data []byte) (*AttachBase, error) {

	attachBase := new(AttachBase)
	err := json.Unmarshal(data, attachBase)
	return attachBase, err

}

//服务费回包attach数据结构, 存放是真正订单数据
type OrignOrder struct {
	OrderID string `json:"order_id" validate:"required"` //订单ID
}

func (o *OrignOrder) ToJson() (string, error) {
	jsonBytes, err := json.Marshal(o)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func OrignOrderFromJson(data []byte) (*OrignOrder, error) {

	orignOrder := new(OrignOrder)
	err := json.Unmarshal(data, orignOrder)
	return orignOrder, err

}

//UI构造attach数据结构
type VipUser struct {
	PayType int     `form:"pay_type" json:"pay_type"`     //VIP类型，1-包年，2-包季， 3-包月
	Price   float32 `form:"price" json:"price,omitempty"` //价格, 单位: 元
}

func (m *VipUser) ToJson() (string, error) {
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

func VipUserFromJson(data []byte) (*VipUser, error) {

	vipUser := new(VipUser)
	err := json.Unmarshal(data, vipUser)
	return vipUser, err

}
