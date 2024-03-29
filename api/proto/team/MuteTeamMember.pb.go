// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.7.0
// source: api/proto/team/MuteTeamMember.proto

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
//设置群成员禁言-请求
//群主/管理修改某个群成员发言模式
type MuteTeamMemberReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//群组ID
	//是否必填-是
	TeamId string `protobuf:"bytes,1,opt,name=teamId,proto3" json:"teamId,omitempty"`
	//群成员ID
	//是否必填-是
	Username string `protobuf:"bytes,2,opt,name=username,proto3" json:"username,omitempty"`
	//是否禁言,false/true
	//是否必填-是
	Mute bool `protobuf:"varint,3,opt,name=mute,proto3" json:"mute,omitempty"`
	//禁言天数，如：禁言3天
	Mutedays int32 `protobuf:"varint,4,opt,name=mutedays,proto3" json:"mutedays,omitempty"`
}

func (x *MuteTeamMemberReq) Reset() {
	*x = MuteTeamMemberReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_team_MuteTeamMember_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MuteTeamMemberReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MuteTeamMemberReq) ProtoMessage() {}

func (x *MuteTeamMemberReq) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_team_MuteTeamMember_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MuteTeamMemberReq.ProtoReflect.Descriptor instead.
func (*MuteTeamMemberReq) Descriptor() ([]byte, []int) {
	return file_api_proto_team_MuteTeamMember_proto_rawDescGZIP(), []int{0}
}

func (x *MuteTeamMemberReq) GetTeamId() string {
	if x != nil {
		return x.TeamId
	}
	return ""
}

func (x *MuteTeamMemberReq) GetUsername() string {
	if x != nil {
		return x.Username
	}
	return ""
}

func (x *MuteTeamMemberReq) GetMute() bool {
	if x != nil {
		return x.Mute
	}
	return false
}

func (x *MuteTeamMemberReq) GetMutedays() int32 {
	if x != nil {
		return x.Mutedays
	}
	return 0
}

//
//设置群成员禁言-响应
//群主/管理修改某个群成员发言模式
//只包含状态码，无内容载体
type MuteTeamMemberRsp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *MuteTeamMemberRsp) Reset() {
	*x = MuteTeamMemberRsp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_team_MuteTeamMember_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MuteTeamMemberRsp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MuteTeamMemberRsp) ProtoMessage() {}

func (x *MuteTeamMemberRsp) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_team_MuteTeamMember_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MuteTeamMemberRsp.ProtoReflect.Descriptor instead.
func (*MuteTeamMemberRsp) Descriptor() ([]byte, []int) {
	return file_api_proto_team_MuteTeamMember_proto_rawDescGZIP(), []int{1}
}

var File_api_proto_team_MuteTeamMember_proto protoreflect.FileDescriptor

var file_api_proto_team_MuteTeamMember_proto_rawDesc = []byte{
	0x0a, 0x23, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x74, 0x65, 0x61, 0x6d,
	0x2f, 0x4d, 0x75, 0x74, 0x65, 0x54, 0x65, 0x61, 0x6d, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x14, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61,
	0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x74, 0x65, 0x61, 0x6d, 0x22, 0x77, 0x0a, 0x11, 0x4d,
	0x75, 0x74, 0x65, 0x54, 0x65, 0x61, 0x6d, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x52, 0x65, 0x71,
	0x12, 0x16, 0x0a, 0x06, 0x74, 0x65, 0x61, 0x6d, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x06, 0x74, 0x65, 0x61, 0x6d, 0x49, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x75, 0x73, 0x65, 0x72,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x75, 0x73, 0x65, 0x72,
	0x6e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6d, 0x75, 0x74, 0x65, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x04, 0x6d, 0x75, 0x74, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x6d, 0x75, 0x74, 0x65,
	0x64, 0x61, 0x79, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x6d, 0x75, 0x74, 0x65,
	0x64, 0x61, 0x79, 0x73, 0x22, 0x13, 0x0a, 0x11, 0x4d, 0x75, 0x74, 0x65, 0x54, 0x65, 0x61, 0x6d,
	0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x52, 0x73, 0x70, 0x42, 0x2a, 0x5a, 0x28, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2f, 0x73,
	0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2f, 0x74, 0x65, 0x61, 0x6d, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_proto_team_MuteTeamMember_proto_rawDescOnce sync.Once
	file_api_proto_team_MuteTeamMember_proto_rawDescData = file_api_proto_team_MuteTeamMember_proto_rawDesc
)

func file_api_proto_team_MuteTeamMember_proto_rawDescGZIP() []byte {
	file_api_proto_team_MuteTeamMember_proto_rawDescOnce.Do(func() {
		file_api_proto_team_MuteTeamMember_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_proto_team_MuteTeamMember_proto_rawDescData)
	})
	return file_api_proto_team_MuteTeamMember_proto_rawDescData
}

var file_api_proto_team_MuteTeamMember_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_api_proto_team_MuteTeamMember_proto_goTypes = []interface{}{
	(*MuteTeamMemberReq)(nil), // 0: cloud.lianmi.im.team.MuteTeamMemberReq
	(*MuteTeamMemberRsp)(nil), // 1: cloud.lianmi.im.team.MuteTeamMemberRsp
}
var file_api_proto_team_MuteTeamMember_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_api_proto_team_MuteTeamMember_proto_init() }
func file_api_proto_team_MuteTeamMember_proto_init() {
	if File_api_proto_team_MuteTeamMember_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_proto_team_MuteTeamMember_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MuteTeamMemberReq); i {
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
		file_api_proto_team_MuteTeamMember_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MuteTeamMemberRsp); i {
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
			RawDescriptor: file_api_proto_team_MuteTeamMember_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_proto_team_MuteTeamMember_proto_goTypes,
		DependencyIndexes: file_api_proto_team_MuteTeamMember_proto_depIdxs,
		MessageInfos:      file_api_proto_team_MuteTeamMember_proto_msgTypes,
	}.Build()
	File_api_proto_team_MuteTeamMember_proto = out.File
	file_api_proto_team_MuteTeamMember_proto_rawDesc = nil
	file_api_proto_team_MuteTeamMember_proto_goTypes = nil
	file_api_proto_team_MuteTeamMember_proto_depIdxs = nil
}
