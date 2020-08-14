// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.7.0
// source: api/proto/syn/Sync.proto

package syn

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

//API描述
//客户端登录成功后，将本地维护的各模块的最后更新时间通过sync同步至服务器端，
//如果服务端用户资料更新时间与本地时间存在差异，im服务器则通过该指令将完整用户信息push至客户端。
//使用场景
//登录以及断线重连（清空浏览器缓存localstorage,给对方方消息,sync并不会触发,因为数据库不存在需要执行报错）
//清空本地所有缓存后,重新登录,则执行所有的同步,timetag为0, 包括:syncProfile syncFriendUsers
//API  C2S
type SyncReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//个人信息，触发同步当前用户资料事件事件
	//是否必填-是
	MyInfoAt uint64 `protobuf:"fixed64,1,opt,name=myInfoAt,proto3" json:"myInfoAt,omitempty"`
	//好友关系列表，触发好友列表同步事件事件
	//是否必填-是
	FriendsAt uint64 `protobuf:"fixed64,2,opt,name=friendsAt,proto3" json:"friendsAt,omitempty"`
	//好友用户信息，触发好友信息同步事件事件
	//是否必填-是
	FriendUsersAt uint64 `protobuf:"fixed64,3,opt,name=friendUsersAt,proto3" json:"friendUsersAt,omitempty"`
	//群组信息，触发同步群组事件事件
	//是否必填-是
	TeamsAt uint64 `protobuf:"fixed64,4,opt,name=teamsAt,proto3" json:"teamsAt,omitempty"`
	//是否必填-是
	TagsAt uint64 `protobuf:"fixed64,5,opt,name=tagsAt,proto3" json:"tagsAt,omitempty"`
	//系统消息最后同步时间戳，触发同步系统离线消息事件
	//是否必填-是
	SystemMsgAt uint64 `protobuf:"fixed64,6,opt,name=systemMsgAt,proto3" json:"systemMsgAt,omitempty"`
}

func (x *SyncReq) Reset() {
	*x = SyncReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_syn_Sync_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SyncReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SyncReq) ProtoMessage() {}

func (x *SyncReq) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_syn_Sync_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SyncReq.ProtoReflect.Descriptor instead.
func (*SyncReq) Descriptor() ([]byte, []int) {
	return file_api_proto_syn_Sync_proto_rawDescGZIP(), []int{0}
}

func (x *SyncReq) GetMyInfoAt() uint64 {
	if x != nil {
		return x.MyInfoAt
	}
	return 0
}

func (x *SyncReq) GetFriendsAt() uint64 {
	if x != nil {
		return x.FriendsAt
	}
	return 0
}

func (x *SyncReq) GetFriendUsersAt() uint64 {
	if x != nil {
		return x.FriendUsersAt
	}
	return 0
}

func (x *SyncReq) GetTeamsAt() uint64 {
	if x != nil {
		return x.TeamsAt
	}
	return 0
}

func (x *SyncReq) GetTagsAt() uint64 {
	if x != nil {
		return x.TagsAt
	}
	return 0
}

func (x *SyncReq) GetSystemMsgAt() uint64 {
	if x != nil {
		return x.SystemMsgAt
	}
	return 0
}

//只包含状态码，无内容载体
type SyncRsp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *SyncRsp) Reset() {
	*x = SyncRsp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_syn_Sync_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SyncRsp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SyncRsp) ProtoMessage() {}

func (x *SyncRsp) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_syn_Sync_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SyncRsp.ProtoReflect.Descriptor instead.
func (*SyncRsp) Descriptor() ([]byte, []int) {
	return file_api_proto_syn_Sync_proto_rawDescGZIP(), []int{1}
}

var File_api_proto_syn_Sync_proto protoreflect.FileDescriptor

