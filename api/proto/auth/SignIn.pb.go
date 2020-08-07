// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.7.0
// source: api/proto/auth/SignIn.proto

package auth

import (
	proto "github.com/golang/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

//客户端类型
//是否必填：是
type ClientType int32

const (
	ClientType_Ct_UnKnow  ClientType = 0
	ClientType_Ct_Android ClientType = 1
	ClientType_Ct_iOS     ClientType = 2
	ClientType_Ct_RESTApi ClientType = 3
	ClientType_Ct_Windows ClientType = 4
	ClientType_Ct_MacOS   ClientType = 5
	ClientType_Ct_Web     ClientType = 6
)

// Enum value maps for ClientType.
var (
	ClientType_name = map[int32]string{
		0: "Ct_UnKnow",
		1: "Ct_Android",
		2: "Ct_iOS",
		3: "Ct_RESTApi",
		4: "Ct_Windows",
		5: "Ct_MacOS",
		6: "Ct_Web",
	}
	ClientType_value = map[string]int32{
		"Ct_UnKnow":  0,
		"Ct_Android": 1,
		"Ct_iOS":     2,
		"Ct_RESTApi": 3,
		"Ct_Windows": 4,
		"Ct_MacOS":   5,
		"Ct_Web":     6,
	}
)

func (x ClientType) Enum() *ClientType {
	p := new(ClientType)
	*p = x
	return p
}

func (x ClientType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ClientType) Descriptor() protoreflect.EnumDescriptor {
	return file_api_proto_auth_SignIn_proto_enumTypes[0].Descriptor()
}

func (ClientType) Type() protoreflect.EnumType {
	return &file_api_proto_auth_SignIn_proto_enumTypes[0]
}

func (x ClientType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ClientType.Descriptor instead.
func (ClientType) EnumDescriptor() ([]byte, []int) {
	return file_api_proto_auth_SignIn_proto_rawDescGZIP(), []int{0}
}

//客户端模式
type ClientMode int32

const (
	ClientMode_Clm_UnKnow      ClientMode = 0
	ClientMode_Clm_Im          ClientMode = 1 //一般模式
	ClientMode_Clm_ImEncrypted ClientMode = 2 //加密模式
)

// Enum value maps for ClientMode.
var (
	ClientMode_name = map[int32]string{
		0: "Clm_UnKnow",
		1: "Clm_Im",
		2: "Clm_ImEncrypted",
	}
	ClientMode_value = map[string]int32{
		"Clm_UnKnow":      0,
		"Clm_Im":          1,
		"Clm_ImEncrypted": 2,
	}
)

func (x ClientMode) Enum() *ClientMode {
	p := new(ClientMode)
	*p = x
	return p
}

func (x ClientMode) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ClientMode) Descriptor() protoreflect.EnumDescriptor {
	return file_api_proto_auth_SignIn_proto_enumTypes[1].Descriptor()
}

func (ClientMode) Type() protoreflect.EnumType {
	return &file_api_proto_auth_SignIn_proto_enumTypes[1]
}

func (x ClientMode) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ClientMode.Descriptor instead.
func (ClientMode) EnumDescriptor() ([]byte, []int) {
	return file_api_proto_auth_SignIn_proto_rawDescGZIP(), []int{1}
}

//多端同时登录时，其他在线的客户端的信息
type DeviceInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//用户注册号
	//是否必填-是
	Username string `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
	//连接ID,服务器分配,等同于http服务中的sessionId
	//示例:624ecb8b-308a-451c-be6b-62faca49848b
	//是否必填-是
	ConnectionId string `protobuf:"bytes,2,opt,name=connectionId,proto3" json:"connectionId,omitempty"`
	//设备Id
	//是否必填-是
	DeviceId string `protobuf:"bytes,3,opt,name=deviceId,proto3" json:"deviceId,omitempty"`
	//操作系统版本
	//是否必填-是
	Os string `protobuf:"bytes,4,opt,name=os,proto3" json:"os,omitempty"`
	//设备IP
	//是否必填-是
	Ip string `protobuf:"bytes,5,opt,name=ip,proto3" json:"ip,omitempty"`
	//设备类型
	//是否必填-是
	ClientType ClientType `protobuf:"varint,6,opt,name=clientType,proto3,enum=cloud.lianmi.im.auth.ClientType" json:"clientType,omitempty"`
	//该设备最后登录时间
	//是否必填-是
	Time uint64 `protobuf:"fixed64,7,opt,name=time,proto3" json:"time,omitempty"`
	//设备索引号
	//是否必填-是
	DeviceIndex int32 `protobuf:"varint,8,opt,name=deviceIndex,proto3" json:"deviceIndex,omitempty"`
	//是否是主设备
	//是否必填：是
	IsMaster bool `protobuf:"varint,9,opt,name=isMaster,proto3" json:"isMaster,omitempty"`
}

func (x *DeviceInfo) Reset() {
	*x = DeviceInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_auth_SignIn_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeviceInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeviceInfo) ProtoMessage() {}

func (x *DeviceInfo) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_auth_SignIn_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeviceInfo.ProtoReflect.Descriptor instead.
func (*DeviceInfo) Descriptor() ([]byte, []int) {
	return file_api_proto_auth_SignIn_proto_rawDescGZIP(), []int{0}
}

func (x *DeviceInfo) GetUsername() string {
	if x != nil {
		return x.Username
	}
	return ""
}

func (x *DeviceInfo) GetConnectionId() string {
	if x != nil {
		return x.ConnectionId
	}
	return ""
}

func (x *DeviceInfo) GetDeviceId() string {
	if x != nil {
		return x.DeviceId
	}
	return ""
}

func (x *DeviceInfo) GetOs() string {
	if x != nil {
		return x.Os
	}
	return ""
}

func (x *DeviceInfo) GetIp() string {
	if x != nil {
		return x.Ip
	}
	return ""
}

func (x *DeviceInfo) GetClientType() ClientType {
	if x != nil {
		return x.ClientType
	}
	return ClientType_Ct_UnKnow
}

func (x *DeviceInfo) GetTime() uint64 {
	if x != nil {
		return x.Time
	}
	return 0
}

func (x *DeviceInfo) GetDeviceIndex() int32 {
	if x != nil {
		return x.DeviceIndex
	}
	return 0
}

func (x *DeviceInfo) GetIsMaster() bool {
	if x != nil {
		return x.IsMaster
	}
	return false
}

//2.2.1
//登录请求
type SignInReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//用户注册号，在用户注册的时候由系统自动生成，字母 + 数字
	//是否必填：是
	Username string `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
	//应用ID，IM管理后台分配
	//是否必填：是
	AppKey string `protobuf:"bytes,2,opt,name=appKey,proto3" json:"appKey,omitempty"`
	//鉴权token,由业务系统请求IM REST服务器分配
	//是否必填：是
	Token string `protobuf:"bytes,3,opt,name=token,proto3" json:"token,omitempty"`
	//客户端具体类型
	//是否必填：是
	ClientType ClientType `protobuf:"varint,4,opt,name=clientType,proto3,enum=cloud.lianmi.im.auth.ClientType" json:"clientType,omitempty"`
	//用户标签,业务系统设置,用于根据标签进行消息推送,多个用半角逗号","分割
	//是否必填：否
	CustomTag string `protobuf:"bytes,5,opt,name=customTag,proto3" json:"customTag,omitempty"`
	//设备Id
	//是否必填：是
	DeviceId string `protobuf:"bytes,6,opt,name=deviceId,proto3" json:"deviceId,omitempty"`
	//操作系统版本
	//是否必填：是
	Os string `protobuf:"bytes,7,opt,name=os,proto3" json:"os,omitempty"`
	//协议版本
	//是否必填：是
	ProtocolVersion string `protobuf:"bytes,8,opt,name=protocolVersion,proto3" json:"protocolVersion,omitempty"`
	//SDK版本
	//是否必填：是
	SdkVersion string `protobuf:"bytes,9,opt,name=sdkVersion,proto3" json:"sdkVersion,omitempty"`
	//是否是主设备
	//是否必填：是
	IsMaster bool `protobuf:"varint,10,opt,name=isMaster,proto3" json:"isMaster,omitempty"`
	//客户端模式，一般模式和加密模式
	//是否必填：是
	ClientMode ClientMode `protobuf:"varint,11,opt,name=clientMode,proto3,enum=cloud.lianmi.im.auth.ClientMode" json:"clientMode,omitempty"`
}

