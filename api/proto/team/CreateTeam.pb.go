// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.7.0
// source: api/proto/team/CreateTeam.proto

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

type MemberStatus int32

const (
	MemberStatus_MemStatus_Undefined MemberStatus = 0
	//启用
	MemberStatus_MemStatus_Enabled MemberStatus = 1
	//禁用
	MemberStatus_MemStatus_Disabled MemberStatus = 2
)

// Enum value maps for MemberStatus.
var (
	MemberStatus_name = map[int32]string{
		0: "MemStatus_Undefined",
		1: "MemStatus_Enabled",
		2: "MemStatus_Disabled",
	}
	MemberStatus_value = map[string]int32{
		"MemStatus_Undefined": 0,
		"MemStatus_Enabled":   1,
		"MemStatus_Disabled":  2,
	}
)

func (x MemberStatus) Enum() *MemberStatus {
	p := new(MemberStatus)
	*p = x
	return p
}

func (x MemberStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (MemberStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_api_proto_team_CreateTeam_proto_enumTypes[0].Descriptor()
}

func (MemberStatus) Type() protoreflect.EnumType {
	return &file_api_proto_team_CreateTeam_proto_enumTypes[0]
}

func (x MemberStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use MemberStatus.Descriptor instead.
func (MemberStatus) EnumDescriptor() ([]byte, []int) {
	return file_api_proto_team_CreateTeam_proto_rawDescGZIP(), []int{0}
}

//
//创建普通群/创建群-请求
type CreateTeamReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Creator string `protobuf:"bytes,1,opt,name=creator,proto3" json:"creator,omitempty"`
	Owner   string `protobuf:"bytes,2,opt,name=owner,proto3" json:"owner,omitempty"`
	//群组类型
	//是否必须：是
	Type TeamType `protobuf:"varint,3,opt,name=type,proto3,enum=cloud.lianmi.im.team.TeamType" json:"type,omitempty"`
	Name string   `protobuf:"bytes,4,opt,name=name,proto3" json:"name,omitempty"`
	//入群校验方式
	//是否必须：是
	VerifyType VerifyType `protobuf:"varint,5,opt,name=verifyType,proto3,enum=cloud.lianmi.im.team.VerifyType" json:"verifyType,omitempty"`
	//邀请用户入群前是否需要管理员同意
	//Check - 需管理员同意才能邀请用户入群
	InviteMode InviteMode `protobuf:"varint,6,opt,name=inviteMode,proto3,enum=cloud.lianmi.im.team.InviteMode" json:"inviteMode,omitempty"`
}

func (x *CreateTeamReq) Reset() {
	*x = CreateTeamReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_team_CreateTeam_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateTeamReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateTeamReq) ProtoMessage() {}

func (x *CreateTeamReq) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_team_CreateTeam_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateTeamReq.ProtoReflect.Descriptor instead.
func (*CreateTeamReq) Descriptor() ([]byte, []int) {
	return file_api_proto_team_CreateTeam_proto_rawDescGZIP(), []int{0}
}

func (x *CreateTeamReq) GetCreator() string {
	if x != nil {
		return x.Creator
	}
	return ""
}

func (x *CreateTeamReq) GetOwner() string {
	if x != nil {
		return x.Owner
	}
	return ""
}

func (x *CreateTeamReq) GetType() TeamType {
	if x != nil {
		return x.Type
	}
	return TeamType_Tt_Undefined
}

func (x *CreateTeamReq) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *CreateTeamReq) GetVerifyType() VerifyType {
	if x != nil {
		return x.VerifyType
	}
	return VerifyType_Vt_Undefined
}

func (x *CreateTeamReq) GetInviteMode() InviteMode {
	if x != nil {
		return x.InviteMode
	}
	return InviteMode_Invite_Undefined
}

//
//创建普通群/创建群-响应
type CreateTeamRsp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//群组信息
	//是否必须：是
	TeamInfo *TeamInfo `protobuf:"bytes,1,opt,name=teamInfo,proto3" json:"teamInfo,omitempty"`
	//邀请失败列表
	//是否必须：否
	FailedAccidList []string `protobuf:"bytes,2,rep,name=failedAccidList,proto3" json:"failedAccidList,omitempty"`
}

func (x *CreateTeamRsp) Reset() {
	*x = CreateTeamRsp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_team_CreateTeam_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateTeamRsp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateTeamRsp) ProtoMessage() {}

func (x *CreateTeamRsp) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_team_CreateTeam_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateTeamRsp.ProtoReflect.Descriptor instead.
func (*CreateTeamRsp) Descriptor() ([]byte, []int) {
	return file_api_proto_team_CreateTeam_proto_rawDescGZIP(), []int{1}
}

func (x *CreateTeamRsp) GetTeamInfo() *TeamInfo {
	if x != nil {
		return x.TeamInfo
	}
	return nil
}

