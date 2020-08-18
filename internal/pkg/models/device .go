package models

//定义用户设备的数据结构
type Device struct {
	ID            uint64 `form:"id" json:"id,omitempty"`
	UserID        uint64 `form:"user_id" json:"user_id" binding:"required"`         //用户id
	IsMaster      bool   `form:"is_master" json:"is_master" binding:"required"`     //是否是主设备，true - 是， false -从设备
	DeviceName    string `form:"device_name" json:"device_name" binding:"required"` //设备uuid
	DeviceIndex   uint32 `form:"device_index" json:"device_index,omitempty"`        //设备编号，主设备默认 是1， 增删从设备将会一直递增 max + 1 主设备更换后，全部重置
	Os            string `form:"os" json:"os,omitempty"`                            //操作系统版本
	ClientType    uint32 `form:"client_type" json:"client_type" binding:"required"` //设备类型
	LatestLogonAt int64 `form:"latest_logon_at" json:"latest_logon_at,omitempty""` //最后登录时间
}