func (x *SignInReq) Reset() {
	*x = SignInReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_auth_SignIn_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SignInReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SignInReq) ProtoMessage() {}

func (x *SignInReq) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_auth_SignIn_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SignInReq.ProtoReflect.Descriptor instead.
func (*SignInReq) Descriptor() ([]byte, []int) {
	return file_api_proto_auth_SignIn_proto_rawDescGZIP(), []int{1}
}

func (x *SignInReq) GetUsername() string {
	if x != nil {
		return x.Username
	}
	return ""
}

func (x *SignInReq) GetAppKey() string {
	if x != nil {
		return x.AppKey
	}
	return ""
}

func (x *SignInReq) GetToken() string {
	if x != nil {
		return x.Token
	}
	return ""
}

func (x *SignInReq) GetClientType() ClientType {
	if x != nil {
		return x.ClientType
	}
	return ClientType_Ct_UnKnow
}

func (x *SignInReq) GetCustomTag() string {
	if x != nil {
		return x.CustomTag
	}
	return ""
}

func (x *SignInReq) GetDeviceId() string {
	if x != nil {
		return x.DeviceId
	}
	return ""
}

func (x *SignInReq) GetOs() string {
	if x != nil {
		return x.Os
	}
	return ""
}

func (x *SignInReq) GetProtocolVersion() string {
	if x != nil {
		return x.ProtocolVersion
	}
	return ""
}

