// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.7.0
// source: api/proto/team/PullTeamMembers.proto

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

//2.5.29. 获取指定群组成员
//根据群组用户ID获取最新群成员信息
type PullTeamMembersReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//群组ID
	//是否必填-是
	TeamId string `protobuf:"bytes,1,opt,name=teamId,proto3" json:"teamId,omitempty"`
	//群成员账号ID数组
	//是否必填-是
	Accounts []string `protobuf:"bytes,2,rep,name=accounts,proto3" json:"accounts,omitempty"`
}

func (x *PullTeamMembersReq) Reset() {
	*x = PullTeamMembersReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_team_PullTeamMembers_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PullTeamMembersReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PullTeamMembersReq) ProtoMessage() {}

func (x *PullTeamMembersReq) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_team_PullTeamMembers_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PullTeamMembersReq.ProtoReflect.Descriptor instead.
func (*PullTeamMembersReq) Descriptor() ([]byte, []int) {
	return file_api_proto_team_PullTeamMembers_proto_rawDescGZIP(), []int{0}
}

func (x *PullTeamMembersReq) GetTeamId() string {
	if x != nil {
		return x.TeamId
	}
	return ""
}

func (x *PullTeamMembersReq) GetAccounts() []string {
	if x != nil {
		return x.Accounts
	}
	return nil
}

//2.5.29. 获取指定群组成员
//根据群组用户ID获取最新群成员信息
type PullTeamMembersRsp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//群成员列表
	//是否必填-是
	Tmembers []*TeamMemberGet `protobuf:"bytes,1,rep,name=tmembers,proto3" json:"tmembers,omitempty"`
}

func (x *PullTeamMembersRsp) Reset() {
	*x = PullTeamMembersRsp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_team_PullTeamMembers_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *PullTeamMembersRsp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*PullTeamMembersRsp) ProtoMessage() {}

func (x *PullTeamMembersRsp) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_team_PullTeamMembers_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use PullTeamMembersRsp.ProtoReflect.Descriptor instead.
func (*PullTeamMembersRsp) Descriptor() ([]byte, []int) {
	return file_api_proto_team_PullTeamMembers_proto_rawDescGZIP(), []int{1}
}

func (x *PullTeamMembersRsp) GetTmembers() []*TeamMemberGet {
	if x != nil {
		return x.Tmembers
	}
	return nil
}

var File_api_proto_team_PullTeamMembers_proto protoreflect.FileDescriptor

var file_api_proto_team_PullTeamMembers_proto_rawDesc = []byte{
	0x0a, 0x24, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x74, 0x65, 0x61, 0x6d,
	0x2f, 0x50, 0x75, 0x6c, 0x6c, 0x54, 0x65, 0x61, 0x6d, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x11, 0x63, 0x63, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d,
	0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x74, 0x65, 0x61, 0x6d, 0x1a, 0x19, 0x61, 0x70, 0x69, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x74, 0x65, 0x61, 0x6d, 0x2f, 0x54, 0x65, 0x61, 0x6d, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0x48, 0x0a, 0x12, 0x50, 0x75, 0x6c, 0x6c, 0x54, 0x65, 0x61, 0x6d,
	0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x52, 0x65, 0x71, 0x12, 0x16, 0x0a, 0x06, 0x74, 0x65,
	0x61, 0x6d, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x74, 0x65, 0x61, 0x6d,
	0x49, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x73, 0x18, 0x02,
	0x20, 0x03, 0x28, 0x09, 0x52, 0x08, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x73, 0x22, 0x52,
	0x0a, 0x12, 0x50, 0x75, 0x6c, 0x6c, 0x54, 0x65, 0x61, 0x6d, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72,
	0x73, 0x52, 0x73, 0x70, 0x12, 0x3c, 0x0a, 0x08, 0x74, 0x6d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x73,
	0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x63, 0x63, 0x2e, 0x6c, 0x69, 0x61, 0x6e,
	0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x74, 0x65, 0x61, 0x6d, 0x2e, 0x54, 0x65, 0x61, 0x6d, 0x4d,
	0x65, 0x6d, 0x62, 0x65, 0x72, 0x47, 0x65, 0x74, 0x52, 0x08, 0x74, 0x6d, 0x65, 0x6d, 0x62, 0x65,
	0x72, 0x73, 0x42, 0x2a, 0x5a, 0x28, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d,
	0x2f, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x2f,
	0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x74, 0x65, 0x61, 0x6d, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_proto_team_PullTeamMembers_proto_rawDescOnce sync.Once
	file_api_proto_team_PullTeamMembers_proto_rawDescData = file_api_proto_team_PullTeamMembers_proto_rawDesc
)

func file_api_proto_team_PullTeamMembers_proto_rawDescGZIP() []byte {
	file_api_proto_team_PullTeamMembers_proto_rawDescOnce.Do(func() {
		file_api_proto_team_PullTeamMembers_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_proto_team_PullTeamMembers_proto_rawDescData)
	})
	return file_api_proto_team_PullTeamMembers_proto_rawDescData
}

var file_api_proto_team_PullTeamMembers_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_api_proto_team_PullTeamMembers_proto_goTypes = []interface{}{
	(*PullTeamMembersReq)(nil), // 0: cc.lianmi.im.team.PullTeamMembersReq
	(*PullTeamMembersRsp)(nil), // 1: cc.lianmi.im.team.PullTeamMembersRsp
	(*TeamMemberGet)(nil),      // 2: cc.lianmi.im.team.TeamMemberGet
}
var file_api_proto_team_PullTeamMembers_proto_depIdxs = []int32{
	2, // 0: cc.lianmi.im.team.PullTeamMembersRsp.tmembers:type_name -> cc.lianmi.im.team.TeamMemberGet
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_api_proto_team_PullTeamMembers_proto_init() }
func file_api_proto_team_PullTeamMembers_proto_init() {
	if File_api_proto_team_PullTeamMembers_proto != nil {
		return
	}
	file_api_proto_team_Team_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_api_proto_team_PullTeamMembers_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PullTeamMembersReq); i {
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
		file_api_proto_team_PullTeamMembers_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*PullTeamMembersRsp); i {
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
			RawDescriptor: file_api_proto_team_PullTeamMembers_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_proto_team_PullTeamMembers_proto_goTypes,
		DependencyIndexes: file_api_proto_team_PullTeamMembers_proto_depIdxs,
		MessageInfos:      file_api_proto_team_PullTeamMembers_proto_msgTypes,
	}.Build()
	File_api_proto_team_PullTeamMembers_proto = out.File
	file_api_proto_team_PullTeamMembers_proto_rawDesc = nil
	file_api_proto_team_PullTeamMembers_proto_goTypes = nil
	file_api_proto_team_PullTeamMembers_proto_depIdxs = nil
}
