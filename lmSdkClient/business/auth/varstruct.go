package auth

import (
	"encoding/json"
)

type Login struct {
	Username        string `form:"username" json:"username" binding:"required"`
	Password        string `form:"password" json:"password" binding:"required"`
	SmsCode         string `form:"smscode" json:"smscode" binding:"required"`
	DeviceID        string `form:"deviceid" json:"deviceid" binding:"required"`
	UserType        int    `form:"userType" json:"userType" binding:"required"`
	Os              string `form:"os" json:"os" binding:"required"`
	ProtocolVersion string `form:"protocolversion" json:"protocolversion" binding:"required"`
	SdkVersion      string `form:"sdkversion" json:"sdkversion" binding:"required"`
	IsMaster        bool   `form:"ismaster" json:"ismaster"` //由于golang对false处理不对，所以不能设为必填
}

func (m *Login) ToJson() (string, error) {
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}
