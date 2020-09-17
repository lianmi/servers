// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.7.0
// source: api/proto/msg/RecvCancelMsgEvent.proto

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

// 撤回消息事件
type RecvCancelMsgEventRsp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//传输场景
	//是否必填-是
	Scene MessageScene `protobuf:"varint,1,opt,name=scene,proto3,enum=cloud.lianmi.im.msg.MessageScene" json:"scene,omitempty"`
	//消息数据包的类型
	//是否必填-是
	Type MessageType `protobuf:"varint,2,opt,name=type,proto3,enum=cloud.lianmi.im.msg.MessageType" json:"type,omitempty"`
	//被撤销的消息发送方
	From string `protobuf:"bytes,3,opt,name=from,proto3" json:"from,omitempty"`
	//消息是发给谁的
	To string `protobuf:"bytes,4,opt,name=to,proto3" json:"to,omitempty"`
	//要撤销的消息的由服务器分配的消息id
	//是否必填-是
	ServerMsgId string `protobuf:"bytes,5,opt,name=serverMsgId,proto3" json:"serverMsgId,omitempty"`
}

func (x *RecvCancelMsgEventRsp) Reset() {
	*x = RecvCancelMsgEventRsp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_msg_RecvCancelMsgEvent_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RecvCancelMsgEventRsp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RecvCancelMsgEventRsp) ProtoMessage() {}

func (x *RecvCancelMsgEventRsp) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_msg_RecvCancelMsgEvent_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RecvCancelMsgEventRsp.ProtoReflect.Descriptor instead.
func (*RecvCancelMsgEventRsp) Descriptor() ([]byte, []int) {
	return file_api_proto_msg_RecvCancelMsgEvent_proto_rawDescGZIP(), []int{0}
}

func (x *RecvCancelMsgEventRsp) GetScene() MessageScene {
	if x != nil {
		return x.Scene
	}
	return MessageScene_MsgScene_Undefined
}

func (x *RecvCancelMsgEventRsp) GetType() MessageType {
	if x != nil {
		return x.Type
	}
	return MessageType_MsgType_Undefined
}

func (x *RecvCancelMsgEventRsp) GetFrom() string {
	if x != nil {
		return x.From
	}
	return ""
}

func (x *RecvCancelMsgEventRsp) GetTo() string {
	if x != nil {
		return x.To
	}
	return ""
}

func (x *RecvCancelMsgEventRsp) GetServerMsgId() string {
	if x != nil {
		return x.ServerMsgId
	}
	return ""
}

var File_api_proto_msg_RecvCancelMsgEvent_proto protoreflect.FileDescriptor

var file_api_proto_msg_RecvCancelMsgEvent_proto_rawDesc = []byte{
	0x0a, 0x26, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x73, 0x67, 0x2f,
	0x52, 0x65, 0x63, 0x76, 0x43, 0x61, 0x6e, 0x63, 0x65, 0x6c, 0x4d, 0x73, 0x67, 0x45, 0x76, 0x65,
	0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x13, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e,
	0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x6d, 0x73, 0x67, 0x1a, 0x1f, 0x61,
	0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x73, 0x67, 0x2f, 0x4d, 0x73, 0x67,
	0x54, 0x79, 0x70, 0x65, 0x45, 0x6e, 0x75, 0x6d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xcc,
	0x01, 0x0a, 0x15, 0x52, 0x65, 0x63, 0x76, 0x43, 0x61, 0x6e, 0x63, 0x65, 0x6c, 0x4d, 0x73, 0x67,
	0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x73, 0x70, 0x12, 0x37, 0x0a, 0x05, 0x73, 0x63, 0x65, 0x6e,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x21, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e,
	0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x6d, 0x73, 0x67, 0x2e, 0x4d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x53, 0x63, 0x65, 0x6e, 0x65, 0x52, 0x05, 0x73, 0x63, 0x65, 0x6e,
	0x65, 0x12, 0x34, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32,
	0x20, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69,
	0x6d, 0x2e, 0x6d, 0x73, 0x67, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x54, 0x79, 0x70,
	0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x66, 0x72, 0x6f, 0x6d, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x66, 0x72, 0x6f, 0x6d, 0x12, 0x0e, 0x0a, 0x02, 0x74,
	0x6f, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x74, 0x6f, 0x12, 0x20, 0x0a, 0x0b, 0x73,
	0x65, 0x72, 0x76, 0x65, 0x72, 0x4d, 0x73, 0x67, 0x49, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0b, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x4d, 0x73, 0x67, 0x49, 0x64, 0x42, 0x29, 0x5a,
	0x27, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x69, 0x61, 0x6e,
	0x6d, 0x69, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x73, 0x67, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_proto_msg_RecvCancelMsgEvent_proto_rawDescOnce sync.Once
	file_api_proto_msg_RecvCancelMsgEvent_proto_rawDescData = file_api_proto_msg_RecvCancelMsgEvent_proto_rawDesc
)

func file_api_proto_msg_RecvCancelMsgEvent_proto_rawDescGZIP() []byte {
	file_api_proto_msg_RecvCancelMsgEvent_proto_rawDescOnce.Do(func() {
		file_api_proto_msg_RecvCancelMsgEvent_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_proto_msg_RecvCancelMsgEvent_proto_rawDescData)
	})
	return file_api_proto_msg_RecvCancelMsgEvent_proto_rawDescData
}

var file_api_proto_msg_RecvCancelMsgEvent_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_api_proto_msg_RecvCancelMsgEvent_proto_goTypes = []interface{}{
	(*RecvCancelMsgEventRsp)(nil), // 0: cloud.lianmi.im.msg.RecvCancelMsgEventRsp
	(MessageScene)(0),             // 1: cloud.lianmi.im.msg.MessageScene
	(MessageType)(0),              // 2: cloud.lianmi.im.msg.MessageType
}
var file_api_proto_msg_RecvCancelMsgEvent_proto_depIdxs = []int32{
	1, // 0: cloud.lianmi.im.msg.RecvCancelMsgEventRsp.scene:type_name -> cloud.lianmi.im.msg.MessageScene
	2, // 1: cloud.lianmi.im.msg.RecvCancelMsgEventRsp.type:type_name -> cloud.lianmi.im.msg.MessageType
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_api_proto_msg_RecvCancelMsgEvent_proto_init() }
func file_api_proto_msg_RecvCancelMsgEvent_proto_init() {
	if File_api_proto_msg_RecvCancelMsgEvent_proto != nil {
		return
	}
	file_api_proto_msg_MsgTypeEnum_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_api_proto_msg_RecvCancelMsgEvent_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RecvCancelMsgEventRsp); i {
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
			RawDescriptor: file_api_proto_msg_RecvCancelMsgEvent_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_proto_msg_RecvCancelMsgEvent_proto_goTypes,
		DependencyIndexes: file_api_proto_msg_RecvCancelMsgEvent_proto_depIdxs,
		MessageInfos:      file_api_proto_msg_RecvCancelMsgEvent_proto_msgTypes,
	}.Build()
	File_api_proto_msg_RecvCancelMsgEvent_proto = out.File
	file_api_proto_msg_RecvCancelMsgEvent_proto_rawDesc = nil
	file_api_proto_msg_RecvCancelMsgEvent_proto_goTypes = nil
	file_api_proto_msg_RecvCancelMsgEvent_proto_depIdxs = nil
}