var file_api_proto_syn_Sync_proto_rawDesc = []byte{
	0x0a, 0x18, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x73, 0x79, 0x6e, 0x2f,
	0x53, 0x79, 0x6e, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x13, 0x63, 0x6c, 0x6f, 0x75,
	0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x73, 0x79, 0x6e, 0x22,
	0xbd, 0x01, 0x0a, 0x07, 0x53, 0x79, 0x6e, 0x63, 0x52, 0x65, 0x71, 0x12, 0x1a, 0x0a, 0x08, 0x6d,
	0x79, 0x49, 0x6e, 0x66, 0x6f, 0x41, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x06, 0x52, 0x08, 0x6d,
	0x79, 0x49, 0x6e, 0x66, 0x6f, 0x41, 0x74, 0x12, 0x1c, 0x0a, 0x09, 0x66, 0x72, 0x69, 0x65, 0x6e,
	0x64, 0x73, 0x41, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x06, 0x52, 0x09, 0x66, 0x72, 0x69, 0x65,
	0x6e, 0x64, 0x73, 0x41, 0x74, 0x12, 0x24, 0x0a, 0x0d, 0x66, 0x72, 0x69, 0x65, 0x6e, 0x64, 0x55,
	0x73, 0x65, 0x72, 0x73, 0x41, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x06, 0x52, 0x0d, 0x66, 0x72,
	0x69, 0x65, 0x6e, 0x64, 0x55, 0x73, 0x65, 0x72, 0x73, 0x41, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x74,
	0x65, 0x61, 0x6d, 0x73, 0x41, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x06, 0x52, 0x07, 0x74, 0x65,
	0x61, 0x6d, 0x73, 0x41, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x74, 0x61, 0x67, 0x73, 0x41, 0x74, 0x18,
	0x05, 0x20, 0x01, 0x28, 0x06, 0x52, 0x06, 0x74, 0x61, 0x67, 0x73, 0x41, 0x74, 0x12, 0x20, 0x0a,
	0x0b, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x4d, 0x73, 0x67, 0x41, 0x74, 0x18, 0x06, 0x20, 0x01,
	0x28, 0x06, 0x52, 0x0b, 0x73, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x4d, 0x73, 0x67, 0x41, 0x74, 0x22,
	0x09, 0x0a, 0x07, 0x53, 0x79, 0x6e, 0x63, 0x52, 0x73, 0x70, 0x42, 0x29, 0x5a, 0x27, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2f,
	0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x2f, 0x73, 0x79, 0x6e, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_proto_syn_Sync_proto_rawDescOnce sync.Once
	file_api_proto_syn_Sync_proto_rawDescData = file_api_proto_syn_Sync_proto_rawDesc
)

func file_api_proto_syn_Sync_proto_rawDescGZIP() []byte {
	file_api_proto_syn_Sync_proto_rawDescOnce.Do(func() {
		file_api_proto_syn_Sync_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_proto_syn_Sync_proto_rawDescData)
	})
	return file_api_proto_syn_Sync_proto_rawDescData
}

var file_api_proto_syn_Sync_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_api_proto_syn_Sync_proto_goTypes = []interface{}{
	(*SyncReq)(nil), // 0: cloud.lianmi.im.syn.SyncReq
	(*SyncRsp)(nil), // 1: cloud.lianmi.im.syn.SyncRsp
}
var file_api_proto_syn_Sync_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_api_proto_syn_Sync_proto_init() }
func file_api_proto_syn_Sync_proto_init() {
	if File_api_proto_syn_Sync_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_proto_syn_Sync_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SyncReq); i {
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
		file_api_proto_syn_Sync_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SyncRsp); i {
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
			RawDescriptor: file_api_proto_syn_Sync_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_proto_syn_Sync_proto_goTypes,
		DependencyIndexes: file_api_proto_syn_Sync_proto_depIdxs,
		MessageInfos:      file_api_proto_syn_Sync_proto_msgTypes,
	}.Build()
	File_api_proto_syn_Sync_proto = out.File
	file_api_proto_syn_Sync_proto_rawDesc = nil
	file_api_proto_syn_Sync_proto_goTypes = nil
	file_api_proto_syn_Sync_proto_depIdxs = nil
}
