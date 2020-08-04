// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.7.0
// source: api/proto/team/SyncTeams.proto

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

//2.5.17
//增量同步群组信息事件
type SyncTeamsEvent struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//本次同步时间，unix时间戳
	//是否必须-是
	TimeTag uint64 `protobuf:"fixed64,1,opt,name=timeTag,proto3" json:"timeTag,omitempty"`
	//个人加入的群组列表
	//是否必须-是
	Teams []*Team `protobuf:"bytes,2,rep,name=teams,proto3" json:"teams,omitempty"`
	//退出\被踢出\解散群组列表
	//是否必须-否
	RemovedTeams []string `protobuf:"bytes,3,rep,name=removedTeams,proto3" json:"removedTeams,omitempty"`
}

func (x *SyncTeamsEvent) Reset() {
	*x = SyncTeamsEvent{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_team_SyncTeams_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SyncTeamsEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SyncTeamsEvent) ProtoMessage() {}

func (x *SyncTeamsEvent) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_team_SyncTeams_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SyncTeamsEvent.ProtoReflect.Descriptor instead.
func (*SyncTeamsEvent) Descriptor() ([]byte, []int) {
	return file_api_proto_team_SyncTeams_proto_rawDescGZIP(), []int{0}
}

func (x *SyncTeamsEvent) GetTimeTag() uint64 {
	if x != nil {
		return x.TimeTag
	}
	return 0
}

func (x *SyncTeamsEvent) GetTeams() []*Team {
	if x != nil {
		return x.Teams
	}
	return nil
}

func (x *SyncTeamsEvent) GetRemovedTeams() []string {
	if x != nil {
		return x.RemovedTeams
	}
	return nil
}

var File_api_proto_team_SyncTeams_proto protoreflect.FileDescriptor

var file_api_proto_team_SyncTeams_proto_rawDesc = []byte{
	0x0a, 0x1e, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x74, 0x65, 0x61, 0x6d,
	0x2f, 0x53, 0x79, 0x6e, 0x63, 0x54, 0x65, 0x61, 0x6d, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x11, 0x63, 0x63, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x74,
	0x65, 0x61, 0x6d, 0x1a, 0x19, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x74,
	0x65, 0x61, 0x6d, 0x2f, 0x54, 0x65, 0x61, 0x6d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x7d,
	0x0a, 0x0e, 0x53, 0x79, 0x6e, 0x63, 0x54, 0x65, 0x61, 0x6d, 0x73, 0x45, 0x76, 0x65, 0x6e, 0x74,
	0x12, 0x18, 0x0a, 0x07, 0x74, 0x69, 0x6d, 0x65, 0x54, 0x61, 0x67, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x06, 0x52, 0x07, 0x74, 0x69, 0x6d, 0x65, 0x54, 0x61, 0x67, 0x12, 0x2d, 0x0a, 0x05, 0x74, 0x65,
	0x61, 0x6d, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x63, 0x63, 0x2e, 0x6c,
	0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x74, 0x65, 0x61, 0x6d, 0x2e, 0x54, 0x65,
	0x61, 0x6d, 0x52, 0x05, 0x74, 0x65, 0x61, 0x6d, 0x73, 0x12, 0x22, 0x0a, 0x0c, 0x72, 0x65, 0x6d,
	0x6f, 0x76, 0x65, 0x64, 0x54, 0x65, 0x61, 0x6d, 0x73, 0x18, 0x03, 0x20, 0x03, 0x28, 0x09, 0x52,
	0x0c, 0x72, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x64, 0x54, 0x65, 0x61, 0x6d, 0x73, 0x42, 0x2a, 0x5a,
	0x28, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x69, 0x61, 0x6e,
	0x6d, 0x69, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x74, 0x65, 0x61, 0x6d, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x33,
}

var (
	file_api_proto_team_SyncTeams_proto_rawDescOnce sync.Once
	file_api_proto_team_SyncTeams_proto_rawDescData = file_api_proto_team_SyncTeams_proto_rawDesc
)

func file_api_proto_team_SyncTeams_proto_rawDescGZIP() []byte {
	file_api_proto_team_SyncTeams_proto_rawDescOnce.Do(func() {
		file_api_proto_team_SyncTeams_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_proto_team_SyncTeams_proto_rawDescData)
	})
	return file_api_proto_team_SyncTeams_proto_rawDescData
}

var file_api_proto_team_SyncTeams_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_api_proto_team_SyncTeams_proto_goTypes = []interface{}{
	(*SyncTeamsEvent)(nil), // 0: cc.lianmi.im.team.SyncTeamsEvent
	(*Team)(nil),           // 1: cc.lianmi.im.team.Team
}
var file_api_proto_team_SyncTeams_proto_depIdxs = []int32{
	1, // 0: cc.lianmi.im.team.SyncTeamsEvent.teams:type_name -> cc.lianmi.im.team.Team
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_api_proto_team_SyncTeams_proto_init() }
func file_api_proto_team_SyncTeams_proto_init() {
	if File_api_proto_team_SyncTeams_proto != nil {
		return
	}
	file_api_proto_team_Team_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_api_proto_team_SyncTeams_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SyncTeamsEvent); i {
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
			RawDescriptor: file_api_proto_team_SyncTeams_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_proto_team_SyncTeams_proto_goTypes,
		DependencyIndexes: file_api_proto_team_SyncTeams_proto_depIdxs,
		MessageInfos:      file_api_proto_team_SyncTeams_proto_msgTypes,
	}.Build()
	File_api_proto_team_SyncTeams_proto = out.File
	file_api_proto_team_SyncTeams_proto_rawDesc = nil
	file_api_proto_team_SyncTeams_proto_goTypes = nil
	file_api_proto_team_SyncTeams_proto_depIdxs = nil
}