func (x *SignInReq) GetSdkVersion() string {
	if x != nil {
		return x.SdkVersion
	}
	return ""
}

func (x *SignInReq) GetIsMaster() bool {
	if x != nil {
		return x.IsMaster
	}
	return false
}

func (x *SignInReq) GetClientMode() ClientMode {
	if x != nil {
		return x.ClientMode
	}
	return ClientMode_Clm_UnKnow
}

//2.2.1
//登录请求
type SignInRsp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//连接ID,服务器分配,等同于http服务中的sessionId
	//是否必填：是
	ConnectionId string `protobuf:"bytes,1,opt,name=connectionId,proto3" json:"connectionId,omitempty"`
	//国家iso编码
	//是否必填：否
	Country string `protobuf:"bytes,2,opt,name=country,proto3" json:"country,omitempty"`
	//用户标签
	//是否必填：否
	CustomTag string `protobuf:"bytes,3,opt,name=customTag,proto3" json:"customTag,omitempty"`
	//客户端接入IP
	//是否必填：是
	Ip string `protobuf:"bytes,4,opt,name=ip,proto3" json:"ip,omitempty"`
	//客户端接入端口
	//是否必填：是
	Port string `protobuf:"bytes,5,opt,name=port,proto3" json:"port,omitempty"`
	//最后登录设备ID
	//是否必填：否
	LastLoginDeviceId string `protobuf:"bytes,6,opt,name=lastLoginDeviceId,proto3" json:"lastLoginDeviceId,omitempty"`
	//服务器分配的临时设备注册令牌
	//是否必填-否
	RegToken string `protobuf:"bytes,7,opt,name=regToken,proto3" json:"regToken,omitempty"`
}

