// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.7.0
// source: api/proto/msg/MsgTypeEnum.proto

package msg

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

type MessageScene int32

const (
	//无效
	MessageScene_MsgScene_Undefined MessageScene = 0
	//人对人通讯
	MessageScene_MsgScene_C2C MessageScene = 1 //  用户对用户
	//群组通讯
	MessageScene_MsgScene_C2G MessageScene = 2 // 用户到群
	//系统消息
	MessageScene_MsgScene_S2C MessageScene = 3 // 服务端到 用户
	MessageScene_MsgScene_C2S MessageScene = 4 // 客户向服务端发起的事件
	//点对点
	MessageScene_MsgScene_P2P MessageScene = 5 // 透传 点对点
)

// Enum value maps for MessageScene.
var (
	MessageScene_name = map[int32]string{
		0: "MsgScene_Undefined",
		1: "MsgScene_C2C",
		2: "MsgScene_C2G",
		3: "MsgScene_S2C",
		4: "MsgScene_C2S",
		5: "MsgScene_P2P",
	}
	MessageScene_value = map[string]int32{
		"MsgScene_Undefined": 0,
		"MsgScene_C2C":       1,
		"MsgScene_C2G":       2,
		"MsgScene_S2C":       3,
		"MsgScene_C2S":       4,
		"MsgScene_P2P":       5,
	}
)

func (x MessageScene) Enum() *MessageScene {
	p := new(MessageScene)
	*p = x
	return p
}

func (x MessageScene) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (MessageScene) Descriptor() protoreflect.EnumDescriptor {
	return file_api_proto_msg_MsgTypeEnum_proto_enumTypes[0].Descriptor()
}

func (MessageScene) Type() protoreflect.EnumType {
	return &file_api_proto_msg_MsgTypeEnum_proto_enumTypes[0]
}

func (x MessageScene) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use MessageScene.Descriptor instead.
func (MessageScene) EnumDescriptor() ([]byte, []int) {
	return file_api_proto_msg_MsgTypeEnum_proto_rawDescGZIP(), []int{0}
}

//消息的类型枚举定义
type MessageType int32

const (
	//无效
	MessageType_MsgType_Undefined MessageType = 0
	//Text(0)-文本
	MessageType_MsgType_Text MessageType = 1
	// 附件类型
	//　数据经过　json 处理　的
	MessageType_MsgType_Attach MessageType = 2
	// 通知类型的数据
	MessageType_MsgType_Notification MessageType = 3
	// 加密类型
	// 在基础类型　中增加了　加密封装
	MessageType_MsgType_Secret MessageType = 4
	//  二进制
	//　直接二进制流bytes
	MessageType_MsgType_Bin MessageType = 5
	// 订单 类型
	MessageType_MsgType_Order MessageType = 6
	// 系统消息更新 类型
	MessageType_MsgType_SysMsgUpdate MessageType = 7
	//用户自定义
	MessageType_MSgType_Customer MessageType = 100
)

// Enum value maps for MessageType.
var (
	MessageType_name = map[int32]string{
		0:   "MsgType_Undefined",
		1:   "MsgType_Text",
		2:   "MsgType_Attach",
		3:   "MsgType_Notification",
		4:   "MsgType_Secret",
		5:   "MsgType_Bin",
		6:   "MsgType_Order",
		7:   "MsgType_SysMsgUpdate",
		100: "MSgType_Customer",
	}
	MessageType_value = map[string]int32{
		"MsgType_Undefined":    0,
		"MsgType_Text":         1,
		"MsgType_Attach":       2,
		"MsgType_Notification": 3,
		"MsgType_Secret":       4,
		"MsgType_Bin":          5,
		"MsgType_Order":        6,
		"MsgType_SysMsgUpdate": 7,
		"MSgType_Customer":     100,
	}
)

func (x MessageType) Enum() *MessageType {
	p := new(MessageType)
	*p = x
	return p
}

