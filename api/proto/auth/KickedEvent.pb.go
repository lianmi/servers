// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.7.0
// source: api/proto/auth/KickedEvent.proto

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

//被踢原因枚举
type KickReason int32

const (
	//无效
	KickReason_KickReasonUndefined KickReason = 0
	//不允许同一个帐号在多个地方同时登录
	KickReason_SamePlatformKick KickReason = 1
	//系统封号
	KickReason_Blacked KickReason = 2
	//被其它端踢了
	KickReason_OtherPlatformKick KickReason = 3
)

// Enum value maps for KickReason.
var (
	KickReason_name = map[int32]string{
		0: "KickReasonUndefined",
		1: "SamePlatformKick",
		2: "Blacked",
		3: "OtherPlatformKick",
	}
	KickReason_value = map[string]int32{
		"KickReasonUndefined": 0,
		"SamePlatformKick":    1,
		"Blacked":             2,
		"OtherPlatformKick":   3,
	}
)

func (x KickReason) Enum() *KickReason {
	p := new(KickReason)
	*p = x
	return p
}

func (x KickReason) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (KickReason) Descriptor() protoreflect.EnumDescriptor {
	return file_api_proto_auth_KickedEvent_proto_enumTypes[0].Descriptor()
}

func (KickReason) Type() protoreflect.EnumType {
	return &file_api_proto_auth_KickedEvent_proto_enumTypes[0]
}

func (x KickReason) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use KickReason.Descriptor instead.
func (KickReason) EnumDescriptor() ([]byte, []int) {
	return file_api_proto_auth_KickedEvent_proto_rawDescGZIP(), []int{0}
}

//
//Api描述
//当前版本支持单一设备在线,如发生账号在其他设备登录,则发送该事件将当前设备离线
//Api类型
//Event
type KickedEventRsp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//消息来源,如果是服务器端踢出,可以忽略该字段
	//是否必填-是
	ClientType ClientType `protobuf:"varint,1,opt,name=clientType,proto3,enum=cloud.lianmi.im.auth.ClientType" json:"clientType,omitempty"`
	//被踢原因
	//是否必填-是
	Reason KickReason `protobuf:"varint,2,opt,name=reason,proto3,enum=cloud.lianmi.im.auth.KickReason" json:"reason,omitempty"`
	//unix时间戳
	//是否必填-是
	TimeTag uint64 `protobuf:"fixed64,3,opt,name=timeTag,proto3" json:"timeTag,omitempty"`
}

func (x *KickedEventRsp) Reset() {
	*x = KickedEventRsp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_auth_KickedEvent_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *KickedEventRsp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KickedEventRsp) ProtoMessage() {}

func (x *KickedEventRsp) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_auth_KickedEvent_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KickedEventRsp.ProtoReflect.Descriptor instead.
func (*KickedEventRsp) Descriptor() ([]byte, []int) {
	return file_api_proto_auth_KickedEvent_proto_rawDescGZIP(), []int{0}
}

func (x *KickedEventRsp) GetClientType() ClientType {
	if x != nil {
		return x.ClientType
	}
	return ClientType_Ct_UnKnow
}

func (x *KickedEventRsp) GetReason() KickReason {
	if x != nil {
		return x.Reason
	}
	return KickReason_KickReasonUndefined
}

func (x *KickedEventRsp) GetTimeTag() uint64 {
	if x != nil {
		return x.TimeTag
	}
	return 0
}

var File_api_proto_auth_KickedEvent_proto protoreflect.FileDescriptor

