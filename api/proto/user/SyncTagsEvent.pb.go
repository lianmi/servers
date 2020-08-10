// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.7.0
// source: api/proto/user/SyncTagsEvent.proto

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

//  同步用户标签列表
type Tag struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//用户账号
	//是否必须-是
	Username string      `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
	Type     MarkTagType `protobuf:"varint,2,opt,name=type,proto3,enum=cloud.lianmi.im.user.MarkTagType" json:"type,omitempty"`
}

func (x *Tag) Reset() {
	*x = Tag{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_user_SyncTagsEvent_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Tag) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Tag) ProtoMessage() {}

func (x *Tag) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_user_SyncTagsEvent_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Tag.ProtoReflect.Descriptor instead.
func (*Tag) Descriptor() ([]byte, []int) {
	return file_api_proto_user_SyncTagsEvent_proto_rawDescGZIP(), []int{0}
}

func (x *Tag) GetUsername() string {
	if x != nil {
		return x.Username
	}
	return ""
}

func (x *Tag) GetType() MarkTagType {
	if x != nil {
		return x.Type
	}
	return MarkTagType_Mtt_Undefined
}

// 同步用户标签列表
//用户登录成功后，增量同步黑名单列表。详情请参考同步请求
type SyncTagsEventRsp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//待添加的标签列表
	//是否必须-是
	AddTags []*Tag `protobuf:"bytes,1,rep,name=addTags,proto3" json:"addTags,omitempty"`
	//待删除的标签列表
	//是否必须-是
	RemovedTags []*Tag `protobuf:"bytes,2,rep,name=removedTags,proto3" json:"removedTags,omitempty"`
	//当前同步服务器时间
	//是否必须-是
	TimeTag uint64 `protobuf:"fixed64,3,opt,name=timeTag,proto3" json:"timeTag,omitempty"`
}

func (x *SyncTagsEventRsp) Reset() {
	*x = SyncTagsEventRsp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_user_SyncTagsEvent_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SyncTagsEventRsp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SyncTagsEventRsp) ProtoMessage() {}

func (x *SyncTagsEventRsp) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_user_SyncTagsEvent_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SyncTagsEventRsp.ProtoReflect.Descriptor instead.
func (*SyncTagsEventRsp) Descriptor() ([]byte, []int) {
	return file_api_proto_user_SyncTagsEvent_proto_rawDescGZIP(), []int{1}
}

func (x *SyncTagsEventRsp) GetAddTags() []*Tag {
	if x != nil {
		return x.AddTags
	}
	return nil
}

func (x *SyncTagsEventRsp) GetRemovedTags() []*Tag {
	if x != nil {
		return x.RemovedTags
	}
	return nil
}

func (x *SyncTagsEventRsp) GetTimeTag() uint64 {
	if x != nil {
		return x.TimeTag
	}
	return 0
}

var File_api_proto_user_SyncTagsEvent_proto protoreflect.FileDescriptor

var file_api_proto_user_SyncTagsEvent_proto_rawDesc = []byte{
	0x0a, 0x22, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x75, 0x73, 0x65, 0x72,
	0x2f, 0x53, 0x79, 0x6e, 0x63, 0x54, 0x61, 0x67, 0x73, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x14, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e,
	0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x75, 0x73, 0x65, 0x72, 0x1a, 0x1c, 0x61, 0x70, 0x69, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x75, 0x73, 0x65, 0x72, 0x2f, 0x4d, 0x61, 0x72, 0x6b, 0x54,
	0x61, 0x67, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x58, 0x0a, 0x03, 0x54, 0x61, 0x67, 0x12,
	0x1a, 0x0a, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x35, 0x0a, 0x04, 0x74,
	0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x21, 0x2e, 0x63, 0x6c, 0x6f, 0x75,
	0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x75, 0x73, 0x65, 0x72,
	0x2e, 0x4d, 0x61, 0x72, 0x6b, 0x54, 0x61, 0x67, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79,
	0x70, 0x65, 0x22, 0x9e, 0x01, 0x0a, 0x10, 0x53, 0x79, 0x6e, 0x63, 0x54, 0x61, 0x67, 0x73, 0x45,
	0x76, 0x65, 0x6e, 0x74, 0x52, 0x73, 0x70, 0x12, 0x33, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x54, 0x61,
	0x67, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64,
	0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x75, 0x73, 0x65, 0x72, 0x2e,
	0x54, 0x61, 0x67, 0x52, 0x07, 0x61, 0x64, 0x64, 0x54, 0x61, 0x67, 0x73, 0x12, 0x3b, 0x0a, 0x0b,
	0x72, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x64, 0x54, 0x61, 0x67, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x19, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69,
	0x2e, 0x69, 0x6d, 0x2e, 0x75, 0x73, 0x65, 0x72, 0x2e, 0x54, 0x61, 0x67, 0x52, 0x0b, 0x72, 0x65,
	0x6d, 0x6f, 0x76, 0x65, 0x64, 0x54, 0x61, 0x67, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x74, 0x69, 0x6d,
	0x65, 0x54, 0x61, 0x67, 0x18, 0x03, 0x20, 0x01, 0x28, 0x06, 0x52, 0x07, 0x74, 0x69, 0x6d, 0x65,
	0x54, 0x61, 0x67, 0x42, 0x2a, 0x5a, 0x28, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73,
	0x2f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x75, 0x73, 0x65, 0x72, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_proto_user_SyncTagsEvent_proto_rawDescOnce sync.Once
	file_api_proto_user_SyncTagsEvent_proto_rawDescData = file_api_proto_user_SyncTagsEvent_proto_rawDesc
)

func file_api_proto_user_SyncTagsEvent_proto_rawDescGZIP() []byte {
	file_api_proto_user_SyncTagsEvent_proto_rawDescOnce.Do(func() {
		file_api_proto_user_SyncTagsEvent_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_proto_user_SyncTagsEvent_proto_rawDescData)
	})
	return file_api_proto_user_SyncTagsEvent_proto_rawDescData
}

var file_api_proto_user_SyncTagsEvent_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_api_proto_user_SyncTagsEvent_proto_goTypes = []interface{}{
	(*Tag)(nil),              // 0: cloud.lianmi.im.user.Tag
	(*SyncTagsEventRsp)(nil), // 1: cloud.lianmi.im.user.SyncTagsEventRsp
	(MarkTagType)(0),         // 2: cloud.lianmi.im.user.MarkTagType
}
var file_api_proto_user_SyncTagsEvent_proto_depIdxs = []int32{
	2, // 0: cloud.lianmi.im.user.Tag.type:type_name -> cloud.lianmi.im.user.MarkTagType
	0, // 1: cloud.lianmi.im.user.SyncTagsEventRsp.addTags:type_name -> cloud.lianmi.im.user.Tag
	0, // 2: cloud.lianmi.im.user.SyncTagsEventRsp.removedTags:type_name -> cloud.lianmi.im.user.Tag
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_api_proto_user_SyncTagsEvent_proto_init() }
func file_api_proto_user_SyncTagsEvent_proto_init() {
	if File_api_proto_user_SyncTagsEvent_proto != nil {
		return
	}
	file_api_proto_user_MarkTag_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_api_proto_user_SyncTagsEvent_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Tag); i {
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
		file_api_proto_user_SyncTagsEvent_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SyncTagsEventRsp); i {
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
			RawDescriptor: file_api_proto_user_SyncTagsEvent_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_proto_user_SyncTagsEvent_proto_goTypes,
		DependencyIndexes: file_api_proto_user_SyncTagsEvent_proto_depIdxs,
		MessageInfos:      file_api_proto_user_SyncTagsEvent_proto_msgTypes,
	}.Build()
	File_api_proto_user_SyncTagsEvent_proto = out.File
	file_api_proto_user_SyncTagsEvent_proto_rawDesc = nil
	file_api_proto_user_SyncTagsEvent_proto_goTypes = nil
	file_api_proto_user_SyncTagsEvent_proto_depIdxs = nil
}
