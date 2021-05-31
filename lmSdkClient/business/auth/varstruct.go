package auth

import (
	"encoding/json"
)

type Login struct {
	Username   string `form:"username" json:"username"`
	Password   string `form:"password" json:"password"`
	SmsCode    string `form:"smscode" json:"smscode"`
	DeviceID   string `form:"deviceid" json:"deviceid"`
	UserType   int    `form:"userType" json:"userType"`
	Os         string `form:"os" json:"os" `
	WechatCode string `form:"wechat_code" json:"wechat_code"`
	SdkVersion string `form:"sdkversion" json:"sdkversion"`
	IsMaster   bool   `form:"ismaster" json:"ismaster"` //由于golang对false处理不对，所以不能设为必填
}

func (m *Login) ToJson() (string, error) {
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}