func (x *SignInRsp) Reset() {
	*x = SignInRsp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_auth_SignIn_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SignInRsp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SignInRsp) ProtoMessage() {}

func (x *SignInRsp) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_auth_SignIn_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SignInRsp.ProtoReflect.Descriptor instead.
func (*SignInRsp) Descriptor() ([]byte, []int) {
	return file_api_proto_auth_SignIn_proto_rawDescGZIP(), []int{2}
}

func (x *SignInRsp) GetConnectionId() string {
	if x != nil {
		return x.ConnectionId
	}
	return ""
}

func (x *SignInRsp) GetCountry() string {
	if x != nil {
		return x.Country
	}
	return ""
}

func (x *SignInRsp) GetCustomTag() string {
	if x != nil {
		return x.CustomTag
	}
	return ""
}

func (x *SignInRsp) GetIp() string {
	if x != nil {
		return x.Ip
	}
	return ""
}

func (x *SignInRsp) GetPort() string {
	if x != nil {
		return x.Port
	}
	return ""
}

func (x *SignInRsp) GetLastLoginDeviceId() string {
	if x != nil {
		return x.LastLoginDeviceId
	}
	return ""
}

func (x *SignInRsp) GetRegToken() string {
	if x != nil {
		return x.RegToken
	}
	return ""
}

var File_api_proto_auth_SignIn_proto protoreflect.FileDescriptor

