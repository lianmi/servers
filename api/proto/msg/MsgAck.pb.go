// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.7.0
// source: api/proto/msg/MsgAck.proto

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

//向服务端发送确认消息送达的请求
type MsgAckReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// 消息的传输场景
	Scene MessageScene `protobuf:"varint,1,opt,name=scene,proto3,enum=cloud.lianmi.im.msg.MessageScene" json:"scene,omitempty"`
	// 消息类型
	Type MessageType `protobuf:"varint,2,opt,name=type,proto3,enum=cloud.lianmi.im.msg.MessageType" json:"type,omitempty"`
	// 消息id 服务器生成 全局唯一
	ServerMsgId string `protobuf:"bytes,3,opt,name=serverMsgId,proto3" json:"serverMsgId,omitempty"`
	// 客户端生成的uuid
	Uuid string `protobuf:"bytes,4,opt,name=uuid,proto3" json:"uuid,omitempty"`
	// 消息序号
	Seq uint64 `protobuf:"fixed64,5,opt,name=seq,proto3" json:"seq,omitempty"`
	// 消息状态
	Status MessageStatus `protobuf:"varint,6,opt,name=status,proto3,enum=cloud.lianmi.im.msg.MessageStatus" json:"status,omitempty"`
	// 消息的发送方
	From string `protobuf:"bytes,7,opt,name=from,proto3" json:"from,omitempty"`
	// 消息的接受方
	To string `protobuf:"bytes,8,opt,name=to,proto3" json:"to,omitempty"`
	//系统消息收到时间，Unix时间戳，更新本地timetag表的systemMsgAt字段
	//是否必填-是
	TimeTag uint64 `protobuf:"fixed64,9,opt,name=timeTag,proto3" json:"timeTag,omitempty"`
}

func (x *MsgAckReq) Reset() {
	*x = MsgAckReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_msg_MsgAck_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MsgAckReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MsgAckReq) ProtoMessage() {}

func (x *MsgAckReq) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_msg_MsgAck_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MsgAckReq.ProtoReflect.Descriptor instead.
func (*MsgAckReq) Descriptor() ([]byte, []int) {
	return file_api_proto_msg_MsgAck_proto_rawDescGZIP(), []int{0}
}

func (x *MsgAckReq) GetScene() MessageScene {
	if x != nil {
		return x.Scene
	}
	return MessageScene_MsgScene_Undefined
}

func (x *MsgAckReq) GetType() MessageType {
	if x != nil {
		return x.Type
	}
	return MessageType_MsgType_Undefined
}

func (x *MsgAckReq) GetServerMsgId() string {
	if x != nil {
		return x.ServerMsgId
	}
	return ""
}

func (x *MsgAckReq) GetUuid() string {
	if x != nil {
		return x.Uuid
	}
	return ""
}

func (x *MsgAckReq) GetSeq() uint64 {
	if x != nil {
		return x.Seq
	}
	return 0
}

func (x *MsgAckReq) GetStatus() MessageStatus {
	if x != nil {
		return x.Status
	}
	return MessageStatus_MOS_UDEFINE
}

func (x *MsgAckReq) GetFrom() string {
	if x != nil {
		return x.From
	}
	return ""
}

func (x *MsgAckReq) GetTo() string {
	if x != nil {
		return x.To
	}
	return ""
}

func (x *MsgAckReq) GetTimeTag() uint64 {
	if x != nil {
		return x.TimeTag
	}
	return 0
}

//code=200
type MsgAckRsp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *MsgAckRsp) Reset() {
	*x = MsgAckRsp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_msg_MsgAck_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MsgAckRsp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MsgAckRsp) ProtoMessage() {}

func (x *MsgAckRsp) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_msg_MsgAck_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MsgAckRsp.ProtoReflect.Descriptor instead.
func (*MsgAckRsp) Descriptor() ([]byte, []int) {
	return file_api_proto_msg_MsgAck_proto_rawDescGZIP(), []int{1}
}

var File_api_proto_msg_MsgAck_proto protoreflect.FileDescriptor

