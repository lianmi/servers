// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.7.0
// source: api/proto/msg/RecvTeamMsg.proto

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

//2.6.5. 接收群组消息事件
//接收群消息摘要
type RecvTeamMsgEvent struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//用于UI展示的文本信息
	//是否必须-否
	Body string `protobuf:"bytes,1,opt,name=body,proto3" json:"body,omitempty"`
	//群组ID
	//是否必须-是
	TeamId string `protobuf:"bytes,2,opt,name=teamId,proto3" json:"teamId,omitempty"`
	//消息来源,用户ID
	//是否必须-是
	From string `protobuf:"bytes,3,opt,name=from,proto3" json:"from,omitempty"`
	//发消息，用户昵称
	//是否必须-是
	FromNick string `protobuf:"bytes,4,opt,name=fromNick,proto3" json:"fromNick,omitempty"`
	//消息类型
	//是否必填-是
	Type MsgType `protobuf:"varint,5,opt,name=type,proto3,enum=cloud.lianmi.im.msg.MsgType" json:"type,omitempty"`
	//服务器分配的消息ID
	//是否必须-是
	IdServer string `protobuf:"bytes,6,opt,name=idServer,proto3" json:"idServer,omitempty"`
	//消息序号，单个会话内自然递增
	//是否必填-是
	Seq uint64 `protobuf:"fixed64,7,opt,name=seq,proto3" json:"seq,omitempty"`
	//客户端分配的消息ID，SDK生成的消息id, 在发送消息之后会返回给开发者,
	//开发者可以在发送消息的结果回调里面根据这个ID来判断相应消息的发送状态,
	//到底是发送成功了还是发送失败了, 然后根据此状态来更
	//新页面的UI。如果发送失败, 那么可以重新发送此消息，推荐UUID
	//是否必须-是
	IdClient string `protobuf:"bytes,8,opt,name=idClient,proto3" json:"idClient,omitempty"`
	//服务器处理消息时间,Unix时间戳
	//是否必须-是
	Time uint64 `protobuf:"fixed64,9,opt,name=time,proto3" json:"time,omitempty"`
}

