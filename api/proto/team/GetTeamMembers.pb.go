// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.7.0
// source: api/proto/team/GetTeamMembers.proto

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
//获取群成员信息-请求
//权限说明
//普通群/高级群时： 根据timetag增量返回所有群成员
//群组：timetag固定取值0，只能拉取部分成员列表，包括群主、管理员和部分成员。
type GetTeamMembersReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//群组ID
	//是否必须-是
	TeamId string `protobuf:"bytes,1,opt,name=teamId,proto3" json:"teamId,omitempty"`
	//群成员信息最大修改时间戳，对应updateTime字段，为0时获取全量群成员副本
	//是否必须-是
	TimeAt uint64 `protobuf:"fixed64,2,opt,name=timeAt,proto3" json:"timeAt,omitempty"`
}

func (x *GetTeamMembersReq) Reset() {
	*x = GetTeamMembersReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_team_GetTeamMembers_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetTeamMembersReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetTeamMembersReq) ProtoMessage() {}

func (x *GetTeamMembersReq) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_team_GetTeamMembers_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetTeamMembersReq.ProtoReflect.Descriptor instead.
func (*GetTeamMembersReq) Descriptor() ([]byte, []int) {
	return file_api_proto_team_GetTeamMembers_proto_rawDescGZIP(), []int{0}
}

func (x *GetTeamMembersReq) GetTeamId() string {
	if x != nil {
		return x.TeamId
	}
	return ""
}

func (x *GetTeamMembersReq) GetTimeAt() uint64 {
	if x != nil {
		return x.TimeAt
	}
	return 0
}

//
//获取群成员信息-响应
//权限说明
//普通群/高级群时： 根据timetag增量返回所有群成员
//群组：timetag固定取值0，只能拉取部分成员列表，包括群主、管理员和部分成员。
type GetTeamMembersRsp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//群成员列表
	//是否必须-是
	Tmembers []*Tmember `protobuf:"bytes,1,rep,name=tmembers,proto3" json:"tmembers,omitempty"`
	//该群退出或者被踢出群群成员id，该字段普通群、普通群有效，群组该字段不传输
	//是否必须-否
	RemovedUsers []string `protobuf:"bytes,2,rep,name=removedUsers,proto3" json:"removedUsers,omitempty"`
	//本次同步后，服务器时间
	//是否必须-是
	TimeAt uint64 `protobuf:"fixed64,3,opt,name=timeAt,proto3" json:"timeAt,omitempty"`
}

func (x *GetTeamMembersRsp) Reset() {
	*x = GetTeamMembersRsp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_team_GetTeamMembers_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetTeamMembersRsp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetTeamMembersRsp) ProtoMessage() {}

func (x *GetTeamMembersRsp) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_team_GetTeamMembers_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetTeamMembersRsp.ProtoReflect.Descriptor instead.
func (*GetTeamMembersRsp) Descriptor() ([]byte, []int) {
	return file_api_proto_team_GetTeamMembers_proto_rawDescGZIP(), []int{1}
}

func (x *GetTeamMembersRsp) GetTmembers() []*Tmember {
	if x != nil {
		return x.Tmembers
	}
	return nil
}

func (x *GetTeamMembersRsp) GetRemovedUsers() []string {
	if x != nil {
		return x.RemovedUsers
	}
	return nil
}

func (x *GetTeamMembersRsp) GetTimeAt() uint64 {
	if x != nil {
		return x.TimeAt
	}
	return 0
}

var File_api_proto_team_GetTeamMembers_proto protoreflect.FileDescriptor