func (x *CreateTeamRsp) GetFailedAccidList() []string {
	if x != nil {
		return x.FailedAccidList
	}
	return nil
}

//
//获取群成员信息-响应
//权限说明
//普通群/高级群时： 根据timetag增量返回所有群成员
//部落：timetag固定取值0，只能拉取部分成员列表，包括群主、管理员和部分成员。
type Tmember struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//群组ID
	//是否必须-是
	TeamId string `protobuf:"bytes,1,opt,name=teamId,proto3" json:"teamId,omitempty"`
	//用户账号id
	//是否必须-是
	Username string `protobuf:"bytes,2,opt,name=username,proto3" json:"username,omitempty"`
	//群昵称
	//是否必须-是
	Nick string `protobuf:"bytes,3,opt,name=nick,proto3" json:"nick,omitempty"`
	//群成员头像
	//是否必须-否
	Avatar string `protobuf:"bytes,4,opt,name=avatar,proto3" json:"avatar,omitempty"`
	//群成员来源,需要添加“AddSource_Type_”前缀，后面自由拼接，如：PC、SHARE、SEARCH、IOS等,
	//UI可以根据该字段动态显示该成员来源,例: "abc邀请cde加入群组"
	//是否必须-是
	Source string `protobuf:"bytes,5,opt,name=source,proto3" json:"source,omitempty"`
	//成员类型
	//是否必须-是
	Type TeamMemberType `protobuf:"varint,6,opt,name=type,proto3,enum=cloud.lianmi.im.team.TeamMemberType" json:"type,omitempty"`
	//群消息通知方式
	//是否必须-是
	NotifyType NotifyType `protobuf:"varint,7,opt,name=notifyType,proto3,enum=cloud.lianmi.im.team.NotifyType" json:"notifyType,omitempty"`
	//是否被禁言
	//是否必须-是
	Mute bool `protobuf:"varint,8,opt,name=mute,proto3" json:"mute,omitempty"`
	//扩展字段
	//是否必须-否
	Ex string `protobuf:"bytes,9,opt,name=ex,proto3" json:"ex,omitempty"`
	//入群时间，unix时间戳
	//是否必须-是
	JoinTime uint64 `protobuf:"fixed64,10,opt,name=joinTime,proto3" json:"joinTime,omitempty"`
	//最近更新时间，unix时间戳
	//是否必须-是
	UpdateTime uint64 `protobuf:"fixed64,11,opt,name=updateTime,proto3" json:"updateTime,omitempty"`
}

func (x *Tmember) Reset() {
	*x = Tmember{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_team_CreateTeam_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Tmember) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Tmember) ProtoMessage() {}

func (x *Tmember) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_team_CreateTeam_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Tmember.ProtoReflect.Descriptor instead.
func (*Tmember) Descriptor() ([]byte, []int) {
	return file_api_proto_team_CreateTeam_proto_rawDescGZIP(), []int{2}
}

func (x *Tmember) GetTeamId() string {
	if x != nil {
		return x.TeamId
	}
	return ""
}

func (x *Tmember) GetUsername() string {
	if x != nil {
		return x.Username
	}
	return ""
}

func (x *Tmember) GetNick() string {
	if x != nil {
		return x.Nick
	}
	return ""
}

func (x *Tmember) GetAvatar() string {
	if x != nil {
		return x.Avatar
	}
	return ""
}

func (x *Tmember) GetSource() string {
	if x != nil {
		return x.Source
	}
	return ""
}

func (x *Tmember) GetType() TeamMemberType {
	if x != nil {
		return x.Type
	}
	return TeamMemberType_Tmt_Undefined
}

func (x *Tmember) GetNotifyType() NotifyType {
	if x != nil {
		return x.NotifyType
	}
	return NotifyType_Notify_Undefined
}

func (x *Tmember) GetMute() bool {
	if x != nil {
		return x.Mute
	}
	return false
}

func (x *Tmember) GetEx() string {
	if x != nil {
		return x.Ex
	}
	return ""
}

func (x *Tmember) GetJoinTime() uint64 {
	if x != nil {
		return x.JoinTime
	}
	return 0
}

func (x *Tmember) GetUpdateTime() uint64 {
	if x != nil {
		return x.UpdateTime
	}
	return 0
}

var File_api_proto_team_CreateTeam_proto protoreflect.FileDescriptor

