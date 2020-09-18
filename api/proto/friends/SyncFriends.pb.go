// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.7.0
// source: api/proto/friends/SyncFriends.proto

package friends

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

//添加方式枚举
type VerifyAddType int32

const (
	//无效
	VerifyAddType_Va_Undefined VerifyAddType = 0
	//直接添加
	VerifyAddType_Va_Direct VerifyAddType = 1
	//校验添加
	VerifyAddType_Va_VerifyRequest VerifyAddType = 2
)

// Enum value maps for VerifyAddType.
var (
	VerifyAddType_name = map[int32]string{
		0: "Va_Undefined",
		1: "Va_Direct",
		2: "Va_VerifyRequest",
	}
	VerifyAddType_value = map[string]int32{
		"Va_Undefined":     0,
		"Va_Direct":        1,
		"Va_VerifyRequest": 2,
	}
)

func (x VerifyAddType) Enum() *VerifyAddType {
	p := new(VerifyAddType)
	*p = x
	return p
}

func (x VerifyAddType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (VerifyAddType) Descriptor() protoreflect.EnumDescriptor {
	return file_api_proto_friends_SyncFriends_proto_enumTypes[0].Descriptor()
}

func (VerifyAddType) Type() protoreflect.EnumType {
	return &file_api_proto_friends_SyncFriends_proto_enumTypes[0]
}

func (x VerifyAddType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use VerifyAddType.Descriptor instead.
func (VerifyAddType) EnumDescriptor() ([]byte, []int) {
	return file_api_proto_friends_SyncFriends_proto_rawDescGZIP(), []int{0}
}

type Friend struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//用户ID
	//是否必填-是
	Username string `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
	//备注
	//是否必填-否
	Nick string `protobuf:"bytes,2,opt,name=nick,proto3" json:"nick,omitempty"`
	//好友来源,默认0
	//由好友请求接口决定来源
	//是否必填-否
	Source string `protobuf:"bytes,3,opt,name=source,proto3" json:"source,omitempty"`
	//扩展字段，josn
	//是否必填-否
	Ex string `protobuf:"bytes,4,opt,name=ex,proto3" json:"ex,omitempty"`
	//创建时间，unix时间戳
	//是否必填-是
	CreateAt uint64 `protobuf:"fixed64,5,opt,name=createAt,proto3" json:"createAt,omitempty"`
	//最后更新时间，unix时间戳
	//是否必填-是
	UpdateAt uint64 `protobuf:"fixed64,6,opt,name=updateAt,proto3" json:"updateAt,omitempty"`
}

func (x *Friend) Reset() {
	*x = Friend{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_friends_SyncFriends_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Friend) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Friend) ProtoMessage() {}

func (x *Friend) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_friends_SyncFriends_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Friend.ProtoReflect.Descriptor instead.
func (*Friend) Descriptor() ([]byte, []int) {
	return file_api_proto_friends_SyncFriends_proto_rawDescGZIP(), []int{0}
}

func (x *Friend) GetUsername() string {
	if x != nil {
		return x.Username
	}
	return ""
}

func (x *Friend) GetNick() string {
	if x != nil {
		return x.Nick
	}
	return ""
}

func (x *Friend) GetSource() string {
	if x != nil {
		return x.Source
	}
	return ""
}

func (x *Friend) GetEx() string {
	if x != nil {
		return x.Ex
	}
	return ""
}

func (x *Friend) GetCreateAt() uint64 {
	if x != nil {
		return x.CreateAt
	}
	return 0
}

func (x *Friend) GetUpdateAt() uint64 {
	if x != nil {
		return x.UpdateAt
	}
	return 0
}

type SyncFriendsEventRsp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TimeTag         uint64    `protobuf:"fixed64,1,opt,name=timeTag,proto3" json:"timeTag,omitempty"`
	Friends         []*Friend `protobuf:"bytes,2,rep,name=friends,proto3" json:"friends,omitempty"`
	RemovedAccounts []*Friend `protobuf:"bytes,3,rep,name=removedAccounts,proto3" json:"removedAccounts,omitempty"`
}

func (x *SyncFriendsEventRsp) Reset() {
	*x = SyncFriendsEventRsp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_friends_SyncFriends_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SyncFriendsEventRsp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SyncFriendsEventRsp) ProtoMessage() {}

func (x *SyncFriendsEventRsp) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_friends_SyncFriends_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SyncFriendsEventRsp.ProtoReflect.Descriptor instead.
func (*SyncFriendsEventRsp) Descriptor() ([]byte, []int) {
	return file_api_proto_friends_SyncFriends_proto_rawDescGZIP(), []int{1}
}

func (x *SyncFriendsEventRsp) GetTimeTag() uint64 {
	if x != nil {
		return x.TimeTag
	}
	return 0
}

func (x *SyncFriendsEventRsp) GetFriends() []*Friend {
	if x != nil {
		return x.Friends
	}
	return nil
}

func (x *SyncFriendsEventRsp) GetRemovedAccounts() []*Friend {
	if x != nil {
		return x.RemovedAccounts
	}
	return nil
}

var File_api_proto_friends_SyncFriends_proto protoreflect.FileDescriptor

var file_api_proto_friends_SyncFriends_proto_rawDesc = []byte{
	0x0a, 0x23, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x66, 0x72, 0x69, 0x65,
	0x6e, 0x64, 0x73, 0x2f, 0x53, 0x79, 0x6e, 0x63, 0x46, 0x72, 0x69, 0x65, 0x6e, 0x64, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x17, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61,
	0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x66, 0x72, 0x69, 0x65, 0x6e, 0x64, 0x73, 0x22, 0x98,
	0x01, 0x0a, 0x06, 0x46, 0x72, 0x69, 0x65, 0x6e, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x75, 0x73, 0x65,
	0x72, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x75, 0x73, 0x65,
	0x72, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x69, 0x63, 0x6b, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x69, 0x63, 0x6b, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x12, 0x0e, 0x0a, 0x02, 0x65, 0x78, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x65,
	0x78, 0x12, 0x1a, 0x0a, 0x08, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x41, 0x74, 0x18, 0x05, 0x20,
	0x01, 0x28, 0x06, 0x52, 0x08, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x41, 0x74, 0x12, 0x1a, 0x0a,
	0x08, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x41, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x06, 0x52,
	0x08, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x41, 0x74, 0x22, 0xb5, 0x01, 0x0a, 0x13, 0x53, 0x79,
	0x6e, 0x63, 0x46, 0x72, 0x69, 0x65, 0x6e, 0x64, 0x73, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x73,
	0x70, 0x12, 0x18, 0x0a, 0x07, 0x74, 0x69, 0x6d, 0x65, 0x54, 0x61, 0x67, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x06, 0x52, 0x07, 0x74, 0x69, 0x6d, 0x65, 0x54, 0x61, 0x67, 0x12, 0x39, 0x0a, 0x07, 0x66,
	0x72, 0x69, 0x65, 0x6e, 0x64, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1f, 0x2e, 0x63,
	0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x66,
	0x72, 0x69, 0x65, 0x6e, 0x64, 0x73, 0x2e, 0x46, 0x72, 0x69, 0x65, 0x6e, 0x64, 0x52, 0x07, 0x66,
	0x72, 0x69, 0x65, 0x6e, 0x64, 0x73, 0x12, 0x49, 0x0a, 0x0f, 0x72, 0x65, 0x6d, 0x6f, 0x76, 0x65,
	0x64, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x1f, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69,
	0x6d, 0x2e, 0x66, 0x72, 0x69, 0x65, 0x6e, 0x64, 0x73, 0x2e, 0x46, 0x72, 0x69, 0x65, 0x6e, 0x64,
	0x52, 0x0f, 0x72, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x64, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74,
	0x73, 0x2a, 0x46, 0x0a, 0x0d, 0x56, 0x65, 0x72, 0x69, 0x66, 0x79, 0x41, 0x64, 0x64, 0x54, 0x79,
	0x70, 0x65, 0x12, 0x10, 0x0a, 0x0c, 0x56, 0x61, 0x5f, 0x55, 0x6e, 0x64, 0x65, 0x66, 0x69, 0x6e,
	0x65, 0x64, 0x10, 0x00, 0x12, 0x0d, 0x0a, 0x09, 0x56, 0x61, 0x5f, 0x44, 0x69, 0x72, 0x65, 0x63,
	0x74, 0x10, 0x01, 0x12, 0x14, 0x0a, 0x10, 0x56, 0x61, 0x5f, 0x56, 0x65, 0x72, 0x69, 0x66, 0x79,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x10, 0x02, 0x42, 0x2d, 0x5a, 0x2b, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2f, 0x73,
	0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2f, 0x66, 0x72, 0x69, 0x65, 0x6e, 0x64, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_proto_friends_SyncFriends_proto_rawDescOnce sync.Once
	file_api_proto_friends_SyncFriends_proto_rawDescData = file_api_proto_friends_SyncFriends_proto_rawDesc
)

func file_api_proto_friends_SyncFriends_proto_rawDescGZIP() []byte {
	file_api_proto_friends_SyncFriends_proto_rawDescOnce.Do(func() {
		file_api_proto_friends_SyncFriends_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_proto_friends_SyncFriends_proto_rawDescData)
	})
	return file_api_proto_friends_SyncFriends_proto_rawDescData
}

var file_api_proto_friends_SyncFriends_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_api_proto_friends_SyncFriends_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_api_proto_friends_SyncFriends_proto_goTypes = []interface{}{
	(VerifyAddType)(0),          // 0: cloud.lianmi.im.friends.VerifyAddType
	(*Friend)(nil),              // 1: cloud.lianmi.im.friends.Friend
	(*SyncFriendsEventRsp)(nil), // 2: cloud.lianmi.im.friends.SyncFriendsEventRsp
}
var file_api_proto_friends_SyncFriends_proto_depIdxs = []int32{
	1, // 0: cloud.lianmi.im.friends.SyncFriendsEventRsp.friends:type_name -> cloud.lianmi.im.friends.Friend
	1, // 1: cloud.lianmi.im.friends.SyncFriendsEventRsp.removedAccounts:type_name -> cloud.lianmi.im.friends.Friend
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_api_proto_friends_SyncFriends_proto_init() }
func file_api_proto_friends_SyncFriends_proto_init() {
	if File_api_proto_friends_SyncFriends_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_proto_friends_SyncFriends_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Friend); i {
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
		file_api_proto_friends_SyncFriends_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SyncFriendsEventRsp); i {
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
			RawDescriptor: file_api_proto_friends_SyncFriends_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_proto_friends_SyncFriends_proto_goTypes,
		DependencyIndexes: file_api_proto_friends_SyncFriends_proto_depIdxs,
		EnumInfos:         file_api_proto_friends_SyncFriends_proto_enumTypes,
		MessageInfos:      file_api_proto_friends_SyncFriends_proto_msgTypes,
	}.Build()
	File_api_proto_friends_SyncFriends_proto = out.File
	file_api_proto_friends_SyncFriends_proto_rawDesc = nil
	file_api_proto_friends_SyncFriends_proto_goTypes = nil
	file_api_proto_friends_SyncFriends_proto_depIdxs = nil
}