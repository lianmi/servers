package order

import (
	"encoding/json"
)

//数据结构
type VipUser struct {
	PayType int `form:"pay_type" json:"pay_type"` //VIP类型，1-包年，2-包季， 3-包月
}

func (m *VipUser) ToJson() (string, error) {
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}