var file_api_proto_team_CreateTeam_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x74, 0x65, 0x61, 0x6d,
	0x2f, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x54, 0x65, 0x61, 0x6d, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x14, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e,
	0x69, 0x6d, 0x2e, 0x74, 0x65, 0x61, 0x6d, 0x1a, 0x1d, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2f, 0x74, 0x65, 0x61, 0x6d, 0x2f, 0x54, 0x65, 0x61, 0x6d, 0x49, 0x6e, 0x66, 0x6f,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x8b, 0x02, 0x0a, 0x0d, 0x43, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x54, 0x65, 0x61, 0x6d, 0x52, 0x65, 0x71, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x72, 0x65, 0x61,
	0x74, 0x6f, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x72, 0x65, 0x61, 0x74,
	0x6f, 0x72, 0x12, 0x14, 0x0a, 0x05, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x05, 0x6f, 0x77, 0x6e, 0x65, 0x72, 0x12, 0x32, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1e, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c,
	0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x74, 0x65, 0x61, 0x6d, 0x2e, 0x54, 0x65,
	0x61, 0x6d, 0x54, 0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x12, 0x0a, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x12, 0x40, 0x0a, 0x0a, 0x76, 0x65, 0x72, 0x69, 0x66, 0x79, 0x54, 0x79, 0x70, 0x65, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x0e, 0x32, 0x20, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61,
	0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x74, 0x65, 0x61, 0x6d, 0x2e, 0x56, 0x65, 0x72, 0x69,
	0x66, 0x79, 0x54, 0x79, 0x70, 0x65, 0x52, 0x0a, 0x76, 0x65, 0x72, 0x69, 0x66, 0x79, 0x54, 0x79,
	0x70, 0x65, 0x12, 0x40, 0x0a, 0x0a, 0x69, 0x6e, 0x76, 0x69, 0x74, 0x65, 0x4d, 0x6f, 0x64, 0x65,
	0x18, 0x06, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x20, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c,
	0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x74, 0x65, 0x61, 0x6d, 0x2e, 0x49, 0x6e,
	0x76, 0x69, 0x74, 0x65, 0x4d, 0x6f, 0x64, 0x65, 0x52, 0x0a, 0x69, 0x6e, 0x76, 0x69, 0x74, 0x65,
	0x4d, 0x6f, 0x64, 0x65, 0x22, 0x75, 0x0a, 0x0d, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x54, 0x65,
	0x61, 0x6d, 0x52, 0x73, 0x70, 0x12, 0x3a, 0x0a, 0x08, 0x74, 0x65, 0x61, 0x6d, 0x49, 0x6e, 0x66,
	0x6f, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e,
	0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x74, 0x65, 0x61, 0x6d, 0x2e, 0x54,
	0x65, 0x61, 0x6d, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x08, 0x74, 0x65, 0x61, 0x6d, 0x49, 0x6e, 0x66,
	0x6f, 0x12, 0x28, 0x0a, 0x0f, 0x66, 0x61, 0x69, 0x6c, 0x65, 0x64, 0x41, 0x63, 0x63, 0x69, 0x64,
	0x4c, 0x69, 0x73, 0x74, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0f, 0x66, 0x61, 0x69, 0x6c,
	0x65, 0x64, 0x41, 0x63, 0x63, 0x69, 0x64, 0x4c, 0x69, 0x73, 0x74, 0x22, 0xdd, 0x02, 0x0a, 0x07,
	0x54, 0x6d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x16, 0x0a, 0x06, 0x74, 0x65, 0x61, 0x6d, 0x49,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x74, 0x65, 0x61, 0x6d, 0x49, 0x64, 0x12,
	0x1a, 0x0a, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e,
	0x69, 0x63, 0x6b, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x69, 0x63, 0x6b, 0x12,
	0x16, 0x0a, 0x06, 0x61, 0x76, 0x61, 0x74, 0x61, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x06, 0x61, 0x76, 0x61, 0x74, 0x61, 0x72, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12,
	0x38, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x24, 0x2e,
	0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e,
	0x74, 0x65, 0x61, 0x6d, 0x2e, 0x54, 0x65, 0x61, 0x6d, 0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x54,
	0x79, 0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x40, 0x0a, 0x0a, 0x6e, 0x6f, 0x74,
	0x69, 0x66, 0x79, 0x54, 0x79, 0x70, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x20, 0x2e,
	0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e,
	0x74, 0x65, 0x61, 0x6d, 0x2e, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x54, 0x79, 0x70, 0x65, 0x52,
	0x0a, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x54, 0x79, 0x70, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6d,
	0x75, 0x74, 0x65, 0x18, 0x08, 0x20, 0x01, 0x28, 0x08, 0x52, 0x04, 0x6d, 0x75, 0x74, 0x65, 0x12,
	0x0e, 0x0a, 0x02, 0x65, 0x78, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x65, 0x78, 0x12,
	0x1a, 0x0a, 0x08, 0x6a, 0x6f, 0x69, 0x6e, 0x54, 0x69, 0x6d, 0x65, 0x18, 0x0a, 0x20, 0x01, 0x28,
	0x06, 0x52, 0x08, 0x6a, 0x6f, 0x69, 0x6e, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x75,
	0x70, 0x64, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x06, 0x52,
	0x0a, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x2a, 0x56, 0x0a, 0x0c, 0x4d,
	0x65, 0x6d, 0x62, 0x65, 0x72, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x17, 0x0a, 0x13, 0x4d,
	0x65, 0x6d, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x5f, 0x55, 0x6e, 0x64, 0x65, 0x66, 0x69, 0x6e,
	0x65, 0x64, 0x10, 0x00, 0x12, 0x15, 0x0a, 0x11, 0x4d, 0x65, 0x6d, 0x53, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x5f, 0x45, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x64, 0x10, 0x01, 0x12, 0x16, 0x0a, 0x12, 0x4d,
	0x65, 0x6d, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x5f, 0x44, 0x69, 0x73, 0x61, 0x62, 0x6c, 0x65,
	0x64, 0x10, 0x02, 0x42, 0x2a, 0x5a, 0x28, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73,
	0x2f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x74, 0x65, 0x61, 0x6d, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_proto_team_CreateTeam_proto_rawDescOnce sync.Once
	file_api_proto_team_CreateTeam_proto_rawDescData = file_api_proto_team_CreateTeam_proto_rawDesc
)