func (x *RecvTeamMsgEvent) Reset() {
	*x = RecvTeamMsgEvent{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_msg_RecvTeamMsg_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RecvTeamMsgEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RecvTeamMsgEvent) ProtoMessage() {}

func (x *RecvTeamMsgEvent) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_msg_RecvTeamMsg_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RecvTeamMsgEvent.ProtoReflect.Descriptor instead.
func (*RecvTeamMsgEvent) Descriptor() ([]byte, []int) {
	return file_api_proto_msg_RecvTeamMsg_proto_rawDescGZIP(), []int{0}
}

func (x *RecvTeamMsgEvent) GetBody() string {
	if x != nil {
		return x.Body
	}
	return ""
}

func (x *RecvTeamMsgEvent) GetTeamId() string {
	if x != nil {
		return x.TeamId
	}
	return ""
}

func (x *RecvTeamMsgEvent) GetFrom() string {
	if x != nil {
		return x.From
	}
	return ""
}

func (x *RecvTeamMsgEvent) GetFromNick() string {
	if x != nil {
		return x.FromNick
	}
	return ""
}

func (x *RecvTeamMsgEvent) GetType() MsgType {
	if x != nil {
		return x.Type
	}
	return MsgType_Mt_Undefined
}

func (x *RecvTeamMsgEvent) GetIdServer() string {
	if x != nil {
		return x.IdServer
	}
	return ""
}

func (x *RecvTeamMsgEvent) GetSeq() uint64 {
	if x != nil {
		return x.Seq
	}
	return 0
}

func (x *RecvTeamMsgEvent) GetIdClient() string {
	if x != nil {
		return x.IdClient
	}
	return ""
}

func (x *RecvTeamMsgEvent) GetTime() uint64 {
	if x != nil {
		return x.Time
	}
	return 0
}

var File_api_proto_msg_RecvTeamMsg_proto protoreflect.FileDescriptor

var file_api_proto_msg_RecvTeamMsg_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x73, 0x67, 0x2f,
	0x52, 0x65, 0x63, 0x76, 0x54, 0x65, 0x61, 0x6d, 0x4d, 0x73, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x13, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e,
	0x69, 0x6d, 0x2e, 0x6d, 0x73, 0x67, 0x1a, 0x17, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x2f, 0x6d, 0x73, 0x67, 0x2f, 0x4d, 0x73, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0xfe, 0x01, 0x0a, 0x10, 0x52, 0x65, 0x63, 0x76, 0x54, 0x65, 0x61, 0x6d, 0x4d, 0x73, 0x67, 0x45,
	0x76, 0x65, 0x6e, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x62, 0x6f, 0x64, 0x79, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x62, 0x6f, 0x64, 0x79, 0x12, 0x16, 0x0a, 0x06, 0x74, 0x65, 0x61, 0x6d,
	0x49, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x74, 0x65, 0x61, 0x6d, 0x49, 0x64,
	0x12, 0x12, 0x0a, 0x04, 0x66, 0x72, 0x6f, 0x6d, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x66, 0x72, 0x6f, 0x6d, 0x12, 0x1a, 0x0a, 0x08, 0x66, 0x72, 0x6f, 0x6d, 0x4e, 0x69, 0x63, 0x6b,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x66, 0x72, 0x6f, 0x6d, 0x4e, 0x69, 0x63, 0x6b,
	0x12, 0x30, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1c,
	0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d,
	0x2e, 0x6d, 0x73, 0x67, 0x2e, 0x4d, 0x73, 0x67, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79,
	0x70, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x69, 0x64, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x18, 0x06,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x69, 0x64, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x12, 0x10,
	0x0a, 0x03, 0x73, 0x65, 0x71, 0x18, 0x07, 0x20, 0x01, 0x28, 0x06, 0x52, 0x03, 0x73, 0x65, 0x71,
	0x12, 0x1a, 0x0a, 0x08, 0x69, 0x64, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x18, 0x08, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x08, 0x69, 0x64, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x12, 0x12, 0x0a, 0x04,
	0x74, 0x69, 0x6d, 0x65, 0x18, 0x09, 0x20, 0x01, 0x28, 0x06, 0x52, 0x04, 0x74, 0x69, 0x6d, 0x65,
	0x42, 0x29, 0x5a, 0x27, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c,
	0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x2f, 0x61, 0x70,
	0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x73, 0x67, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_api_proto_msg_RecvTeamMsg_proto_rawDescOnce sync.Once
	file_api_proto_msg_RecvTeamMsg_proto_rawDescData = file_api_proto_msg_RecvTeamMsg_proto_rawDesc
)

func file_api_proto_msg_RecvTeamMsg_proto_rawDescGZIP() []byte {
	file_api_proto_msg_RecvTeamMsg_proto_rawDescOnce.Do(func() {
		file_api_proto_msg_RecvTeamMsg_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_proto_msg_RecvTeamMsg_proto_rawDescData)
	})
	return file_api_proto_msg_RecvTeamMsg_proto_rawDescData
}

var file_api_proto_msg_RecvTeamMsg_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_api_proto_msg_RecvTeamMsg_proto_goTypes = []interface{}{
	(*RecvTeamMsgEvent)(nil), // 0: cloud.lianmi.im.msg.RecvTeamMsgEvent
	(MsgType)(0),             // 1: cloud.lianmi.im.msg.MsgType
}
var file_api_proto_msg_RecvTeamMsg_proto_depIdxs = []int32{
	1, // 0: cloud.lianmi.im.msg.RecvTeamMsgEvent.type:type_name -> cloud.lianmi.im.msg.MsgType
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_api_proto_msg_RecvTeamMsg_proto_init() }
func file_api_proto_msg_RecvTeamMsg_proto_init() {
	if File_api_proto_msg_RecvTeamMsg_proto != nil {
		return
	}
	file_api_proto_msg_Msg_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_api_proto_msg_RecvTeamMsg_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RecvTeamMsgEvent); i {
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
			RawDescriptor: file_api_proto_msg_RecvTeamMsg_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_proto_msg_RecvTeamMsg_proto_goTypes,
		DependencyIndexes: file_api_proto_msg_RecvTeamMsg_proto_depIdxs,
		MessageInfos:      file_api_proto_msg_RecvTeamMsg_proto_msgTypes,
	}.Build()
	File_api_proto_msg_RecvTeamMsg_proto = out.File
	file_api_proto_msg_RecvTeamMsg_proto_rawDesc = nil
	file_api_proto_msg_RecvTeamMsg_proto_goTypes = nil
	file_api_proto_msg_RecvTeamMsg_proto_depIdxs = nil
}