var file_api_proto_team_GetTeamMembers_proto_rawDesc = []byte{
	0x0a, 0x23, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x74, 0x65, 0x61, 0x6d,
	0x2f, 0x47, 0x65, 0x74, 0x54, 0x65, 0x61, 0x6d, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x14, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61,
	0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x74, 0x65, 0x61, 0x6d, 0x1a, 0x1f, 0x61, 0x70, 0x69,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x74, 0x65, 0x61, 0x6d, 0x2f, 0x43, 0x72, 0x65, 0x61,
	0x74, 0x65, 0x54, 0x65, 0x61, 0x6d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x43, 0x0a, 0x11,
	0x47, 0x65, 0x74, 0x54, 0x65, 0x61, 0x6d, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x52, 0x65,
	0x71, 0x12, 0x16, 0x0a, 0x06, 0x74, 0x65, 0x61, 0x6d, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x06, 0x74, 0x65, 0x61, 0x6d, 0x49, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x74, 0x69, 0x6d,
	0x65, 0x41, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x06, 0x52, 0x06, 0x74, 0x69, 0x6d, 0x65, 0x41,
	0x74, 0x22, 0x8a, 0x01, 0x0a, 0x11, 0x47, 0x65, 0x74, 0x54, 0x65, 0x61, 0x6d, 0x4d, 0x65, 0x6d,
	0x62, 0x65, 0x72, 0x73, 0x52, 0x73, 0x70, 0x12, 0x39, 0x0a, 0x08, 0x74, 0x6d, 0x65, 0x6d, 0x62,
	0x65, 0x72, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1d, 0x2e, 0x63, 0x6c, 0x6f, 0x75,
	0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x74, 0x65, 0x61, 0x6d,
	0x2e, 0x54, 0x6d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x52, 0x08, 0x74, 0x6d, 0x65, 0x6d, 0x62, 0x65,
	0x72, 0x73, 0x12, 0x22, 0x0a, 0x0c, 0x72, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x64, 0x55, 0x73, 0x65,
	0x72, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0c, 0x72, 0x65, 0x6d, 0x6f, 0x76, 0x65,
	0x64, 0x55, 0x73, 0x65, 0x72, 0x73, 0x12, 0x16, 0x0a, 0x06, 0x74, 0x69, 0x6d, 0x65, 0x41, 0x74,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x06, 0x52, 0x06, 0x74, 0x69, 0x6d, 0x65, 0x41, 0x74, 0x42, 0x2a,
	0x5a, 0x28, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x69, 0x61,
	0x6e, 0x6d, 0x69, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x74, 0x65, 0x61, 0x6d, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_api_proto_team_GetTeamMembers_proto_rawDescOnce sync.Once
	file_api_proto_team_GetTeamMembers_proto_rawDescData = file_api_proto_team_GetTeamMembers_proto_rawDesc
)

func file_api_proto_team_GetTeamMembers_proto_rawDescGZIP() []byte {
	file_api_proto_team_GetTeamMembers_proto_rawDescOnce.Do(func() {
		file_api_proto_team_GetTeamMembers_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_proto_team_GetTeamMembers_proto_rawDescData)
	})
	return file_api_proto_team_GetTeamMembers_proto_rawDescData
}

var file_api_proto_team_GetTeamMembers_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_api_proto_team_GetTeamMembers_proto_goTypes = []interface{}{
	(*GetTeamMembersReq)(nil), // 0: cloud.lianmi.im.team.GetTeamMembersReq
	(*GetTeamMembersRsp)(nil), // 1: cloud.lianmi.im.team.GetTeamMembersRsp
	(*Tmember)(nil),           // 2: cloud.lianmi.im.team.Tmember
}
var file_api_proto_team_GetTeamMembers_proto_depIdxs = []int32{
	2, // 0: cloud.lianmi.im.team.GetTeamMembersRsp.tmembers:type_name -> cloud.lianmi.im.team.Tmember
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_api_proto_team_GetTeamMembers_proto_init() }
func file_api_proto_team_GetTeamMembers_proto_init() {
	if File_api_proto_team_GetTeamMembers_proto != nil {
		return
	}
	file_api_proto_team_CreateTeam_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_api_proto_team_GetTeamMembers_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetTeamMembersReq); i {
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
		file_api_proto_team_GetTeamMembers_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetTeamMembersRsp); i {
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
			RawDescriptor: file_api_proto_team_GetTeamMembers_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_proto_team_GetTeamMembers_proto_goTypes,
		DependencyIndexes: file_api_proto_team_GetTeamMembers_proto_depIdxs,
		MessageInfos:      file_api_proto_team_GetTeamMembers_proto_msgTypes,
	}.Build()
	File_api_proto_team_GetTeamMembers_proto = out.File
	file_api_proto_team_GetTeamMembers_proto_rawDesc = nil
	file_api_proto_team_GetTeamMembers_proto_goTypes = nil
	file_api_proto_team_GetTeamMembers_proto_depIdxs = nil
}
