// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.7.0
// source: api/proto/msg/SendMsg.proto

// 发送消息

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

// 发送消息 请求包
type SendMsgReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//传输场景
	//是否必填-是
	Scene MessageScene `protobuf:"varint,1,opt,name=scene,proto3,enum=cloud.lianmi.im.msg.MessageScene" json:"scene,omitempty"`
	//消息数据包的类型
	//是否必填-是
	Type MessageType `protobuf:"varint,2,opt,name=type,proto3,enum=cloud.lianmi.im.msg.MessageType" json:"type,omitempty"`
	//接受方的用户C2C 的时候 是 对方账号 / C2G 是群id
	//是否必填-是
	To string `protobuf:"bytes,3,opt,name=to,proto3" json:"to,omitempty"`
	// 客户端 生成 的 消息唯一id
	Uuid string `protobuf:"bytes,4,opt,name=uuid,proto3" json:"uuid,omitempty"`
	//消息体 服务端透传 ， 客户端 通过类型 拼接 对应的 数据
	Body []byte `protobuf:"bytes,5,opt,name=body,proto3" json:"body,omitempty"`
	//客户端发送时间,Unix时间戳
	//是否必填-是
	SendAt uint64 `protobuf:"fixed64,6,opt,name=sendAt,proto3" json:"sendAt,omitempty"`
	//指定该消息接收的设备 P2P 的时候 使用
	ToDeviceId string `protobuf:"bytes,7,opt,name=toDeviceId,proto3" json:"toDeviceId,omitempty"`
}

func (x *SendMsgReq) Reset() {
	*x = SendMsgReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_msg_SendMsg_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SendMsgReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SendMsgReq) ProtoMessage() {}

func (x *SendMsgReq) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_msg_SendMsg_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SendMsgReq.ProtoReflect.Descriptor instead.
func (*SendMsgReq) Descriptor() ([]byte, []int) {
	return file_api_proto_msg_SendMsg_proto_rawDescGZIP(), []int{0}
}

func (x *SendMsgReq) GetScene() MessageScene {
	if x != nil {
		return x.Scene
	}
	return MessageScene_MsgScene_Undefined
}

func (x *SendMsgReq) GetType() MessageType {
	if x != nil {
		return x.Type
	}
	return MessageType_MsgType_Undefined
}

func (x *SendMsgReq) GetTo() string {
	if x != nil {
		return x.To
	}
	return ""
}

func (x *SendMsgReq) GetUuid() string {
	if x != nil {
		return x.Uuid
	}
	return ""
}

func (x *SendMsgReq) GetBody() []byte {
	if x != nil {
		return x.Body
	}
	return nil
}

func (x *SendMsgReq) GetSendAt() uint64 {
	if x != nil {
		return x.SendAt
	}
	return 0
}

func (x *SendMsgReq) GetToDeviceId() string {
	if x != nil {
		return x.ToDeviceId
	}
	return ""
}

//发送消息响应
type SendMsgRsp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//消息客户端ID
	//是否必填-是
	Uuid string `protobuf:"bytes,1,opt,name=uuid,proto3" json:"uuid,omitempty"`
	//消息服务器ID
	//是否必填-是
	ServerMsgId string `protobuf:"bytes,2,opt,name=serverMsgId,proto3" json:"serverMsgId,omitempty"`
	//消息序号，单个会话内自然递增
	//是否必填-是
	Seq uint64 `protobuf:"fixed64,3,opt,name=seq,proto3" json:"seq,omitempty"`
	//消息服务器处理时间,Unix时间戳
	//是否必填-是
	Time uint64 `protobuf:"fixed64,4,opt,name=time,proto3" json:"time,omitempty"`
}

func (x *SendMsgRsp) Reset() {
	*x = SendMsgRsp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_msg_SendMsg_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SendMsgRsp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SendMsgRsp) ProtoMessage() {}

func (x *SendMsgRsp) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_msg_SendMsg_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SendMsgRsp.ProtoReflect.Descriptor instead.
func (*SendMsgRsp) Descriptor() ([]byte, []int) {
	return file_api_proto_msg_SendMsg_proto_rawDescGZIP(), []int{1}
}

func (x *SendMsgRsp) GetUuid() string {
	if x != nil {
		return x.Uuid
	}
	return ""
}

func (x *SendMsgRsp) GetServerMsgId() string {
	if x != nil {
		return x.ServerMsgId
	}
	return ""
}

func (x *SendMsgRsp) GetSeq() uint64 {
	if x != nil {
		return x.Seq
	}
	return 0
}

func (x *SendMsgRsp) GetTime() uint64 {
	if x != nil {
		return x.Time
	}
	return 0
}

var File_api_proto_msg_SendMsg_proto protoreflect.FileDescriptor

