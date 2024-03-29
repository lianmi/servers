// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.7.0
// source: api/proto/friends/FriendRequest.proto

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

//
//操作类型枚举
type OptType int32

const (
	//无效
	OptType_Fr_Undefined OptType = 0
	//发起好友验证
	OptType_Fr_ApplyFriend OptType = 1
	//对方同意加你为好友
	OptType_Fr_PassFriendApply OptType = 2
	//对方拒绝添加好友
	OptType_Fr_RejectFriendApply OptType = 3
)

// Enum value maps for OptType.
var (
	OptType_name = map[int32]string{
		0: "Fr_Undefined",
		1: "Fr_ApplyFriend",
		2: "Fr_PassFriendApply",
		3: "Fr_RejectFriendApply",
	}
	OptType_value = map[string]int32{
		"Fr_Undefined":         0,
		"Fr_ApplyFriend":       1,
		"Fr_PassFriendApply":   2,
		"Fr_RejectFriendApply": 3,
	}
)

func (x OptType) Enum() *OptType {
	p := new(OptType)
	*p = x
	return p
}

func (x OptType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (OptType) Descriptor() protoreflect.EnumDescriptor {
	return file_api_proto_friends_FriendRequest_proto_enumTypes[0].Descriptor()
}

func (OptType) Type() protoreflect.EnumType {
	return &file_api_proto_friends_FriendRequest_proto_enumTypes[0]
}

func (x OptType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use OptType.Descriptor instead.
func (OptType) EnumDescriptor() ([]byte, []int) {
	return file_api_proto_friends_FriendRequest_proto_rawDescGZIP(), []int{0}
}

//操作状态枚举
type OpStatusType int32

const (
	//无效
	OpStatusType_Ost_Undefined OpStatusType = 0
	//添加好友成功
	OpStatusType_Ost_ApplySucceed OpStatusType = 1
	//等待对方同意加你为好友
	OpStatusType_Ost_WaitConfirm OpStatusType = 2
	//对方设置了拒绝任何人添加好友
	OpStatusType_Ost_RejectFriendApply OpStatusType = 3
)

// Enum value maps for OpStatusType.
var (
	OpStatusType_name = map[int32]string{
		0: "Ost_Undefined",
		1: "Ost_ApplySucceed",
		2: "Ost_WaitConfirm",
		3: "Ost_RejectFriendApply",
	}
	OpStatusType_value = map[string]int32{
		"Ost_Undefined":         0,
		"Ost_ApplySucceed":      1,
		"Ost_WaitConfirm":       2,
		"Ost_RejectFriendApply": 3,
	}
)

func (x OpStatusType) Enum() *OpStatusType {
	p := new(OpStatusType)
	*p = x
	return p
}

func (x OpStatusType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (OpStatusType) Descriptor() protoreflect.EnumDescriptor {
	return file_api_proto_friends_FriendRequest_proto_enumTypes[1].Descriptor()
}

func (OpStatusType) Type() protoreflect.EnumType {
	return &file_api_proto_friends_FriendRequest_proto_enumTypes[1]
}

func (x OpStatusType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use OpStatusType.Descriptor instead.
func (OpStatusType) EnumDescriptor() ([]byte, []int) {
	return file_api_proto_friends_FriendRequest_proto_rawDescGZIP(), []int{1}
}

//处理好友请求相关操作-请求
type FriendRequestReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//对方用户ID
	//是否必填-是
	Username string `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
	//备注
	//是否必填-否
	Ps string `protobuf:"bytes,2,opt,name=ps,proto3" json:"ps,omitempty"`
	//来源
	//是否必填-是
	//添加来源，需要添加“AddSource_Type_”前缀，后面自由拼接，如：Team、SHARE、SEARCH、QrCode等
	Source string `protobuf:"bytes,3,opt,name=source,proto3" json:"source,omitempty"`
	//操作类型
	//是否必填-是
	Type OptType `protobuf:"varint,4,opt,name=type,proto3,enum=cloud.lianmi.im.friends.OptType" json:"type,omitempty"`
}

func (x *FriendRequestReq) Reset() {
	*x = FriendRequestReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_friends_FriendRequest_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FriendRequestReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FriendRequestReq) ProtoMessage() {}

func (x *FriendRequestReq) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_friends_FriendRequest_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FriendRequestReq.ProtoReflect.Descriptor instead.
func (*FriendRequestReq) Descriptor() ([]byte, []int) {
	return file_api_proto_friends_FriendRequest_proto_rawDescGZIP(), []int{0}
}

func (x *FriendRequestReq) GetUsername() string {
	if x != nil {
		return x.Username
	}
	return ""
}

func (x *FriendRequestReq) GetPs() string {
	if x != nil {
		return x.Ps
	}
	return ""
}

func (x *FriendRequestReq) GetSource() string {
	if x != nil {
		return x.Source
	}
	return ""
}

func (x *FriendRequestReq) GetType() OptType {
	if x != nil {
		return x.Type
	}
	return OptType_Fr_Undefined
}

//处理好友请求相关操作-响应
//只包含状态码code，无内容载体
type FriendRequestRsp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//对方用户ID
	//是否必填-是
	Username string `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
	//工作流ID, 此ID会同步到其它端，一旦用对方同意 或拒绝，将会携带这个工作流ID
	//是否必填-是
	WorkflowID string `protobuf:"bytes,2,opt,name=workflowID,proto3" json:"workflowID,omitempty"`
	//操作状态
	//是否必填-是
	Status OpStatusType `protobuf:"varint,3,opt,name=status,proto3,enum=cloud.lianmi.im.friends.OpStatusType" json:"status,omitempty"`
}

func (x *FriendRequestRsp) Reset() {
	*x = FriendRequestRsp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_friends_FriendRequest_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FriendRequestRsp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FriendRequestRsp) ProtoMessage() {}

func (x *FriendRequestRsp) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_friends_FriendRequest_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FriendRequestRsp.ProtoReflect.Descriptor instead.
func (*FriendRequestRsp) Descriptor() ([]byte, []int) {
	return file_api_proto_friends_FriendRequest_proto_rawDescGZIP(), []int{1}
}

func (x *FriendRequestRsp) GetUsername() string {
	if x != nil {
		return x.Username
	}
	return ""
}

func (x *FriendRequestRsp) GetWorkflowID() string {
	if x != nil {
		return x.WorkflowID
	}
	return ""
}

func (x *FriendRequestRsp) GetStatus() OpStatusType {
	if x != nil {
		return x.Status
	}
	return OpStatusType_Ost_Undefined
}

var File_api_proto_friends_FriendRequest_proto protoreflect.FileDescriptor

var file_api_proto_friends_FriendRequest_proto_rawDesc = []byte{
	0x0a, 0x25, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x66, 0x72, 0x69, 0x65,
	0x6e, 0x64, 0x73, 0x2f, 0x46, 0x72, 0x69, 0x65, 0x6e, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x17, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c,
	0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x66, 0x72, 0x69, 0x65, 0x6e, 0x64, 0x73,
	0x22, 0x8c, 0x01, 0x0a, 0x10, 0x46, 0x72, 0x69, 0x65, 0x6e, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x52, 0x65, 0x71, 0x12, 0x1a, 0x0a, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d,
	0x65, 0x12, 0x0e, 0x0a, 0x02, 0x70, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x70,
	0x73, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x06, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x34, 0x0a, 0x04, 0x74, 0x79, 0x70,
	0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x20, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e,
	0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x66, 0x72, 0x69, 0x65, 0x6e, 0x64,
	0x73, 0x2e, 0x4f, 0x70, 0x74, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x22,
	0x8d, 0x01, 0x0a, 0x10, 0x46, 0x72, 0x69, 0x65, 0x6e, 0x64, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x52, 0x73, 0x70, 0x12, 0x1a, 0x0a, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65,
	0x12, 0x1e, 0x0a, 0x0a, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x49, 0x44, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x77, 0x6f, 0x72, 0x6b, 0x66, 0x6c, 0x6f, 0x77, 0x49, 0x44,
	0x12, 0x3d, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e,
	0x32, 0x25, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e,
	0x69, 0x6d, 0x2e, 0x66, 0x72, 0x69, 0x65, 0x6e, 0x64, 0x73, 0x2e, 0x4f, 0x70, 0x53, 0x74, 0x61,
	0x74, 0x75, 0x73, 0x54, 0x79, 0x70, 0x65, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x2a,
	0x61, 0x0a, 0x07, 0x4f, 0x70, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x10, 0x0a, 0x0c, 0x46, 0x72,
	0x5f, 0x55, 0x6e, 0x64, 0x65, 0x66, 0x69, 0x6e, 0x65, 0x64, 0x10, 0x00, 0x12, 0x12, 0x0a, 0x0e,
	0x46, 0x72, 0x5f, 0x41, 0x70, 0x70, 0x6c, 0x79, 0x46, 0x72, 0x69, 0x65, 0x6e, 0x64, 0x10, 0x01,
	0x12, 0x16, 0x0a, 0x12, 0x46, 0x72, 0x5f, 0x50, 0x61, 0x73, 0x73, 0x46, 0x72, 0x69, 0x65, 0x6e,
	0x64, 0x41, 0x70, 0x70, 0x6c, 0x79, 0x10, 0x02, 0x12, 0x18, 0x0a, 0x14, 0x46, 0x72, 0x5f, 0x52,
	0x65, 0x6a, 0x65, 0x63, 0x74, 0x46, 0x72, 0x69, 0x65, 0x6e, 0x64, 0x41, 0x70, 0x70, 0x6c, 0x79,
	0x10, 0x03, 0x2a, 0x67, 0x0a, 0x0c, 0x4f, 0x70, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x54, 0x79,
	0x70, 0x65, 0x12, 0x11, 0x0a, 0x0d, 0x4f, 0x73, 0x74, 0x5f, 0x55, 0x6e, 0x64, 0x65, 0x66, 0x69,
	0x6e, 0x65, 0x64, 0x10, 0x00, 0x12, 0x14, 0x0a, 0x10, 0x4f, 0x73, 0x74, 0x5f, 0x41, 0x70, 0x70,
	0x6c, 0x79, 0x53, 0x75, 0x63, 0x63, 0x65, 0x65, 0x64, 0x10, 0x01, 0x12, 0x13, 0x0a, 0x0f, 0x4f,
	0x73, 0x74, 0x5f, 0x57, 0x61, 0x69, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x72, 0x6d, 0x10, 0x02,
	0x12, 0x19, 0x0a, 0x15, 0x4f, 0x73, 0x74, 0x5f, 0x52, 0x65, 0x6a, 0x65, 0x63, 0x74, 0x46, 0x72,
	0x69, 0x65, 0x6e, 0x64, 0x41, 0x70, 0x70, 0x6c, 0x79, 0x10, 0x03, 0x42, 0x2d, 0x5a, 0x2b, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69,
	0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2f, 0x66, 0x72, 0x69, 0x65, 0x6e, 0x64, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_api_proto_friends_FriendRequest_proto_rawDescOnce sync.Once
	file_api_proto_friends_FriendRequest_proto_rawDescData = file_api_proto_friends_FriendRequest_proto_rawDesc
)

func file_api_proto_friends_FriendRequest_proto_rawDescGZIP() []byte {
	file_api_proto_friends_FriendRequest_proto_rawDescOnce.Do(func() {
		file_api_proto_friends_FriendRequest_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_proto_friends_FriendRequest_proto_rawDescData)
	})
	return file_api_proto_friends_FriendRequest_proto_rawDescData
}

var file_api_proto_friends_FriendRequest_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_api_proto_friends_FriendRequest_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_api_proto_friends_FriendRequest_proto_goTypes = []interface{}{
	(OptType)(0),             // 0: cloud.lianmi.im.friends.OptType
	(OpStatusType)(0),        // 1: cloud.lianmi.im.friends.OpStatusType
	(*FriendRequestReq)(nil), // 2: cloud.lianmi.im.friends.FriendRequestReq
	(*FriendRequestRsp)(nil), // 3: cloud.lianmi.im.friends.FriendRequestRsp
}
var file_api_proto_friends_FriendRequest_proto_depIdxs = []int32{
	0, // 0: cloud.lianmi.im.friends.FriendRequestReq.type:type_name -> cloud.lianmi.im.friends.OptType
	1, // 1: cloud.lianmi.im.friends.FriendRequestRsp.status:type_name -> cloud.lianmi.im.friends.OpStatusType
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_api_proto_friends_FriendRequest_proto_init() }
func file_api_proto_friends_FriendRequest_proto_init() {
	if File_api_proto_friends_FriendRequest_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_proto_friends_FriendRequest_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FriendRequestReq); i {
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
		file_api_proto_friends_FriendRequest_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FriendRequestRsp); i {
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
			RawDescriptor: file_api_proto_friends_FriendRequest_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_proto_friends_FriendRequest_proto_goTypes,
		DependencyIndexes: file_api_proto_friends_FriendRequest_proto_depIdxs,
		EnumInfos:         file_api_proto_friends_FriendRequest_proto_enumTypes,
		MessageInfos:      file_api_proto_friends_FriendRequest_proto_msgTypes,
	}.Build()
	File_api_proto_friends_FriendRequest_proto = out.File
	file_api_proto_friends_FriendRequest_proto_rawDesc = nil
	file_api_proto_friends_FriendRequest_proto_goTypes = nil
	file_api_proto_friends_FriendRequest_proto_depIdxs = nil
}
