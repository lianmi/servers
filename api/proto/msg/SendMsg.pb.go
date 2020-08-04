// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.7.0
// source: api/proto/msg/SendMsg.proto

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

//2.6.1+2.6.4
//发送消息请求
//C2C消息发送接口
type SendMsgReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//通过type确定具体内容(视/音频文件、文件、图片、地理位置、音视频通话)
	//是否必填-否
	Attach string `protobuf:"bytes,1,opt,name=attach,proto3" json:"attach,omitempty"`
	//用于UI展示的文本信息
	//是否必填-否
	Body string `protobuf:"bytes,2,opt,name=body,proto3" json:"body,omitempty"`
	//客户端分配的消息ID，SDK生成的消息id,
	//在发送消息之后会返回给开发者,
	//开发者可以在发送消息的结果回调里面根据这个ID来判断相应消息的发送状态,
	//到底是发送成功了还是发送失败了, 然后根据此状态来更新页面的UI。
	//如果发送失败, 那么可以重新发送此消息，推荐UUID
	//是否必填-是
	IdClient string `protobuf:"bytes,3,opt,name=idClient,proto3" json:"idClient,omitempty"`
	//传输场景
	//是否必填-是
	Scene Scene `protobuf:"varint,4,opt,name=scene,proto3,enum=cc.lianmi.im.msg.Scene" json:"scene,omitempty"`
	//接收人ID
	//是否必填-是
	To string `protobuf:"bytes,5,opt,name=to,proto3" json:"to,omitempty"`
	//消息类型
	//是否必填-是
	Type MsgType `protobuf:"varint,6,opt,name=type,proto3,enum=cc.lianmi.im.msg.MsgType" json:"type,omitempty"`
	//客户端发送时间,Unix时间戳
	//是否必填-是
	UserUpdateTime uint64 `protobuf:"fixed64,7,opt,name=userUpdateTime,proto3" json:"userUpdateTime,omitempty"`
	//指定该消息接收设备(to登录的任一设备)
	//是否必填-否
	ToDeviceId string `protobuf:"bytes,8,opt,name=toDeviceId,proto3" json:"toDeviceId,omitempty"`
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

func (x *SendMsgReq) GetAttach() string {
	if x != nil {
		return x.Attach
	}
	return ""
}

func (x *SendMsgReq) GetBody() string {
	if x != nil {
		return x.Body
	}
	return ""
}

func (x *SendMsgReq) GetIdClient() string {
	if x != nil {
		return x.IdClient
	}
	return ""
}

func (x *SendMsgReq) GetScene() Scene {
	if x != nil {
		return x.Scene
	}
	return Scene_Sc_Undefined
}

func (x *SendMsgReq) GetTo() string {
	if x != nil {
		return x.To
	}
	return ""
}

func (x *SendMsgReq) GetType() MsgType {
	if x != nil {
		return x.Type
	}
	return MsgType_Mt_Undefined
}

func (x *SendMsgReq) GetUserUpdateTime() uint64 {
	if x != nil {
		return x.UserUpdateTime
	}
	return 0
}

func (x *SendMsgReq) GetToDeviceId() string {
	if x != nil {
		return x.ToDeviceId
	}
	return ""
}

//2.6.1+2.6.4
//发送消息响应
type SendMsgRsp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//消息客户端ID
	//是否必填-是
	IdClient string `protobuf:"bytes,1,opt,name=idClient,proto3" json:"idClient,omitempty"`
	//消息服务器ID
	//是否必填-是
	IdServer string `protobuf:"bytes,2,opt,name=idServer,proto3" json:"idServer,omitempty"`
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

func (x *SendMsgRsp) GetIdClient() string {
	if x != nil {
		return x.IdClient
	}
	return ""
}

func (x *SendMsgRsp) GetIdServer() string {
	if x != nil {
		return x.IdServer
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
	0x53, 0x65, 0x6e, 0x64, 0x4d, 0x73, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x10, 0x63,
	0x63, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x6d, 0x73, 0x67, 0x1a,
	0x17, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x73, 0x67, 0x2f, 0x4d,
	0x73, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x8a, 0x02, 0x0a, 0x0a, 0x53, 0x65, 0x6e,
	0x64, 0x4d, 0x73, 0x67, 0x52, 0x65, 0x71, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x74, 0x74, 0x61, 0x63,
	0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x12,
	0x12, 0x0a, 0x04, 0x62, 0x6f, 0x64, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x62,
	0x6f, 0x64, 0x79, 0x12, 0x1a, 0x0a, 0x08, 0x69, 0x64, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x69, 0x64, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x12,
	0x2d, 0x0a, 0x05, 0x73, 0x63, 0x65, 0x6e, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x17,
	0x2e, 0x63, 0x63, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x6d, 0x73,
	0x67, 0x2e, 0x53, 0x63, 0x65, 0x6e, 0x65, 0x52, 0x05, 0x73, 0x63, 0x65, 0x6e, 0x65, 0x12, 0x0e,
	0x0a, 0x02, 0x74, 0x6f, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x74, 0x6f, 0x12, 0x2d,
	0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x19, 0x2e, 0x63,
	0x63, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x6d, 0x73, 0x67, 0x2e,
	0x4d, 0x73, 0x67, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x26, 0x0a,
	0x0e, 0x75, 0x73, 0x65, 0x72, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x18,
	0x07, 0x20, 0x01, 0x28, 0x06, 0x52, 0x0e, 0x75, 0x73, 0x65, 0x72, 0x55, 0x70, 0x64, 0x61, 0x74,
	0x65, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x74, 0x6f, 0x44, 0x65, 0x76, 0x69, 0x63,
	0x65, 0x49, 0x64, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x74, 0x6f, 0x44, 0x65, 0x76,
	0x69, 0x63, 0x65, 0x49, 0x64, 0x22, 0x6a, 0x0a, 0x0a, 0x53, 0x65, 0x6e, 0x64, 0x4d, 0x73, 0x67,
	0x52, 0x73, 0x70, 0x12, 0x1a, 0x0a, 0x08, 0x69, 0x64, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x69, 0x64, 0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x12,
	0x1a, 0x0a, 0x08, 0x69, 0x64, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x69, 0x64, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x12, 0x10, 0x0a, 0x03, 0x73,
	0x65, 0x71, 0x18, 0x03, 0x20, 0x01, 0x28, 0x06, 0x52, 0x03, 0x73, 0x65, 0x71, 0x12, 0x12, 0x0a,
	0x04, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x06, 0x52, 0x04, 0x74, 0x69, 0x6d,
	0x65, 0x42, 0x29, 0x5a, 0x27, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x2f, 0x61,
	0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x73, 0x67, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
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
	(*SendMsgReq)(nil), // 0: cc.lianmi.im.msg.SendMsgReq
	(*SendMsgRsp)(nil), // 1: cc.lianmi.im.msg.SendMsgRsp
	(Scene)(0),         // 2: cc.lianmi.im.msg.Scene
	(MsgType)(0),       // 3: cc.lianmi.im.msg.MsgType
}
var file_api_proto_msg_SendMsg_proto_depIdxs = []int32{
	2, // 0: cc.lianmi.im.msg.SendMsgReq.scene:type_name -> cc.lianmi.im.msg.Scene
	3, // 1: cc.lianmi.im.msg.SendMsgReq.type:type_name -> cc.lianmi.im.msg.MsgType
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
	file_api_proto_msg_Msg_proto_init()
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