var file_api_proto_auth_SignIn_proto_rawDesc = []byte{
	0x0a, 0x1b, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x61, 0x75, 0x74, 0x68,
	0x2f, 0x53, 0x69, 0x67, 0x6e, 0x49, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x14, 0x63,
	0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x61,
	0x75, 0x74, 0x68, 0x22, 0x9c, 0x02, 0x0a, 0x0a, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x49, 0x6e,
	0x66, 0x6f, 0x12, 0x1a, 0x0a, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x22,
	0x0a, 0x0c, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x49, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x49, 0x64, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x49, 0x64, 0x12, 0x0e,
	0x0a, 0x02, 0x6f, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x6f, 0x73, 0x12, 0x0e,
	0x0a, 0x02, 0x69, 0x70, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x70, 0x12, 0x40,
	0x0a, 0x0a, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x18, 0x06, 0x20, 0x01,
	0x28, 0x0e, 0x32, 0x20, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d,
	0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74,
	0x54, 0x79, 0x70, 0x65, 0x52, 0x0a, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65,
	0x12, 0x12, 0x0a, 0x04, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x06, 0x52, 0x04,
	0x74, 0x69, 0x6d, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x64, 0x65, 0x76, 0x69, 0x63, 0x65, 0x49, 0x6e,
	0x64, 0x65, 0x78, 0x18, 0x08, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0b, 0x64, 0x65, 0x76, 0x69, 0x63,
	0x65, 0x49, 0x6e, 0x64, 0x65, 0x78, 0x12, 0x1a, 0x0a, 0x08, 0x69, 0x73, 0x4d, 0x61, 0x73, 0x74,
	0x65, 0x72, 0x18, 0x09, 0x20, 0x01, 0x28, 0x08, 0x52, 0x08, 0x69, 0x73, 0x4d, 0x61, 0x73, 0x74,
	0x65, 0x72, 0x22, 0x89, 0x03, 0x0a, 0x09, 0x53, 0x69, 0x67, 0x6e, 0x49, 0x6e, 0x52, 0x65, 0x71,
	0x12, 0x1a, 0x0a, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x16, 0x0a, 0x06,
	0x61, 0x70, 0x70, 0x4b, 0x65, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x61, 0x70,
	0x70, 0x4b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x40, 0x0a, 0x0a, 0x63, 0x6c,
	0x69, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x20,
	0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d,
	0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65,
	0x52, 0x0a, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x1c, 0x0a, 0x09,
	0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x54, 0x61, 0x67, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x09, 0x63, 0x75, 0x73, 0x74, 0x6f, 0x6d, 0x54, 0x61, 0x67, 0x12, 0x1a, 0x0a, 0x08, 0x64, 0x65,
	0x76, 0x69, 0x63, 0x65, 0x49, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x64, 0x65,
	0x76, 0x69, 0x63, 0x65, 0x49, 0x64, 0x12, 0x0e, 0x0a, 0x02, 0x6f, 0x73, 0x18, 0x07, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x02, 0x6f, 0x73, 0x12, 0x28, 0x0a, 0x0f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63,
	0x6f, 0x6c, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e,
	0x12, 0x1e, 0x0a, 0x0a, 0x73, 0x64, 0x6b, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x09,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x73, 0x64, 0x6b, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e,
	0x12, 0x1a, 0x0a, 0x08, 0x69, 0x73, 0x4d, 0x61, 0x73, 0x74, 0x65, 0x72, 0x18, 0x0a, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x08, 0x69, 0x73, 0x4d, 0x61, 0x73, 0x74, 0x65, 0x72, 0x12, 0x40, 0x0a, 0x0a,
	0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x4d, 0x6f, 0x64, 0x65, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x0e,
	0x32, 0x20, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e,
	0x69, 0x6d, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x4d, 0x6f,
	0x64, 0x65, 0x52, 0x0a, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x4d, 0x6f, 0x64, 0x65, 0x22, 0xd5,
	0x01, 0x0a, 0x09, 0x53, 0x69, 0x67, 0x6e, 0x49, 0x6e, 0x52, 0x73, 0x70, 0x12, 0x22, 0x0a, 0x0c,
	0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0c, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64,
	0x12, 0x18, 0x0a, 0x07, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x07, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x1c, 0x0a, 0x09, 0x63, 0x75,
	0x73, 0x74, 0x6f, 0x6d, 0x54, 0x61, 0x67, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x63,
	0x75, 0x73, 0x74, 0x6f, 0x6d, 0x54, 0x61, 0x67, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x70, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x70, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x6f, 0x72, 0x74,
	0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x70, 0x6f, 0x72, 0x74, 0x12, 0x2c, 0x0a, 0x11,
	0x6c, 0x61, 0x73, 0x74, 0x4c, 0x6f, 0x67, 0x69, 0x6e, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x49,
	0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x11, 0x6c, 0x61, 0x73, 0x74, 0x4c, 0x6f, 0x67,
	0x69, 0x6e, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x49, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x72, 0x65,
	0x67, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x72, 0x65,
	0x67, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x2a, 0x71, 0x0a, 0x0a, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74,
	0x54, 0x79, 0x70, 0x65, 0x12, 0x0d, 0x0a, 0x09, 0x43, 0x74, 0x5f, 0x55, 0x6e, 0x4b, 0x6e, 0x6f,
	0x77, 0x10, 0x00, 0x12, 0x0e, 0x0a, 0x0a, 0x43, 0x74, 0x5f, 0x41, 0x6e, 0x64, 0x72, 0x6f, 0x69,
	0x64, 0x10, 0x01, 0x12, 0x0a, 0x0a, 0x06, 0x43, 0x74, 0x5f, 0x69, 0x4f, 0x53, 0x10, 0x02, 0x12,
	0x0e, 0x0a, 0x0a, 0x43, 0x74, 0x5f, 0x52, 0x45, 0x53, 0x54, 0x41, 0x70, 0x69, 0x10, 0x03, 0x12,
	0x0e, 0x0a, 0x0a, 0x43, 0x74, 0x5f, 0x57, 0x69, 0x6e, 0x64, 0x6f, 0x77, 0x73, 0x10, 0x04, 0x12,
	0x0c, 0x0a, 0x08, 0x43, 0x74, 0x5f, 0x4d, 0x61, 0x63, 0x4f, 0x53, 0x10, 0x05, 0x12, 0x0a, 0x0a,
	0x06, 0x43, 0x74, 0x5f, 0x57, 0x65, 0x62, 0x10, 0x06, 0x2a, 0x3d, 0x0a, 0x0a, 0x43, 0x6c, 0x69,
	0x65, 0x6e, 0x74, 0x4d, 0x6f, 0x64, 0x65, 0x12, 0x0e, 0x0a, 0x0a, 0x43, 0x6c, 0x6d, 0x5f, 0x55,
	0x6e, 0x4b, 0x6e, 0x6f, 0x77, 0x10, 0x00, 0x12, 0x0a, 0x0a, 0x06, 0x43, 0x6c, 0x6d, 0x5f, 0x49,
	0x6d, 0x10, 0x01, 0x12, 0x13, 0x0a, 0x0f, 0x43, 0x6c, 0x6d, 0x5f, 0x49, 0x6d, 0x45, 0x6e, 0x63,
	0x72, 0x79, 0x70, 0x74, 0x65, 0x64, 0x10, 0x02, 0x42, 0x2a, 0x5a, 0x28, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2f, 0x73, 0x65,
	0x72, 0x76, 0x65, 0x72, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f,
	0x61, 0x75, 0x74, 0x68, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_proto_auth_SignIn_proto_rawDescOnce sync.Once
	file_api_proto_auth_SignIn_proto_rawDescData = file_api_proto_auth_SignIn_proto_rawDesc
)

