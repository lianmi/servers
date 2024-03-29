// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.7.0
// source: api/proto/team/UpdateTeam.proto

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
//更新群组信息请求
//群信息更新
type UpdateTeamReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// 群ID
	//是否必须：是
	TeamId string `protobuf:"bytes,1,opt,name=teamId,proto3" json:"teamId,omitempty"`
	//采用字典表方式提交更新内容 key定义成枚举(TeamFieldEnum)
	//TeamProtocal.proto 中TeamField的索引值为key
	//value为字符串，值定义为枚举的则为对应枚举值的索引的字符串表示
	//value值如果为枚举则枚举定义在TeamProtocal中
	Fields map[int32]string `protobuf:"bytes,2,rep,name=fields,proto3" json:"fields,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *UpdateTeamReq) Reset() {
	*x = UpdateTeamReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_team_UpdateTeam_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateTeamReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateTeamReq) ProtoMessage() {}

func (x *UpdateTeamReq) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_team_UpdateTeam_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateTeamReq.ProtoReflect.Descriptor instead.
func (*UpdateTeamReq) Descriptor() ([]byte, []int) {
	return file_api_proto_team_UpdateTeam_proto_rawDescGZIP(), []int{0}
}

func (x *UpdateTeamReq) GetTeamId() string {
	if x != nil {
		return x.TeamId
	}
	return ""
}

func (x *UpdateTeamReq) GetFields() map[int32]string {
	if x != nil {
		return x.Fields
	}
	return nil
}

//
//更新群组响应
//群信息更新
type UpdateTeamRsp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//群组ID
	//是否必须：是
	TeamId string `protobuf:"bytes,1,opt,name=teamId,proto3" json:"teamId,omitempty"`
	//时间标记，unix时间戳
	//是否必须:是
	TimeAt uint64 `protobuf:"fixed64,2,opt,name=timeAt,proto3" json:"timeAt,omitempty"`
}

func (x *UpdateTeamRsp) Reset() {
	*x = UpdateTeamRsp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_team_UpdateTeam_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateTeamRsp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateTeamRsp) ProtoMessage() {}

func (x *UpdateTeamRsp) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_team_UpdateTeam_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateTeamRsp.ProtoReflect.Descriptor instead.
func (*UpdateTeamRsp) Descriptor() ([]byte, []int) {
	return file_api_proto_team_UpdateTeam_proto_rawDescGZIP(), []int{1}
}

func (x *UpdateTeamRsp) GetTeamId() string {
	if x != nil {
		return x.TeamId
	}
	return ""
}

func (x *UpdateTeamRsp) GetTimeAt() uint64 {
	if x != nil {
		return x.TimeAt
	}
	return 0
}

var File_api_proto_team_UpdateTeam_proto protoreflect.FileDescriptor

var file_api_proto_team_UpdateTeam_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x74, 0x65, 0x61, 0x6d,
	0x2f, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x54, 0x65, 0x61, 0x6d, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x14, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e,
	0x69, 0x6d, 0x2e, 0x74, 0x65, 0x61, 0x6d, 0x22, 0xab, 0x01, 0x0a, 0x0d, 0x55, 0x70, 0x64, 0x61,
	0x74, 0x65, 0x54, 0x65, 0x61, 0x6d, 0x52, 0x65, 0x71, 0x12, 0x16, 0x0a, 0x06, 0x74, 0x65, 0x61,
	0x6d, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x74, 0x65, 0x61, 0x6d, 0x49,
	0x64, 0x12, 0x47, 0x0a, 0x06, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x2f, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69,
	0x2e, 0x69, 0x6d, 0x2e, 0x74, 0x65, 0x61, 0x6d, 0x2e, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x54,
	0x65, 0x61, 0x6d, 0x52, 0x65, 0x71, 0x2e, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x73, 0x45, 0x6e, 0x74,
	0x72, 0x79, 0x52, 0x06, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x73, 0x1a, 0x39, 0x0a, 0x0b, 0x46, 0x69,
	0x65, 0x6c, 0x64, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x3a, 0x02, 0x38, 0x01, 0x22, 0x3f, 0x0a, 0x0d, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x54,
	0x65, 0x61, 0x6d, 0x52, 0x73, 0x70, 0x12, 0x16, 0x0a, 0x06, 0x74, 0x65, 0x61, 0x6d, 0x49, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x74, 0x65, 0x61, 0x6d, 0x49, 0x64, 0x12, 0x16,
	0x0a, 0x06, 0x74, 0x69, 0x6d, 0x65, 0x41, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x06, 0x52, 0x06,
	0x74, 0x69, 0x6d, 0x65, 0x41, 0x74, 0x42, 0x2a, 0x5a, 0x28, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2f, 0x73, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x74, 0x65,
	0x61, 0x6d, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_proto_team_UpdateTeam_proto_rawDescOnce sync.Once
	file_api_proto_team_UpdateTeam_proto_rawDescData = file_api_proto_team_UpdateTeam_proto_rawDesc
)

func file_api_proto_team_UpdateTeam_proto_rawDescGZIP() []byte {
	file_api_proto_team_UpdateTeam_proto_rawDescOnce.Do(func() {
		file_api_proto_team_UpdateTeam_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_proto_team_UpdateTeam_proto_rawDescData)
	})
	return file_api_proto_team_UpdateTeam_proto_rawDescData
}

var file_api_proto_team_UpdateTeam_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_api_proto_team_UpdateTeam_proto_goTypes = []interface{}{
	(*UpdateTeamReq)(nil), // 0: cloud.lianmi.im.team.UpdateTeamReq
	(*UpdateTeamRsp)(nil), // 1: cloud.lianmi.im.team.UpdateTeamRsp
	nil,                   // 2: cloud.lianmi.im.team.UpdateTeamReq.FieldsEntry
}
var file_api_proto_team_UpdateTeam_proto_depIdxs = []int32{
	2, // 0: cloud.lianmi.im.team.UpdateTeamReq.fields:type_name -> cloud.lianmi.im.team.UpdateTeamReq.FieldsEntry
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_api_proto_team_UpdateTeam_proto_init() }
func file_api_proto_team_UpdateTeam_proto_init() {
	if File_api_proto_team_UpdateTeam_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_proto_team_UpdateTeam_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateTeamReq); i {
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
		file_api_proto_team_UpdateTeam_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateTeamRsp); i {
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
			RawDescriptor: file_api_proto_team_UpdateTeam_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_proto_team_UpdateTeam_proto_goTypes,
		DependencyIndexes: file_api_proto_team_UpdateTeam_proto_depIdxs,
		MessageInfos:      file_api_proto_team_UpdateTeam_proto_msgTypes,
	}.Build()
	File_api_proto_team_UpdateTeam_proto = out.File
	file_api_proto_team_UpdateTeam_proto_rawDesc = nil
	file_api_proto_team_UpdateTeam_proto_goTypes = nil
	file_api_proto_team_UpdateTeam_proto_depIdxs = nil
}