func (x MessageType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (MessageType) Descriptor() protoreflect.EnumDescriptor {
	return file_api_proto_msg_MsgTypeEnum_proto_enumTypes[1].Descriptor()
}

func (MessageType) Type() protoreflect.EnumType {
	return &file_api_proto_msg_MsgTypeEnum_proto_enumTypes[1]
}

func (x MessageType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use MessageType.Descriptor instead.
func (MessageType) EnumDescriptor() ([]byte, []int) {
	return file_api_proto_msg_MsgTypeEnum_proto_rawDescGZIP(), []int{1}
}

//附件类型枚举定义
type AttachType int32

const (
	//　未定义的附件类型
	AttachType_AttachType_Undefined AttachType = 0
	//   图片
	AttachType_AttachType_Image AttachType = 1
	//Audio(2) - 音频文件
	AttachType_AttachType_Audio AttachType = 2
	//Video(3) - 视频文件
	AttachType_AttachType_Video AttachType = 3
	//File(4) - 文件
	AttachType_AttachType_File AttachType = 4
	// 地理位置
	AttachType_AttachType_Geo AttachType = 5
	// 订单数据
	AttachType_AttachType_Order AttachType = 6
	//钱包相关的交易数据(提现，充值 ，转账，收款 ，退款等)
	AttachType_AttachType_Transaction AttachType = 7
)

// Enum value maps for AttachType.
var (
	AttachType_name = map[int32]string{
		0: "AttachType_Undefined",
		1: "AttachType_Image",
		2: "AttachType_Audio",
		3: "AttachType_Video",
		4: "AttachType_File",
		5: "AttachType_Geo",
		6: "AttachType_Order",
		7: "AttachType_Transaction",
	}
	AttachType_value = map[string]int32{
		"AttachType_Undefined":   0,
		"AttachType_Image":       1,
		"AttachType_Audio":       2,
		"AttachType_Video":       3,
		"AttachType_File":        4,
		"AttachType_Geo":         5,
		"AttachType_Order":       6,
		"AttachType_Transaction": 7,
	}
)

func (x AttachType) Enum() *AttachType {
	p := new(AttachType)
	*p = x
	return p
}

func (x AttachType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (AttachType) Descriptor() protoreflect.EnumDescriptor {
	return file_api_proto_msg_MsgTypeEnum_proto_enumTypes[2].Descriptor()
}

func (AttachType) Type() protoreflect.EnumType {
	return &file_api_proto_msg_MsgTypeEnum_proto_enumTypes[2]
}

func (x AttachType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use AttachType.Descriptor instead.
func (AttachType) EnumDescriptor() ([]byte, []int) {
	return file_api_proto_msg_MsgTypeEnum_proto_rawDescGZIP(), []int{2}
}

var File_api_proto_msg_MsgTypeEnum_proto protoreflect.FileDescriptor

var file_api_proto_msg_MsgTypeEnum_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x73, 0x67, 0x2f,
	0x4d, 0x73, 0x67, 0x54, 0x79, 0x70, 0x65, 0x45, 0x6e, 0x75, 0x6d, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x13, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e,
	0x69, 0x6d, 0x2e, 0x6d, 0x73, 0x67, 0x2a, 0x80, 0x01, 0x0a, 0x0c, 0x4d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x53, 0x63, 0x65, 0x6e, 0x65, 0x12, 0x16, 0x0a, 0x12, 0x4d, 0x73, 0x67, 0x53, 0x63,
	0x65, 0x6e, 0x65, 0x5f, 0x55, 0x6e, 0x64, 0x65, 0x66, 0x69, 0x6e, 0x65, 0x64, 0x10, 0x00, 0x12,
	0x10, 0x0a, 0x0c, 0x4d, 0x73, 0x67, 0x53, 0x63, 0x65, 0x6e, 0x65, 0x5f, 0x43, 0x32, 0x43, 0x10,
	0x01, 0x12, 0x10, 0x0a, 0x0c, 0x4d, 0x73, 0x67, 0x53, 0x63, 0x65, 0x6e, 0x65, 0x5f, 0x43, 0x32,
	0x47, 0x10, 0x02, 0x12, 0x10, 0x0a, 0x0c, 0x4d, 0x73, 0x67, 0x53, 0x63, 0x65, 0x6e, 0x65, 0x5f,
	0x53, 0x32, 0x43, 0x10, 0x03, 0x12, 0x10, 0x0a, 0x0c, 0x4d, 0x73, 0x67, 0x53, 0x63, 0x65, 0x6e,
	0x65, 0x5f, 0x43, 0x32, 0x53, 0x10, 0x04, 0x12, 0x10, 0x0a, 0x0c, 0x4d, 0x73, 0x67, 0x53, 0x63,
	0x65, 0x6e, 0x65, 0x5f, 0x50, 0x32, 0x50, 0x10, 0x05, 0x2a, 0xcc, 0x01, 0x0a, 0x0b, 0x4d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x15, 0x0a, 0x11, 0x4d, 0x73, 0x67,
	0x54, 0x79, 0x70, 0x65, 0x5f, 0x55, 0x6e, 0x64, 0x65, 0x66, 0x69, 0x6e, 0x65, 0x64, 0x10, 0x00,
	0x12, 0x10, 0x0a, 0x0c, 0x4d, 0x73, 0x67, 0x54, 0x79, 0x70, 0x65, 0x5f, 0x54, 0x65, 0x78, 0x74,
	0x10, 0x01, 0x12, 0x12, 0x0a, 0x0e, 0x4d, 0x73, 0x67, 0x54, 0x79, 0x70, 0x65, 0x5f, 0x41, 0x74,
	0x74, 0x61, 0x63, 0x68, 0x10, 0x02, 0x12, 0x18, 0x0a, 0x14, 0x4d, 0x73, 0x67, 0x54, 0x79, 0x70,
	0x65, 0x5f, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x10, 0x03,
	0x12, 0x12, 0x0a, 0x0e, 0x4d, 0x73, 0x67, 0x54, 0x79, 0x70, 0x65, 0x5f, 0x53, 0x65, 0x63, 0x72,
	0x65, 0x74, 0x10, 0x04, 0x12, 0x0f, 0x0a, 0x0b, 0x4d, 0x73, 0x67, 0x54, 0x79, 0x70, 0x65, 0x5f,
	0x42, 0x69, 0x6e, 0x10, 0x05, 0x12, 0x11, 0x0a, 0x0d, 0x4d, 0x73, 0x67, 0x54, 0x79, 0x70, 0x65,
	0x5f, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x10, 0x06, 0x12, 0x18, 0x0a, 0x14, 0x4d, 0x73, 0x67, 0x54,
	0x79, 0x70, 0x65, 0x5f, 0x53, 0x79, 0x73, 0x4d, 0x73, 0x67, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65,
	0x10, 0x07, 0x12, 0x14, 0x0a, 0x10, 0x4d, 0x53, 0x67, 0x54, 0x79, 0x70, 0x65, 0x5f, 0x43, 0x75,
	0x73, 0x74, 0x6f, 0x6d, 0x65, 0x72, 0x10, 0x64, 0x2a, 0xc3, 0x01, 0x0a, 0x0a, 0x41, 0x74, 0x74,
	0x61, 0x63, 0x68, 0x54, 0x79, 0x70, 0x65, 0x12, 0x18, 0x0a, 0x14, 0x41, 0x74, 0x74, 0x61, 0x63,
	0x68, 0x54, 0x79, 0x70, 0x65, 0x5f, 0x55, 0x6e, 0x64, 0x65, 0x66, 0x69, 0x6e, 0x65, 0x64, 0x10,
	0x00, 0x12, 0x14, 0x0a, 0x10, 0x41, 0x74, 0x74, 0x61, 0x63, 0x68, 0x54, 0x79, 0x70, 0x65, 0x5f,
	0x49, 0x6d, 0x61, 0x67, 0x65, 0x10, 0x01, 0x12, 0x14, 0x0a, 0x10, 0x41, 0x74, 0x74, 0x61, 0x63,
	0x68, 0x54, 0x79, 0x70, 0x65, 0x5f, 0x41, 0x75, 0x64, 0x69, 0x6f, 0x10, 0x02, 0x12, 0x14, 0x0a,
	0x10, 0x41, 0x74, 0x74, 0x61, 0x63, 0x68, 0x54, 0x79, 0x70, 0x65, 0x5f, 0x56, 0x69, 0x64, 0x65,
	0x6f, 0x10, 0x03, 0x12, 0x13, 0x0a, 0x0f, 0x41, 0x74, 0x74, 0x61, 0x63, 0x68, 0x54, 0x79, 0x70,
	0x65, 0x5f, 0x46, 0x69, 0x6c, 0x65, 0x10, 0x04, 0x12, 0x12, 0x0a, 0x0e, 0x41, 0x74, 0x74, 0x61,
	0x63, 0x68, 0x54, 0x79, 0x70, 0x65, 0x5f, 0x47, 0x65, 0x6f, 0x10, 0x05, 0x12, 0x14, 0x0a, 0x10,
	0x41, 0x74, 0x74, 0x61, 0x63, 0x68, 0x54, 0x79, 0x70, 0x65, 0x5f, 0x4f, 0x72, 0x64, 0x65, 0x72,
	0x10, 0x06, 0x12, 0x1a, 0x0a, 0x16, 0x41, 0x74, 0x74, 0x61, 0x63, 0x68, 0x54, 0x79, 0x70, 0x65,
	0x5f, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x10, 0x07, 0x42, 0x29,
	0x5a, 0x27, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x69, 0x61,
	0x6e, 0x6d, 0x69, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x73, 0x67, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_api_proto_msg_MsgTypeEnum_proto_rawDescOnce sync.Once
	file_api_proto_msg_MsgTypeEnum_proto_rawDescData = file_api_proto_msg_MsgTypeEnum_proto_rawDesc
)

func file_api_proto_msg_MsgTypeEnum_proto_rawDescGZIP() []byte {
	file_api_proto_msg_MsgTypeEnum_proto_rawDescOnce.Do(func() {
		file_api_proto_msg_MsgTypeEnum_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_proto_msg_MsgTypeEnum_proto_rawDescData)
	})
	return file_api_proto_msg_MsgTypeEnum_proto_rawDescData
}

var file_api_proto_msg_MsgTypeEnum_proto_enumTypes = make([]protoimpl.EnumInfo, 3)
var file_api_proto_msg_MsgTypeEnum_proto_goTypes = []interface{}{
	(MessageScene)(0), // 0: cloud.lianmi.im.msg.MessageScene
	(MessageType)(0),  // 1: cloud.lianmi.im.msg.MessageType
	(AttachType)(0),   // 2: cloud.lianmi.im.msg.AttachType
}
var file_api_proto_msg_MsgTypeEnum_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_api_proto_msg_MsgTypeEnum_proto_init() }
func file_api_proto_msg_MsgTypeEnum_proto_init() {
	if File_api_proto_msg_MsgTypeEnum_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_api_proto_msg_MsgTypeEnum_proto_rawDesc,
			NumEnums:      3,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_proto_msg_MsgTypeEnum_proto_goTypes,
		DependencyIndexes: file_api_proto_msg_MsgTypeEnum_proto_depIdxs,
		EnumInfos:         file_api_proto_msg_MsgTypeEnum_proto_enumTypes,
	}.Build()
	File_api_proto_msg_MsgTypeEnum_proto = out.File
	file_api_proto_msg_MsgTypeEnum_proto_rawDesc = nil
	file_api_proto_msg_MsgTypeEnum_proto_goTypes = nil
	file_api_proto_msg_MsgTypeEnum_proto_depIdxs = nil
}