func file_api_proto_auth_SignIn_proto_rawDescGZIP() []byte {
	file_api_proto_auth_SignIn_proto_rawDescOnce.Do(func() {
		file_api_proto_auth_SignIn_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_proto_auth_SignIn_proto_rawDescData)
	})
	return file_api_proto_auth_SignIn_proto_rawDescData
}

var file_api_proto_auth_SignIn_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_api_proto_auth_SignIn_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_api_proto_auth_SignIn_proto_goTypes = []interface{}{
	(ClientType)(0),    // 0: cloud.lianmi.im.auth.ClientType
	(ClientMode)(0),    // 1: cloud.lianmi.im.auth.ClientMode
	(*DeviceInfo)(nil), // 2: cloud.lianmi.im.auth.DeviceInfo
	(*SignInReq)(nil),  // 3: cloud.lianmi.im.auth.SignInReq
	(*SignInRsp)(nil),  // 4: cloud.lianmi.im.auth.SignInRsp
}
var file_api_proto_auth_SignIn_proto_depIdxs = []int32{
	0, // 0: cloud.lianmi.im.auth.DeviceInfo.clientType:type_name -> cloud.lianmi.im.auth.ClientType
	0, // 1: cloud.lianmi.im.auth.SignInReq.clientType:type_name -> cloud.lianmi.im.auth.ClientType
	1, // 2: cloud.lianmi.im.auth.SignInReq.clientMode:type_name -> cloud.lianmi.im.auth.ClientMode
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_api_proto_auth_SignIn_proto_init() }
func file_api_proto_auth_SignIn_proto_init() {
	if File_api_proto_auth_SignIn_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_proto_auth_SignIn_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeviceInfo); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_proto_auth_SignIn_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SignInReq); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_api_proto_auth_SignIn_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SignInRsp); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_api_proto_auth_SignIn_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_proto_auth_SignIn_proto_goTypes,
		DependencyIndexes: file_api_proto_auth_SignIn_proto_depIdxs,
		EnumInfos:         file_api_proto_auth_SignIn_proto_enumTypes,
		MessageInfos:      file_api_proto_auth_SignIn_proto_msgTypes,
	}.Build()
	File_api_proto_auth_SignIn_proto = out.File
	file_api_proto_auth_SignIn_proto_rawDesc = nil
	file_api_proto_auth_SignIn_proto_goTypes = nil
	file_api_proto_auth_SignIn_proto_depIdxs = nil
}