var file_api_proto_msg_SendMsg_proto_rawDesc = []byte{
	0x0a, 0x1b, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x73, 0x67, 0x2f,
	0x53, 0x65, 0x6e, 0x64, 0x4d, 0x73, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x13, 0x63,
	0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x6d,
	0x73, 0x67, 0x1a, 0x1f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x73,
	0x67, 0x2f, 0x4d, 0x73, 0x67, 0x54, 0x79, 0x70, 0x65, 0x45, 0x6e, 0x75, 0x6d, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0xeb, 0x01, 0x0a, 0x0a, 0x53, 0x65, 0x6e, 0x64, 0x4d, 0x73, 0x67, 0x52,
	0x65, 0x71, 0x12, 0x37, 0x0a, 0x05, 0x73, 0x63, 0x65, 0x6e, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x21, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69,
	0x2e, 0x69, 0x6d, 0x2e, 0x6d, 0x73, 0x67, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x53,
	0x63, 0x65, 0x6e, 0x65, 0x52, 0x05, 0x73, 0x63, 0x65, 0x6e, 0x65, 0x12, 0x34, 0x0a, 0x04, 0x74,
	0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x20, 0x2e, 0x63, 0x6c, 0x6f, 0x75,
	0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x6d, 0x73, 0x67, 0x2e,
	0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70,
	0x65, 0x12, 0x0e, 0x0a, 0x02, 0x74, 0x6f, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x74,
	0x6f, 0x12, 0x12, 0x0a, 0x04, 0x75, 0x75, 0x69, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x75, 0x75, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x62, 0x6f, 0x64, 0x79, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x04, 0x62, 0x6f, 0x64, 0x79, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x65, 0x6e,
	0x64, 0x41, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x06, 0x52, 0x06, 0x73, 0x65, 0x6e, 0x64, 0x41,
	0x74, 0x12, 0x1e, 0x0a, 0x0a, 0x74, 0x6f, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x49, 0x64, 0x18,
	0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x74, 0x6f, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x49,
	0x64, 0x22, 0x68, 0x0a, 0x0a, 0x53, 0x65, 0x6e, 0x64, 0x4d, 0x73, 0x67, 0x52, 0x73, 0x70, 0x12,
	0x12, 0x0a, 0x04, 0x75, 0x75, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x75,
	0x75, 0x69, 0x64, 0x12, 0x20, 0x0a, 0x0b, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x4d, 0x73, 0x67,
	0x49, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x4d, 0x73, 0x67, 0x49, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x73, 0x65, 0x71, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x06, 0x52, 0x03, 0x73, 0x65, 0x71, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x69, 0x6d, 0x65, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x06, 0x52, 0x04, 0x74, 0x69, 0x6d, 0x65, 0x42, 0x29, 0x5a, 0x27, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69,
	0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2f, 0x6d, 0x73, 0x67, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_proto_msg_SendMsg_proto_rawDescOnce sync.Once
	file_api_proto_msg_SendMsg_proto_rawDescData = file_api_proto_msg_SendMsg_proto_rawDesc
)

func file_api_proto_msg_SendMsg_proto_rawDescGZIP() []byte {
	file_api_proto_msg_SendMsg_proto_rawDescOnce.Do(func() {
		file_api_proto_msg_SendMsg_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_proto_msg_SendMsg_proto_rawDescData)
	})
	return file_api_proto_msg_SendMsg_proto_rawDescData
}

var file_api_proto_msg_SendMsg_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_api_proto_msg_SendMsg_proto_goTypes = []interface{}{
	(*SendMsgReq)(nil), // 0: cloud.lianmi.im.msg.SendMsgReq
	(*SendMsgRsp)(nil), // 1: cloud.lianmi.im.msg.SendMsgRsp
	(MessageScene)(0),  // 2: cloud.lianmi.im.msg.MessageScene
	(MessageType)(0),   // 3: cloud.lianmi.im.msg.MessageType
}
var file_api_proto_msg_SendMsg_proto_depIdxs = []int32{
	2, // 0: cloud.lianmi.im.msg.SendMsgReq.scene:type_name -> cloud.lianmi.im.msg.MessageScene
	3, // 1: cloud.lianmi.im.msg.SendMsgReq.type:type_name -> cloud.lianmi.im.msg.MessageType
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_api_proto_msg_SendMsg_proto_init() }
func file_api_proto_msg_SendMsg_proto_init() {
	if File_api_proto_msg_SendMsg_proto != nil {
		return
	}
	file_api_proto_msg_MsgTypeEnum_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_api_proto_msg_SendMsg_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SendMsgReq); i {
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
		file_api_proto_msg_SendMsg_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SendMsgRsp); i {
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
			RawDescriptor: file_api_proto_msg_SendMsg_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_proto_msg_SendMsg_proto_goTypes,
		DependencyIndexes: file_api_proto_msg_SendMsg_proto_depIdxs,
		MessageInfos:      file_api_proto_msg_SendMsg_proto_msgTypes,
	}.Build()
	File_api_proto_msg_SendMsg_proto = out.File
	file_api_proto_msg_SendMsg_proto_rawDesc = nil
	file_api_proto_msg_SendMsg_proto_goTypes = nil
	file_api_proto_msg_SendMsg_proto_depIdxs = nil
}
