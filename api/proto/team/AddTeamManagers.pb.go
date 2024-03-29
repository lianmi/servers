// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.7.0
// source: api/proto/team/AddTeamManagers.proto

package team

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
//群主设置群管理员
type AddTeamManagersReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//群组ID
	//是否必须-是
	TeamId string `protobuf:"bytes,1,opt,name=teamId,proto3" json:"teamId,omitempty"`
	//群组成员账号ID
	//是否必须-是
	Usernames []string `protobuf:"bytes,2,rep,name=usernames,proto3" json:"usernames,omitempty"`
}

func (x *AddTeamManagersReq) Reset() {
	*x = AddTeamManagersReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_team_AddTeamManagers_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AddTeamManagersReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddTeamManagersReq) ProtoMessage() {}

func (x *AddTeamManagersReq) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_team_AddTeamManagers_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddTeamManagersReq.ProtoReflect.Descriptor instead.
func (*AddTeamManagersReq) Descriptor() ([]byte, []int) {
	return file_api_proto_team_AddTeamManagers_proto_rawDescGZIP(), []int{0}
}

func (x *AddTeamManagersReq) GetTeamId() string {
	if x != nil {
		return x.TeamId
	}
	return ""
}

func (x *AddTeamManagersReq) GetUsernames() []string {
	if x != nil {
		return x.Usernames
	}
	return nil
}

//
//群主设置群管理员
type AddTeamManagersRsp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//邀请失败的用户列表(用户不存在群中、用户被封号、用户已是管理员等)
	//是否必须-是
	AbortedUsernames []string `protobuf:"bytes,1,rep,name=abortedUsernames,proto3" json:"abortedUsernames,omitempty"`
}

func (x *AddTeamManagersRsp) Reset() {
	*x = AddTeamManagersRsp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_team_AddTeamManagers_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AddTeamManagersRsp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddTeamManagersRsp) ProtoMessage() {}

func (x *AddTeamManagersRsp) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_team_AddTeamManagers_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddTeamManagersRsp.ProtoReflect.Descriptor instead.
func (*AddTeamManagersRsp) Descriptor() ([]byte, []int) {
	return file_api_proto_team_AddTeamManagers_proto_rawDescGZIP(), []int{1}
}

func (x *AddTeamManagersRsp) GetAbortedUsernames() []string {
	if x != nil {
		return x.AbortedUsernames
	}
	return nil
}

var File_api_proto_team_AddTeamManagers_proto protoreflect.FileDescriptor

var file_api_proto_team_AddTeamManagers_proto_rawDesc = []byte{
	0x0a, 0x24, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x74, 0x65, 0x61, 0x6d,
	0x2f, 0x41, 0x64, 0x64, 0x54, 0x65, 0x61, 0x6d, 0x4d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x14, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69,
	0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x74, 0x65, 0x61, 0x6d, 0x22, 0x4a, 0x0a, 0x12,
	0x41, 0x64, 0x64, 0x54, 0x65, 0x61, 0x6d, 0x4d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x73, 0x52,
	0x65, 0x71, 0x12, 0x16, 0x0a, 0x06, 0x74, 0x65, 0x61, 0x6d, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x06, 0x74, 0x65, 0x61, 0x6d, 0x49, 0x64, 0x12, 0x1c, 0x0a, 0x09, 0x75, 0x73,
	0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x09, 0x75,
	0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x22, 0x40, 0x0a, 0x12, 0x41, 0x64, 0x64, 0x54,
	0x65, 0x61, 0x6d, 0x4d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x73, 0x52, 0x73, 0x70, 0x12, 0x2a,
	0x0a, 0x10, 0x61, 0x62, 0x6f, 0x72, 0x74, 0x65, 0x64, 0x55, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d,
	0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x09, 0x52, 0x10, 0x61, 0x62, 0x6f, 0x72, 0x74, 0x65,
	0x64, 0x55, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x42, 0x2a, 0x5a, 0x28, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2f,
	0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x2f, 0x74, 0x65, 0x61, 0x6d, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_proto_team_AddTeamManagers_proto_rawDescOnce sync.Once
	file_api_proto_team_AddTeamManagers_proto_rawDescData = file_api_proto_team_AddTeamManagers_proto_rawDesc
)

func file_api_proto_team_AddTeamManagers_proto_rawDescGZIP() []byte {
	file_api_proto_team_AddTeamManagers_proto_rawDescOnce.Do(func() {
		file_api_proto_team_AddTeamManagers_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_proto_team_AddTeamManagers_proto_rawDescData)
	})
	return file_api_proto_team_AddTeamManagers_proto_rawDescData
}

var file_api_proto_team_AddTeamManagers_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_api_proto_team_AddTeamManagers_proto_goTypes = []interface{}{
	(*AddTeamManagersReq)(nil), // 0: cloud.lianmi.im.team.AddTeamManagersReq
	(*AddTeamManagersRsp)(nil), // 1: cloud.lianmi.im.team.AddTeamManagersRsp
}
var file_api_proto_team_AddTeamManagers_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_api_proto_team_AddTeamManagers_proto_init() }
func file_api_proto_team_AddTeamManagers_proto_init() {
	if File_api_proto_team_AddTeamManagers_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_proto_team_AddTeamManagers_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AddTeamManagersReq); i {
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
		file_api_proto_team_AddTeamManagers_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AddTeamManagersRsp); i {
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
			RawDescriptor: file_api_proto_team_AddTeamManagers_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_proto_team_AddTeamManagers_proto_goTypes,
		DependencyIndexes: file_api_proto_team_AddTeamManagers_proto_depIdxs,
		MessageInfos:      file_api_proto_team_AddTeamManagers_proto_msgTypes,
	}.Build()
	File_api_proto_team_AddTeamManagers_proto = out.File
	file_api_proto_team_AddTeamManagers_proto_rawDesc = nil
	file_api_proto_team_AddTeamManagers_proto_goTypes = nil
	file_api_proto_team_AddTeamManagers_proto_depIdxs = nil
}