var file_api_proto_auth_KickedEvent_proto_rawDesc = []byte{
	0x0a, 0x20, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x61, 0x75, 0x74, 0x68,
	0x2f, 0x4b, 0x69, 0x63, 0x6b, 0x65, 0x64, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x14, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69,
	0x2e, 0x69, 0x6d, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x1a, 0x1b, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2f, 0x61, 0x75, 0x74, 0x68, 0x2f, 0x53, 0x69, 0x67, 0x6e, 0x49, 0x6e, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xa6, 0x01, 0x0a, 0x0e, 0x4b, 0x69, 0x63, 0x6b, 0x65, 0x64,
	0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x73, 0x70, 0x12, 0x40, 0x0a, 0x0a, 0x63, 0x6c, 0x69, 0x65,
	0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x20, 0x2e, 0x63,
	0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x61,
	0x75, 0x74, 0x68, 0x2e, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x52, 0x0a,
	0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x38, 0x0a, 0x06, 0x72, 0x65,
	0x61, 0x73, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x20, 0x2e, 0x63, 0x6c, 0x6f,
	0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x61, 0x75, 0x74,
	0x68, 0x2e, 0x4b, 0x69, 0x63, 0x6b, 0x52, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x52, 0x06, 0x72, 0x65,
	0x61, 0x73, 0x6f, 0x6e, 0x12, 0x18, 0x0a, 0x07, 0x74, 0x69, 0x6d, 0x65, 0x54, 0x61, 0x67, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x06, 0x52, 0x07, 0x74, 0x69, 0x6d, 0x65, 0x54, 0x61, 0x67, 0x2a, 0x5f,
	0x0a, 0x0a, 0x4b, 0x69, 0x63, 0x6b, 0x52, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x12, 0x17, 0x0a, 0x13,
	0x4b, 0x69, 0x63, 0x6b, 0x52, 0x65, 0x61, 0x73, 0x6f, 0x6e, 0x55, 0x6e, 0x64, 0x65, 0x66, 0x69,
	0x6e, 0x65, 0x64, 0x10, 0x00, 0x12, 0x14, 0x0a, 0x10, 0x53, 0x61, 0x6d, 0x65, 0x50, 0x6c, 0x61,
	0x74, 0x66, 0x6f, 0x72, 0x6d, 0x4b, 0x69, 0x63, 0x6b, 0x10, 0x01, 0x12, 0x0b, 0x0a, 0x07, 0x42,
	0x6c, 0x61, 0x63, 0x6b, 0x65, 0x64, 0x10, 0x02, 0x12, 0x15, 0x0a, 0x11, 0x4f, 0x74, 0x68, 0x65,
	0x72, 0x50, 0x6c, 0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x4b, 0x69, 0x63, 0x6b, 0x10, 0x03, 0x42,
	0x2a, 0x5a, 0x28, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x69,
	0x61, 0x6e, 0x6d, 0x69, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x2f, 0x61, 0x70, 0x69,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x61, 0x75, 0x74, 0x68, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_api_proto_auth_KickedEvent_proto_rawDescOnce sync.Once
	file_api_proto_auth_KickedEvent_proto_rawDescData = file_api_proto_auth_KickedEvent_proto_rawDesc
)

func file_api_proto_auth_KickedEvent_proto_rawDescGZIP() []byte {
	file_api_proto_auth_KickedEvent_proto_rawDescOnce.Do(func() {
		file_api_proto_auth_KickedEvent_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_proto_auth_KickedEvent_proto_rawDescData)
	})
	return file_api_proto_auth_KickedEvent_proto_rawDescData
}

var file_api_proto_auth_KickedEvent_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_api_proto_auth_KickedEvent_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_api_proto_auth_KickedEvent_proto_goTypes = []interface{}{
	(KickReason)(0),        // 0: cloud.lianmi.im.auth.KickReason
	(*KickedEventRsp)(nil), // 1: cloud.lianmi.im.auth.KickedEventRsp
	(ClientType)(0),        // 2: cloud.lianmi.im.auth.ClientType
}
var file_api_proto_auth_KickedEvent_proto_depIdxs = []int32{
	2, // 0: cloud.lianmi.im.auth.KickedEventRsp.clientType:type_name -> cloud.lianmi.im.auth.ClientType
	0, // 1: cloud.lianmi.im.auth.KickedEventRsp.reason:type_name -> cloud.lianmi.im.auth.KickReason
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_api_proto_auth_KickedEvent_proto_init() }
func file_api_proto_auth_KickedEvent_proto_init() {
	if File_api_proto_auth_KickedEvent_proto != nil {
		return
	}
	file_api_proto_auth_SignIn_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_api_proto_auth_KickedEvent_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*KickedEventRsp); i {
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
			RawDescriptor: file_api_proto_auth_KickedEvent_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_proto_auth_KickedEvent_proto_goTypes,
		DependencyIndexes: file_api_proto_auth_KickedEvent_proto_depIdxs,
		EnumInfos:         file_api_proto_auth_KickedEvent_proto_enumTypes,
		MessageInfos:      file_api_proto_auth_KickedEvent_proto_msgTypes,
	}.Build()
	File_api_proto_auth_KickedEvent_proto = out.File
	file_api_proto_auth_KickedEvent_proto_rawDesc = nil
	file_api_proto_auth_KickedEvent_proto_goTypes = nil
	file_api_proto_auth_KickedEvent_proto_depIdxs = nil
}