func file_api_proto_team_CreateTeam_proto_rawDescGZIP() []byte {
	file_api_proto_team_CreateTeam_proto_rawDescOnce.Do(func() {
		file_api_proto_team_CreateTeam_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_proto_team_CreateTeam_proto_rawDescData)
	})
	return file_api_proto_team_CreateTeam_proto_rawDescData
}

var file_api_proto_team_CreateTeam_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_api_proto_team_CreateTeam_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_api_proto_team_CreateTeam_proto_goTypes = []interface{}{
	(MemberStatus)(0),     // 0: cloud.lianmi.im.team.MemberStatus
	(*CreateTeamReq)(nil), // 1: cloud.lianmi.im.team.CreateTeamReq
	(*CreateTeamRsp)(nil), // 2: cloud.lianmi.im.team.CreateTeamRsp
	(*Tmember)(nil),       // 3: cloud.lianmi.im.team.Tmember
	(TeamType)(0),         // 4: cloud.lianmi.im.team.TeamType
	(VerifyType)(0),       // 5: cloud.lianmi.im.team.VerifyType
	(InviteMode)(0),       // 6: cloud.lianmi.im.team.InviteMode
	(*TeamInfo)(nil),      // 7: cloud.lianmi.im.team.TeamInfo
	(TeamMemberType)(0),   // 8: cloud.lianmi.im.team.TeamMemberType
	(NotifyType)(0),       // 9: cloud.lianmi.im.team.NotifyType
}
var file_api_proto_team_CreateTeam_proto_depIdxs = []int32{
	4, // 0: cloud.lianmi.im.team.CreateTeamReq.type:type_name -> cloud.lianmi.im.team.TeamType
	5, // 1: cloud.lianmi.im.team.CreateTeamReq.verifyType:type_name -> cloud.lianmi.im.team.VerifyType
	6, // 2: cloud.lianmi.im.team.CreateTeamReq.inviteMode:type_name -> cloud.lianmi.im.team.InviteMode
	7, // 3: cloud.lianmi.im.team.CreateTeamRsp.teamInfo:type_name -> cloud.lianmi.im.team.TeamInfo
	8, // 4: cloud.lianmi.im.team.Tmember.type:type_name -> cloud.lianmi.im.team.TeamMemberType
	9, // 5: cloud.lianmi.im.team.Tmember.notifyType:type_name -> cloud.lianmi.im.team.NotifyType
	6, // [6:6] is the sub-list for method output_type
	6, // [6:6] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_api_proto_team_CreateTeam_proto_init() }
func file_api_proto_team_CreateTeam_proto_init() {
	if File_api_proto_team_CreateTeam_proto != nil {
		return
	}
	file_api_proto_team_TeamInfo_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_api_proto_team_CreateTeam_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateTeamReq); i {
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
		file_api_proto_team_CreateTeam_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateTeamRsp); i {
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
		file_api_proto_team_CreateTeam_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Tmember); i {
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
			RawDescriptor: file_api_proto_team_CreateTeam_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_proto_team_CreateTeam_proto_goTypes,
		DependencyIndexes: file_api_proto_team_CreateTeam_proto_depIdxs,
		EnumInfos:         file_api_proto_team_CreateTeam_proto_enumTypes,
		MessageInfos:      file_api_proto_team_CreateTeam_proto_msgTypes,
	}.Build()
	File_api_proto_team_CreateTeam_proto = out.File
	file_api_proto_team_CreateTeam_proto_rawDesc = nil
	file_api_proto_team_CreateTeam_proto_goTypes = nil
	file_api_proto_team_CreateTeam_proto_depIdxs = nil
}
