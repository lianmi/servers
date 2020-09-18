// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.7.0
// source: api/proto/user/SyncUpdateProfileEvent.proto

package user

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

//同步其他终端修改资料事件
//当前登录用户在其它端修改自己的个人资料之后，触发该事件
type SyncUpdateProfileEventRsp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//更新时间戳
	//是否必填-是
	TimeTag uint64 `protobuf:"fixed64,1,opt,name=timeTag,proto3" json:"timeTag,omitempty"`
	//采用字典表方式提交更新内容 key定义成枚举(UserFieldEnum)
	//取值范围：
	//Nick(1) - 昵称
	//Gender(2) - 性别
	//Avatar(3) - 头像
	//Birth(4) - 生日
	//Sign(5) - 签名
	//Tel(6) - 手机
	//Email(7) - email
	//Ex(8) - 扩展信息
	//map的key为1到8的整数含义见上
	//是否必填-是
	Fields map[int32]string `protobuf:"bytes,2,rep,name=fields,proto3" json:"fields,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *SyncUpdateProfileEventRsp) Reset() {
	*x = SyncUpdateProfileEventRsp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_user_SyncUpdateProfileEvent_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SyncUpdateProfileEventRsp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SyncUpdateProfileEventRsp) ProtoMessage() {}

func (x *SyncUpdateProfileEventRsp) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_user_SyncUpdateProfileEvent_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SyncUpdateProfileEventRsp.ProtoReflect.Descriptor instead.
func (*SyncUpdateProfileEventRsp) Descriptor() ([]byte, []int) {
	return file_api_proto_user_SyncUpdateProfileEvent_proto_rawDescGZIP(), []int{0}
}

func (x *SyncUpdateProfileEventRsp) GetTimeTag() uint64 {
	if x != nil {
		return x.TimeTag
	}
	return 0
}

func (x *SyncUpdateProfileEventRsp) GetFields() map[int32]string {
	if x != nil {
		return x.Fields
	}
	return nil
}

var File_api_proto_user_SyncUpdateProfileEvent_proto protoreflect.FileDescriptor

var file_api_proto_user_SyncUpdateProfileEvent_proto_rawDesc = []byte{
	0x0a, 0x2b, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x75, 0x73, 0x65, 0x72,
	0x2f, 0x53, 0x79, 0x6e, 0x63, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x50, 0x72, 0x6f, 0x66, 0x69,
	0x6c, 0x65, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x14, 0x63,
	0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x75,
	0x73, 0x65, 0x72, 0x22, 0xc5, 0x01, 0x0a, 0x19, 0x53, 0x79, 0x6e, 0x63, 0x55, 0x70, 0x64, 0x61,
	0x74, 0x65, 0x50, 0x72, 0x6f, 0x66, 0x69, 0x6c, 0x65, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x73,
	0x70, 0x12, 0x18, 0x0a, 0x07, 0x74, 0x69, 0x6d, 0x65, 0x54, 0x61, 0x67, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x06, 0x52, 0x07, 0x74, 0x69, 0x6d, 0x65, 0x54, 0x61, 0x67, 0x12, 0x53, 0x0a, 0x06, 0x66,
	0x69, 0x65, 0x6c, 0x64, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x3b, 0x2e, 0x63, 0x6c,
	0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x75, 0x73,
	0x65, 0x72, 0x2e, 0x53, 0x79, 0x6e, 0x63, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x50, 0x72, 0x6f,
	0x66, 0x69, 0x6c, 0x65, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x73, 0x70, 0x2e, 0x46, 0x69, 0x65,
	0x6c, 0x64, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x06, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x73,
	0x1a, 0x39, 0x0a, 0x0b, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12,
	0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x03, 0x6b, 0x65,
	0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x42, 0x2a, 0x5a, 0x28, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69,
	0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2f, 0x75, 0x73, 0x65, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_proto_user_SyncUpdateProfileEvent_proto_rawDescOnce sync.Once
	file_api_proto_user_SyncUpdateProfileEvent_proto_rawDescData = file_api_proto_user_SyncUpdateProfileEvent_proto_rawDesc
)

func file_api_proto_user_SyncUpdateProfileEvent_proto_rawDescGZIP() []byte {
	file_api_proto_user_SyncUpdateProfileEvent_proto_rawDescOnce.Do(func() {
		file_api_proto_user_SyncUpdateProfileEvent_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_proto_user_SyncUpdateProfileEvent_proto_rawDescData)
	})
	return file_api_proto_user_SyncUpdateProfileEvent_proto_rawDescData
}

var file_api_proto_user_SyncUpdateProfileEvent_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_api_proto_user_SyncUpdateProfileEvent_proto_goTypes = []interface{}{
	(*SyncUpdateProfileEventRsp)(nil), // 0: cloud.lianmi.im.user.SyncUpdateProfileEventRsp
	nil,                               // 1: cloud.lianmi.im.user.SyncUpdateProfileEventRsp.FieldsEntry
}
var file_api_proto_user_SyncUpdateProfileEvent_proto_depIdxs = []int32{
	1, // 0: cloud.lianmi.im.user.SyncUpdateProfileEventRsp.fields:type_name -> cloud.lianmi.im.user.SyncUpdateProfileEventRsp.FieldsEntry
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_api_proto_user_SyncUpdateProfileEvent_proto_init() }
func file_api_proto_user_SyncUpdateProfileEvent_proto_init() {
	if File_api_proto_user_SyncUpdateProfileEvent_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_proto_user_SyncUpdateProfileEvent_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SyncUpdateProfileEventRsp); i {
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
			RawDescriptor: file_api_proto_user_SyncUpdateProfileEvent_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_proto_user_SyncUpdateProfileEvent_proto_goTypes,
		DependencyIndexes: file_api_proto_user_SyncUpdateProfileEvent_proto_depIdxs,
		MessageInfos:      file_api_proto_user_SyncUpdateProfileEvent_proto_msgTypes,
	}.Build()
	File_api_proto_user_SyncUpdateProfileEvent_proto = out.File
	file_api_proto_user_SyncUpdateProfileEvent_proto_rawDesc = nil
	file_api_proto_user_SyncUpdateProfileEvent_proto_goTypes = nil
	file_api_proto_user_SyncUpdateProfileEvent_proto_depIdxs = nil
}