var file_api_proto_msg_MsgAck_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x73, 0x67, 0x2f,
	0x4d, 0x73, 0x67, 0x41, 0x63, 0x6b, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x13, 0x63, 0x6c,
	0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x6d, 0x73,
	0x67, 0x1a, 0x1f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x73, 0x67,
	0x2f, 0x4d, 0x73, 0x67, 0x54, 0x79, 0x70, 0x65, 0x45, 0x6e, 0x75, 0x6d, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x22, 0xbc, 0x02, 0x0a, 0x09, 0x4d, 0x73, 0x67, 0x41, 0x63, 0x6b, 0x52, 0x65, 0x71,
	0x12, 0x37, 0x0a, 0x05, 0x73, 0x63, 0x65, 0x6e, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32,
	0x21, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69,
	0x6d, 0x2e, 0x6d, 0x73, 0x67, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x53, 0x63, 0x65,
	0x6e, 0x65, 0x52, 0x05, 0x73, 0x63, 0x65, 0x6e, 0x65, 0x12, 0x34, 0x0a, 0x04, 0x74, 0x79, 0x70,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x20, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e,
	0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x6d, 0x73, 0x67, 0x2e, 0x4d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12,
	0x20, 0x0a, 0x0b, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x4d, 0x73, 0x67, 0x49, 0x64, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x4d, 0x73, 0x67, 0x49,
	0x64, 0x12, 0x12, 0x0a, 0x04, 0x75, 0x75, 0x69, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x75, 0x75, 0x69, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x73, 0x65, 0x71, 0x18, 0x05, 0x20, 0x01,
	0x28, 0x06, 0x52, 0x03, 0x73, 0x65, 0x71, 0x12, 0x3a, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x22, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e,
	0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x6d, 0x73, 0x67, 0x2e, 0x4d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x66, 0x72, 0x6f, 0x6d, 0x18, 0x07, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x66, 0x72, 0x6f, 0x6d, 0x12, 0x0e, 0x0a, 0x02, 0x74, 0x6f, 0x18, 0x08, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x02, 0x74, 0x6f, 0x12, 0x18, 0x0a, 0x07, 0x74, 0x69, 0x6d, 0x65, 0x54,
	0x61, 0x67, 0x18, 0x09, 0x20, 0x01, 0x28, 0x06, 0x52, 0x07, 0x74, 0x69, 0x6d, 0x65, 0x54, 0x61,
	0x67, 0x22, 0x0b, 0x0a, 0x09, 0x4d, 0x73, 0x67, 0x41, 0x63, 0x6b, 0x52, 0x73, 0x70, 0x42, 0x29,
	0x5a, 0x27, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x69, 0x61,
	0x6e, 0x6d, 0x69, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x73, 0x67, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_api_proto_msg_MsgAck_proto_rawDescOnce sync.Once
	file_api_proto_msg_MsgAck_proto_rawDescData = file_api_proto_msg_MsgAck_proto_rawDesc
)

func file_api_proto_msg_MsgAck_proto_rawDescGZIP() []byte {
	file_api_proto_msg_MsgAck_proto_rawDescOnce.Do(func() {
		file_api_proto_msg_MsgAck_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_proto_msg_MsgAck_proto_rawDescData)
	})
	return file_api_proto_msg_MsgAck_proto_rawDescData
}

var file_api_proto_msg_MsgAck_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_api_proto_msg_MsgAck_proto_goTypes = []interface{}{
	(*MsgAckReq)(nil),  // 0: cloud.lianmi.im.msg.MsgAckReq
	(*MsgAckRsp)(nil),  // 1: cloud.lianmi.im.msg.MsgAckRsp
	(MessageScene)(0),  // 2: cloud.lianmi.im.msg.MessageScene
	(MessageType)(0),   // 3: cloud.lianmi.im.msg.MessageType
	(MessageStatus)(0), // 4: cloud.lianmi.im.msg.MessageStatus
}
var file_api_proto_msg_MsgAck_proto_depIdxs = []int32{
	2, // 0: cloud.lianmi.im.msg.MsgAckReq.scene:type_name -> cloud.lianmi.im.msg.MessageScene
	3, // 1: cloud.lianmi.im.msg.MsgAckReq.type:type_name -> cloud.lianmi.im.msg.MessageType
	4, // 2: cloud.lianmi.im.msg.MsgAckReq.status:type_name -> cloud.lianmi.im.msg.MessageStatus
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_api_proto_msg_MsgAck_proto_init() }
func file_api_proto_msg_MsgAck_proto_init() {
	if File_api_proto_msg_MsgAck_proto != nil {
		return
	}
	file_api_proto_msg_MsgTypeEnum_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_api_proto_msg_MsgAck_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MsgAckReq); i {
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
		file_api_proto_msg_MsgAck_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MsgAckRsp); i {
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
			RawDescriptor: file_api_proto_msg_MsgAck_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_proto_msg_MsgAck_proto_goTypes,
		DependencyIndexes: file_api_proto_msg_MsgAck_proto_depIdxs,
		MessageInfos:      file_api_proto_msg_MsgAck_proto_msgTypes,
	}.Build()
	File_api_proto_msg_MsgAck_proto = out.File
	file_api_proto_msg_MsgAck_proto_rawDesc = nil
	file_api_proto_msg_MsgAck_proto_goTypes = nil
	file_api_proto_msg_MsgAck_proto_depIdxs = nil
}
