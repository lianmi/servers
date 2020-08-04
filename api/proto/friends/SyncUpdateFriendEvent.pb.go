// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.7.0
// source: api/proto/friends/SyncUpdateFriendEvent.proto

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

//2.4.8. 更新好友信息多终端同步事件
//同步其他终端修改好友资料事件，当某一个用户同时有两个终端(a\b)在线，
//a终端执行修改好友信息资料时，b终端会收到该事件。
type SyncUpdateFriendEvent struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//好友账号
	//是否必填-是
	Account string `protobuf:"bytes,1,opt,name=account,proto3" json:"account,omitempty"`
	//采用字典表方式提交更新内容 key定义成枚举(FriendFieldEnum)取值范围：
	//Alias(1) - 好友昵称或备注名
	//Ex(2) - 扩展字段
	Fields map[int32]string `protobuf:"bytes,2,rep,name=fields,proto3" json:"fields,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	//最后更新时间，unix时间戳
	//是否必填-是
	TimeTag uint64 `protobuf:"fixed64,3,opt,name=timeTag,proto3" json:"timeTag,omitempty"`
}

func (x *SyncUpdateFriendEvent) Reset() {
	*x = SyncUpdateFriendEvent{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_friends_SyncUpdateFriendEvent_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SyncUpdateFriendEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SyncUpdateFriendEvent) ProtoMessage() {}

func (x *SyncUpdateFriendEvent) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_friends_SyncUpdateFriendEvent_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SyncUpdateFriendEvent.ProtoReflect.Descriptor instead.
func (*SyncUpdateFriendEvent) Descriptor() ([]byte, []int) {
	return file_api_proto_friends_SyncUpdateFriendEvent_proto_rawDescGZIP(), []int{0}
}

func (x *SyncUpdateFriendEvent) GetAccount() string {
	if x != nil {
		return x.Account
	}
	return ""
}

func (x *SyncUpdateFriendEvent) GetFields() map[int32]string {
	if x != nil {
		return x.Fields
	}
	return nil
}

func (x *SyncUpdateFriendEvent) GetTimeTag() uint64 {
	if x != nil {
		return x.TimeTag
	}
	return 0
}

var File_api_proto_friends_SyncUpdateFriendEvent_proto protoreflect.FileDescriptor

var file_api_proto_friends_SyncUpdateFriendEvent_proto_rawDesc = []byte{
	0x0a, 0x2d, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x66, 0x72, 0x69, 0x65,
	0x6e, 0x64, 0x73, 0x2f, 0x53, 0x79, 0x6e, 0x63, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x46, 0x72,
	0x69, 0x65, 0x6e, 0x64, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x14, 0x63, 0x63, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x66, 0x72,
	0x69, 0x65, 0x6e, 0x64, 0x73, 0x22, 0xd7, 0x01, 0x0a, 0x15, 0x53, 0x79, 0x6e, 0x63, 0x55, 0x70,
	0x64, 0x61, 0x74, 0x65, 0x46, 0x72, 0x69, 0x65, 0x6e, 0x64, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12,
	0x18, 0x0a, 0x07, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x07, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x4f, 0x0a, 0x06, 0x66, 0x69, 0x65,
	0x6c, 0x64, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x37, 0x2e, 0x63, 0x63, 0x2e, 0x6c,
	0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x66, 0x72, 0x69, 0x65, 0x6e, 0x64, 0x73,
	0x2e, 0x53, 0x79, 0x6e, 0x63, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x46, 0x72, 0x69, 0x65, 0x6e,
	0x64, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x2e, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x73, 0x45, 0x6e, 0x74,
	0x72, 0x79, 0x52, 0x06, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x74, 0x69,
	0x6d, 0x65, 0x54, 0x61, 0x67, 0x18, 0x03, 0x20, 0x01, 0x28, 0x06, 0x52, 0x07, 0x74, 0x69, 0x6d,
	0x65, 0x54, 0x61, 0x67, 0x1a, 0x39, 0x0a, 0x0b, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x73, 0x45, 0x6e,
	0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x42,
	0x2d, 0x5a, 0x2b, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x69,
	0x61, 0x6e, 0x6d, 0x69, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x2f, 0x61, 0x70, 0x69,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x66, 0x72, 0x69, 0x65, 0x6e, 0x64, 0x73, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_proto_friends_SyncUpdateFriendEvent_proto_rawDescOnce sync.Once
	file_api_proto_friends_SyncUpdateFriendEvent_proto_rawDescData = file_api_proto_friends_SyncUpdateFriendEvent_proto_rawDesc
)

func file_api_proto_friends_SyncUpdateFriendEvent_proto_rawDescGZIP() []byte {
	file_api_proto_friends_SyncUpdateFriendEvent_proto_rawDescOnce.Do(func() {
		file_api_proto_friends_SyncUpdateFriendEvent_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_proto_friends_SyncUpdateFriendEvent_proto_rawDescData)
	})
	return file_api_proto_friends_SyncUpdateFriendEvent_proto_rawDescData
}

var file_api_proto_friends_SyncUpdateFriendEvent_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_api_proto_friends_SyncUpdateFriendEvent_proto_goTypes = []interface{}{
	(*SyncUpdateFriendEvent)(nil), // 0: cc.lianmi.im.friends.SyncUpdateFriendEvent
	nil,                           // 1: cc.lianmi.im.friends.SyncUpdateFriendEvent.FieldsEntry
}
var file_api_proto_friends_SyncUpdateFriendEvent_proto_depIdxs = []int32{
	1, // 0: cc.lianmi.im.friends.SyncUpdateFriendEvent.fields:type_name -> cc.lianmi.im.friends.SyncUpdateFriendEvent.FieldsEntry
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_api_proto_friends_SyncUpdateFriendEvent_proto_init() }
func file_api_proto_friends_SyncUpdateFriendEvent_proto_init() {
	if File_api_proto_friends_SyncUpdateFriendEvent_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_proto_friends_SyncUpdateFriendEvent_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SyncUpdateFriendEvent); i {
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
			RawDescriptor: file_api_proto_friends_SyncUpdateFriendEvent_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_proto_friends_SyncUpdateFriendEvent_proto_goTypes,
		DependencyIndexes: file_api_proto_friends_SyncUpdateFriendEvent_proto_depIdxs,
		MessageInfos:      file_api_proto_friends_SyncUpdateFriendEvent_proto_msgTypes,
	}.Build()
	File_api_proto_friends_SyncUpdateFriendEvent_proto = out.File
	file_api_proto_friends_SyncUpdateFriendEvent_proto_rawDesc = nil
	file_api_proto_friends_SyncUpdateFriendEvent_proto_goTypes = nil
	file_api_proto_friends_SyncUpdateFriendEvent_proto_depIdxs = nil
}
