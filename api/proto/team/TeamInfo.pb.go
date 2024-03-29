// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.7.0
// source: api/proto/team/TeamInfo.proto

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
//枚举信息
type TeamType int32

const (
	// 无效
	TeamType_Tt_Undefined TeamType = 0
	// 普通群,类似微信群,弱管理方式
	TeamType_Tt_Normal TeamType = 1
	// 普通群,类似qq群有更多管理权限
	TeamType_Tt_Advanced TeamType = 2
	// 群组，超大群
	TeamType_Tt_Vip TeamType = 3
	// 临时群
	TeamType_Tt_Temporary TeamType = 4
)

// Enum value maps for TeamType.
var (
	TeamType_name = map[int32]string{
		0: "Tt_Undefined",
		1: "Tt_Normal",
		2: "Tt_Advanced",
		3: "Tt_Vip",
		4: "Tt_Temporary",
	}
	TeamType_value = map[string]int32{
		"Tt_Undefined": 0,
		"Tt_Normal":    1,
		"Tt_Advanced":  2,
		"Tt_Vip":       3,
		"Tt_Temporary": 4,
	}
)

func (x TeamType) Enum() *TeamType {
	p := new(TeamType)
	*p = x
	return p
}

func (x TeamType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (TeamType) Descriptor() protoreflect.EnumDescriptor {
	return file_api_proto_team_TeamInfo_proto_enumTypes[0].Descriptor()
}

func (TeamType) Type() protoreflect.EnumType {
	return &file_api_proto_team_TeamInfo_proto_enumTypes[0]
}

func (x TeamType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use TeamType.Descriptor instead.
func (TeamType) EnumDescriptor() ([]byte, []int) {
	return file_api_proto_team_TeamInfo_proto_rawDescGZIP(), []int{0}
}

//查询类型
type QueryType int32

const (
	//无效
	QueryType_Tmqt_Undefined QueryType = 0
	//全部,默认
	QueryType_Tmqt_All QueryType = 1
	//管理员
	QueryType_Tmqt_Manager QueryType = 2
	//禁言成员
	QueryType_Tmqt_Muted QueryType = 3
)

// Enum value maps for QueryType.
var (
	QueryType_name = map[int32]string{
		0: "Tmqt_Undefined",
		1: "Tmqt_All",
		2: "Tmqt_Manager",
		3: "Tmqt_Muted",
	}
	QueryType_value = map[string]int32{
		"Tmqt_Undefined": 0,
		"Tmqt_All":       1,
		"Tmqt_Manager":   2,
		"Tmqt_Muted":     3,
	}
)

func (x QueryType) Enum() *QueryType {
	p := new(QueryType)
	*p = x
	return p
}

func (x QueryType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (QueryType) Descriptor() protoreflect.EnumDescriptor {
	return file_api_proto_team_TeamInfo_proto_enumTypes[1].Descriptor()
}

func (QueryType) Type() protoreflect.EnumType {
	return &file_api_proto_team_TeamInfo_proto_enumTypes[1]
}

func (x QueryType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use QueryType.Descriptor instead.
func (QueryType) EnumDescriptor() ([]byte, []int) {
	return file_api_proto_team_TeamInfo_proto_rawDescGZIP(), []int{1}
}

//群状态
type TeamStatus int32

const (
	//初始状态,未审核
	TeamStatus_Status_Init TeamStatus = 0
	//正常状态
	TeamStatus_Status_Normal TeamStatus = 1
	//封禁状态
	TeamStatus_Status_Blocked TeamStatus = 2
	// 解散状态
	TeamStatus_Status_DisMissed TeamStatus = 3
)

// Enum value maps for TeamStatus.
var (
	TeamStatus_name = map[int32]string{
		0: "Status_Init",
		1: "Status_Normal",
		2: "Status_Blocked",
		3: "Status_DisMissed",
	}
	TeamStatus_value = map[string]int32{
		"Status_Init":      0,
		"Status_Normal":    1,
		"Status_Blocked":   2,
		"Status_DisMissed": 3,
	}
)

func (x TeamStatus) Enum() *TeamStatus {
	p := new(TeamStatus)
	*p = x
	return p
}

func (x TeamStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (TeamStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_api_proto_team_TeamInfo_proto_enumTypes[2].Descriptor()
}

func (TeamStatus) Type() protoreflect.EnumType {
	return &file_api_proto_team_TeamInfo_proto_enumTypes[2]
}

func (x TeamStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use TeamStatus.Descriptor instead.
func (TeamStatus) EnumDescriptor() ([]byte, []int) {
	return file_api_proto_team_TeamInfo_proto_rawDescGZIP(), []int{2}
}

//校验模式
type VerifyType int32

const (
	// 无定义
	VerifyType_Vt_Undefined VerifyType = 0
	//所有人可加入
	VerifyType_Vt_Free VerifyType = 1
	//需要审核加入
	VerifyType_Vt_Apply VerifyType = 2
	//仅限邀请加入
	VerifyType_Vt_Private VerifyType = 3
)

// Enum value maps for VerifyType.
var (
	VerifyType_name = map[int32]string{
		0: "Vt_Undefined",
		1: "Vt_Free",
		2: "Vt_Apply",
		3: "Vt_Private",
	}
	VerifyType_value = map[string]int32{
		"Vt_Undefined": 0,
		"Vt_Free":      1,
		"Vt_Apply":     2,
		"Vt_Private":   3,
	}
)

func (x VerifyType) Enum() *VerifyType {
	p := new(VerifyType)
	*p = x
	return p
}

func (x VerifyType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (VerifyType) Descriptor() protoreflect.EnumDescriptor {
	return file_api_proto_team_TeamInfo_proto_enumTypes[3].Descriptor()
}

func (VerifyType) Type() protoreflect.EnumType {
	return &file_api_proto_team_TeamInfo_proto_enumTypes[3]
}

func (x VerifyType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use VerifyType.Descriptor instead.
func (VerifyType) EnumDescriptor() ([]byte, []int) {
	return file_api_proto_team_TeamInfo_proto_rawDescGZIP(), []int{3}
}

//发言方式
type MuteMode int32

const (
	// 无定义
	MuteMode_Mm_Undefined MuteMode = 0
	//所有人可发言
	MuteMode_Mm_None MuteMode = 1
	//群主可发言,集体禁言
	MuteMode_Mm_MuteALL MuteMode = 2
	//管理员可发言,普通成员禁言
	MuteMode_Mm_MuteNormal MuteMode = 3
)

// Enum value maps for MuteMode.
var (
	MuteMode_name = map[int32]string{
		0: "Mm_Undefined",
		1: "Mm_None",
		2: "Mm_MuteALL",
		3: "Mm_MuteNormal",
	}
	MuteMode_value = map[string]int32{
		"Mm_Undefined":  0,
		"Mm_None":       1,
		"Mm_MuteALL":    2,
		"Mm_MuteNormal": 3,
	}
)

func (x MuteMode) Enum() *MuteMode {
	p := new(MuteMode)
	*p = x
	return p
}

func (x MuteMode) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (MuteMode) Descriptor() protoreflect.EnumDescriptor {
	return file_api_proto_team_TeamInfo_proto_enumTypes[4].Descriptor()
}

func (MuteMode) Type() protoreflect.EnumType {
	return &file_api_proto_team_TeamInfo_proto_enumTypes[4]
}

func (x MuteMode) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use MuteMode.Descriptor instead.
func (MuteMode) EnumDescriptor() ([]byte, []int) {
	return file_api_proto_team_TeamInfo_proto_rawDescGZIP(), []int{4}
}

//群被邀请模式：被邀请人的同意方式
type BeInviteMode int32

const (
	// 无定义
	BeInviteMode_Bim_Undefined BeInviteMode = 0
	//需要被邀请方同意
	BeInviteMode_Bim_NeedAuth BeInviteMode = 1
	//不需要被邀请方同意言
	BeInviteMode_Bim_NoAuth BeInviteMode = 2
)

// Enum value maps for BeInviteMode.
var (
	BeInviteMode_name = map[int32]string{
		0: "Bim_Undefined",
		1: "Bim_NeedAuth",
		2: "Bim_NoAuth",
	}
	BeInviteMode_value = map[string]int32{
		"Bim_Undefined": 0,
		"Bim_NeedAuth":  1,
		"Bim_NoAuth":    2,
	}
)

func (x BeInviteMode) Enum() *BeInviteMode {
	p := new(BeInviteMode)
	*p = x
	return p
}

func (x BeInviteMode) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (BeInviteMode) Descriptor() protoreflect.EnumDescriptor {
	return file_api_proto_team_TeamInfo_proto_enumTypes[5].Descriptor()
}

func (BeInviteMode) Type() protoreflect.EnumType {
	return &file_api_proto_team_TeamInfo_proto_enumTypes[5]
}

func (x BeInviteMode) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use BeInviteMode.Descriptor instead.
func (BeInviteMode) EnumDescriptor() ([]byte, []int) {
	return file_api_proto_team_TeamInfo_proto_rawDescGZIP(), []int{5}
}

//群邀请模式：谁可以邀请他人入群
type InviteMode int32

const (
	// 无定义
	InviteMode_Invite_Undefined InviteMode = 0
	//所有人都可以邀请其他人入群
	InviteMode_Invite_All InviteMode = 1
	//只有管理员可以邀请其他人入群
	InviteMode_Invite_Manager InviteMode = 2
	//邀请用户入群时需要管理员审核
	InviteMode_Invite_Check InviteMode = 3
)

// Enum value maps for InviteMode.
var (
	InviteMode_name = map[int32]string{
		0: "Invite_Undefined",
		1: "Invite_All",
		2: "Invite_Manager",
		3: "Invite_Check",
	}
	InviteMode_value = map[string]int32{
		"Invite_Undefined": 0,
		"Invite_All":       1,
		"Invite_Manager":   2,
		"Invite_Check":     3,
	}
)

func (x InviteMode) Enum() *InviteMode {
	p := new(InviteMode)
	*p = x
	return p
}

func (x InviteMode) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (InviteMode) Descriptor() protoreflect.EnumDescriptor {
	return file_api_proto_team_TeamInfo_proto_enumTypes[6].Descriptor()
}

func (InviteMode) Type() protoreflect.EnumType {
	return &file_api_proto_team_TeamInfo_proto_enumTypes[6]
}

func (x InviteMode) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use InviteMode.Descriptor instead.
func (InviteMode) EnumDescriptor() ([]byte, []int) {
	return file_api_proto_team_TeamInfo_proto_rawDescGZIP(), []int{6}
}

type UpdateMode int32

const (
	// 无定义
	UpdateMode_Update_Undefined UpdateMode = 0
	//	所有人可以修改
	UpdateMode_Update_All UpdateMode = 1
	//	只有管理员/群主可以修改
	UpdateMode_Update_Manager UpdateMode = 2
)

// Enum value maps for UpdateMode.
var (
	UpdateMode_name = map[int32]string{
		0: "Update_Undefined",
		1: "Update_All",
		2: "Update_Manager",
	}
	UpdateMode_value = map[string]int32{
		"Update_Undefined": 0,
		"Update_All":       1,
		"Update_Manager":   2,
	}
)

func (x UpdateMode) Enum() *UpdateMode {
	p := new(UpdateMode)
	*p = x
	return p
}

func (x UpdateMode) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (UpdateMode) Descriptor() protoreflect.EnumDescriptor {
	return file_api_proto_team_TeamInfo_proto_enumTypes[7].Descriptor()
}

func (UpdateMode) Type() protoreflect.EnumType {
	return &file_api_proto_team_TeamInfo_proto_enumTypes[7]
}

func (x UpdateMode) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use UpdateMode.Descriptor instead.
func (UpdateMode) EnumDescriptor() ([]byte, []int) {
	return file_api_proto_team_TeamInfo_proto_rawDescGZIP(), []int{7}
}

//冗余字段,记录接收消息提醒方式
type NotifyType int32

const (
	//无效
	NotifyType_Notify_Undefined NotifyType = 0
	//群全部消息提醒
	NotifyType_Notify_All NotifyType = 1
	//管理员消息提醒
	NotifyType_Notify_Manager NotifyType = 2
	//联系人提醒
	NotifyType_Notify_Contact NotifyType = 3
	//全部不提醒
	NotifyType_Notify_Mute NotifyType = 4
)

// Enum value maps for NotifyType.
var (
	NotifyType_name = map[int32]string{
		0: "Notify_Undefined",
		1: "Notify_All",
		2: "Notify_Manager",
		3: "Notify_Contact",
		4: "Notify_Mute",
	}
	NotifyType_value = map[string]int32{
		"Notify_Undefined": 0,
		"Notify_All":       1,
		"Notify_Manager":   2,
		"Notify_Contact":   3,
		"Notify_Mute":      4,
	}
)

func (x NotifyType) Enum() *NotifyType {
	p := new(NotifyType)
	*p = x
	return p
}

func (x NotifyType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (NotifyType) Descriptor() protoreflect.EnumDescriptor {
	return file_api_proto_team_TeamInfo_proto_enumTypes[8].Descriptor()
}

func (NotifyType) Type() protoreflect.EnumType {
	return &file_api_proto_team_TeamInfo_proto_enumTypes[8]
}

func (x NotifyType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use NotifyType.Descriptor instead.
func (NotifyType) EnumDescriptor() ([]byte, []int) {
	return file_api_proto_team_TeamInfo_proto_rawDescGZIP(), []int{8}
}

//群成员类型
type TeamMemberType int32

const (
	//无效
	TeamMemberType_Tmt_Undefined TeamMemberType = 0
	//待审核的申请加入用户
	TeamMemberType_Tmt_Apply TeamMemberType = 1
	//管理员
	TeamMemberType_Tmt_Manager TeamMemberType = 2
	//普通成员
	TeamMemberType_Tmt_Normal TeamMemberType = 3
	//创建者
	TeamMemberType_Tmt_Owner TeamMemberType = 4
)

// Enum value maps for TeamMemberType.
var (
	TeamMemberType_name = map[int32]string{
		0: "Tmt_Undefined",
		1: "Tmt_Apply",
		2: "Tmt_Manager",
		3: "Tmt_Normal",
		4: "Tmt_Owner",
	}
	TeamMemberType_value = map[string]int32{
		"Tmt_Undefined": 0,
		"Tmt_Apply":     1,
		"Tmt_Manager":   2,
		"Tmt_Normal":    3,
		"Tmt_Owner":     4,
	}
)

func (x TeamMemberType) Enum() *TeamMemberType {
	p := new(TeamMemberType)
	*p = x
	return p
}

func (x TeamMemberType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (TeamMemberType) Descriptor() protoreflect.EnumDescriptor {
	return file_api_proto_team_TeamInfo_proto_enumTypes[9].Descriptor()
}

func (TeamMemberType) Type() protoreflect.EnumType {
	return &file_api_proto_team_TeamInfo_proto_enumTypes[9]
}

func (x TeamMemberType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use TeamMemberType.Descriptor instead.
func (TeamMemberType) EnumDescriptor() ([]byte, []int) {
	return file_api_proto_team_TeamInfo_proto_rawDescGZIP(), []int{9}
}

//群组信息
type TeamField int32

const (
	TeamField_Tmf_Undefined TeamField = 0
	//群名称
	//是否必须:否
	TeamField_Tmf_Name TeamField = 1
	//群头像
	//是否必须:否
	TeamField_Tmf_Icon TeamField = 2
	//群公告
	//是否必须:否
	TeamField_Tmf_Announcement TeamField = 3
	//群简介
	//是否必须：否
	TeamField_Tmf_Introduce TeamField = 4
	//入群校验方式
	//是否必须：是
	TeamField_Tmf_VerifyType TeamField = 5
	//邀请模式
	//是否必须：是
	TeamField_Tmf_InviteMode TeamField = 6
	//群资料更新方式
	//是否必须：否
	TeamField_Tmf_UpdateTeamMode TeamField = 7
	//群资料扩展信息
	TeamField_Tmf_Ex TeamField = 8
)

// Enum value maps for TeamField.
var (
	TeamField_name = map[int32]string{
		0: "Tmf_Undefined",
		1: "Tmf_Name",
		2: "Tmf_Icon",
		3: "Tmf_Announcement",
		4: "Tmf_Introduce",
		5: "Tmf_VerifyType",
		6: "Tmf_InviteMode",
		7: "Tmf_UpdateTeamMode",
		8: "Tmf_Ex",
	}
	TeamField_value = map[string]int32{
		"Tmf_Undefined":      0,
		"Tmf_Name":           1,
		"Tmf_Icon":           2,
		"Tmf_Announcement":   3,
		"Tmf_Introduce":      4,
		"Tmf_VerifyType":     5,
		"Tmf_InviteMode":     6,
		"Tmf_UpdateTeamMode": 7,
		"Tmf_Ex":             8,
	}
)

func (x TeamField) Enum() *TeamField {
	p := new(TeamField)
	*p = x
	return p
}

func (x TeamField) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (TeamField) Descriptor() protoreflect.EnumDescriptor {
	return file_api_proto_team_TeamInfo_proto_enumTypes[10].Descriptor()
}

func (TeamField) Type() protoreflect.EnumType {
	return &file_api_proto_team_TeamInfo_proto_enumTypes[10]
}

func (x TeamField) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use TeamField.Descriptor instead.
func (TeamField) EnumDescriptor() ([]byte, []int) {
	return file_api_proto_team_TeamInfo_proto_rawDescGZIP(), []int{10}
}

//群组信息
type TeamInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// 群ID
	//是否必须：是
	TeamId string `protobuf:"bytes,1,opt,name=teamId,proto3" json:"teamId,omitempty"`
	//群名称
	//是否必须:是
	TeamName string `protobuf:"bytes,2,opt,name=teamName,proto3" json:"teamName,omitempty"`
	//群头像
	//是否必须:否
	Icon string `protobuf:"bytes,3,opt,name=icon,proto3" json:"icon,omitempty"`
	//群公告
	//是否必须:否
	Announcement string `protobuf:"bytes,4,opt,name=announcement,proto3" json:"announcement,omitempty"`
	//群简介
	//是否必须：否
	Introduce string `protobuf:"bytes,5,opt,name=introduce,proto3" json:"introduce,omitempty"`
	//群主id
	//是否必须：是
	Owner string `protobuf:"bytes,7,opt,name=owner,proto3" json:"owner,omitempty"`
	//群类型,枚举类型
	//是否必须:是
	Type TeamType `protobuf:"varint,8,opt,name=type,proto3,enum=cloud.lianmi.im.team.TeamType" json:"type,omitempty"`
	//校验模式
	//是否必须：是
	VerifyType VerifyType `protobuf:"varint,9,opt,name=verifyType,proto3,enum=cloud.lianmi.im.team.VerifyType" json:"verifyType,omitempty"`
	//成员上限
	//是否必须：是
	MemberLimit int32 `protobuf:"varint,10,opt,name=memberLimit,proto3" json:"memberLimit,omitempty"`
	//当前成员人数
	//取值范围200~2000
	//是否必须：是
	MemberNum int32 `protobuf:"varint,11,opt,name=memberNum,proto3" json:"memberNum,omitempty"`
	//群状态
	//是否必须：是
	Status TeamStatus `protobuf:"varint,12,opt,name=status,proto3,enum=cloud.lianmi.im.team.TeamStatus" json:"status,omitempty"`
	//发言方式
	//是否必须：是
	MuteType MuteMode `protobuf:"varint,13,opt,name=muteType,proto3,enum=cloud.lianmi.im.team.MuteMode" json:"muteType,omitempty"`
	//邀请模式
	//是否必须：是
	InviteMode InviteMode `protobuf:"varint,14,opt,name=inviteMode,proto3,enum=cloud.lianmi.im.team.InviteMode" json:"inviteMode,omitempty"`
	//群资料修改模式：谁可以修改群资料
	//是否必须：是
	//UpdateMode updateTeamMode = 15;
	//JSON扩展字段,由业务方解析
	//是否必填-否
	Ex string `protobuf:"bytes,16,opt,name=ex,proto3" json:"ex,omitempty"`
	//群组创建时间，unix时间戳
	//是否必须：是
	CreateAt uint64 `protobuf:"fixed64,17,opt,name=createAt,proto3" json:"createAt,omitempty"`
	//最后更新时间，unix时间戳
	//是否必须：是
	UpdateAt   uint64     `protobuf:"fixed64,18,opt,name=updateAt,proto3" json:"updateAt,omitempty"`
	NotifyType NotifyType `protobuf:"varint,20,opt,name=notifyType,proto3,enum=cloud.lianmi.im.team.NotifyType" json:"notifyType,omitempty"`
	IsMute     bool       `protobuf:"varint,21,opt,name=isMute,proto3" json:"isMute,omitempty"` //群是否被禁言
}

func (x *TeamInfo) Reset() {
	*x = TeamInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_team_TeamInfo_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TeamInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TeamInfo) ProtoMessage() {}

func (x *TeamInfo) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_team_TeamInfo_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TeamInfo.ProtoReflect.Descriptor instead.
func (*TeamInfo) Descriptor() ([]byte, []int) {
	return file_api_proto_team_TeamInfo_proto_rawDescGZIP(), []int{0}
}

func (x *TeamInfo) GetTeamId() string {
	if x != nil {
		return x.TeamId
	}
	return ""
}

func (x *TeamInfo) GetTeamName() string {
	if x != nil {
		return x.TeamName
	}
	return ""
}

func (x *TeamInfo) GetIcon() string {
	if x != nil {
		return x.Icon
	}
	return ""
}

func (x *TeamInfo) GetAnnouncement() string {
	if x != nil {
		return x.Announcement
	}
	return ""
}

func (x *TeamInfo) GetIntroduce() string {
	if x != nil {
		return x.Introduce
	}
	return ""
}

func (x *TeamInfo) GetOwner() string {
	if x != nil {
		return x.Owner
	}
	return ""
}

func (x *TeamInfo) GetType() TeamType {
	if x != nil {
		return x.Type
	}
	return TeamType_Tt_Undefined
}

func (x *TeamInfo) GetVerifyType() VerifyType {
	if x != nil {
		return x.VerifyType
	}
	return VerifyType_Vt_Undefined
}

func (x *TeamInfo) GetMemberLimit() int32 {
	if x != nil {
		return x.MemberLimit
	}
	return 0
}

func (x *TeamInfo) GetMemberNum() int32 {
	if x != nil {
		return x.MemberNum
	}
	return 0
}

func (x *TeamInfo) GetStatus() TeamStatus {
	if x != nil {
		return x.Status
	}
	return TeamStatus_Status_Init
}

func (x *TeamInfo) GetMuteType() MuteMode {
	if x != nil {
		return x.MuteType
	}
	return MuteMode_Mm_Undefined
}

func (x *TeamInfo) GetInviteMode() InviteMode {
	if x != nil {
		return x.InviteMode
	}
	return InviteMode_Invite_Undefined
}

func (x *TeamInfo) GetEx() string {
	if x != nil {
		return x.Ex
	}
	return ""
}

func (x *TeamInfo) GetCreateAt() uint64 {
	if x != nil {
		return x.CreateAt
	}
	return 0
}

func (x *TeamInfo) GetUpdateAt() uint64 {
	if x != nil {
		return x.UpdateAt
	}
	return 0
}

func (x *TeamInfo) GetNotifyType() NotifyType {
	if x != nil {
		return x.NotifyType
	}
	return NotifyType_Notify_Undefined
}

func (x *TeamInfo) GetIsMute() bool {
	if x != nil {
		return x.IsMute
	}
	return false
}

var File_api_proto_team_TeamInfo_proto protoreflect.FileDescriptor

var file_api_proto_team_TeamInfo_proto_rawDesc = []byte{
	0x0a, 0x1d, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x74, 0x65, 0x61, 0x6d,
	0x2f, 0x54, 0x65, 0x61, 0x6d, 0x49, 0x6e, 0x66, 0x6f, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x14, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d,
	0x2e, 0x74, 0x65, 0x61, 0x6d, 0x22, 0xba, 0x05, 0x0a, 0x08, 0x54, 0x65, 0x61, 0x6d, 0x49, 0x6e,
	0x66, 0x6f, 0x12, 0x16, 0x0a, 0x06, 0x74, 0x65, 0x61, 0x6d, 0x49, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x06, 0x74, 0x65, 0x61, 0x6d, 0x49, 0x64, 0x12, 0x1a, 0x0a, 0x08, 0x74, 0x65,
	0x61, 0x6d, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x74, 0x65,
	0x61, 0x6d, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x69, 0x63, 0x6f, 0x6e, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x69, 0x63, 0x6f, 0x6e, 0x12, 0x22, 0x0a, 0x0c, 0x61, 0x6e,
	0x6e, 0x6f, 0x75, 0x6e, 0x63, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0c, 0x61, 0x6e, 0x6e, 0x6f, 0x75, 0x6e, 0x63, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x12, 0x1c,
	0x0a, 0x09, 0x69, 0x6e, 0x74, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x09, 0x69, 0x6e, 0x74, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x65, 0x12, 0x14, 0x0a, 0x05,
	0x6f, 0x77, 0x6e, 0x65, 0x72, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x6f, 0x77, 0x6e,
	0x65, 0x72, 0x12, 0x32, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0e,
	0x32, 0x1e, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e,
	0x69, 0x6d, 0x2e, 0x74, 0x65, 0x61, 0x6d, 0x2e, 0x54, 0x65, 0x61, 0x6d, 0x54, 0x79, 0x70, 0x65,
	0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x40, 0x0a, 0x0a, 0x76, 0x65, 0x72, 0x69, 0x66, 0x79,
	0x54, 0x79, 0x70, 0x65, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x20, 0x2e, 0x63, 0x6c, 0x6f,
	0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x74, 0x65, 0x61,
	0x6d, 0x2e, 0x56, 0x65, 0x72, 0x69, 0x66, 0x79, 0x54, 0x79, 0x70, 0x65, 0x52, 0x0a, 0x76, 0x65,
	0x72, 0x69, 0x66, 0x79, 0x54, 0x79, 0x70, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x6d, 0x65, 0x6d, 0x62,
	0x65, 0x72, 0x4c, 0x69, 0x6d, 0x69, 0x74, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0b, 0x6d,
	0x65, 0x6d, 0x62, 0x65, 0x72, 0x4c, 0x69, 0x6d, 0x69, 0x74, 0x12, 0x1c, 0x0a, 0x09, 0x6d, 0x65,
	0x6d, 0x62, 0x65, 0x72, 0x4e, 0x75, 0x6d, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x05, 0x52, 0x09, 0x6d,
	0x65, 0x6d, 0x62, 0x65, 0x72, 0x4e, 0x75, 0x6d, 0x12, 0x38, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x20, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64,
	0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x74, 0x65, 0x61, 0x6d, 0x2e,
	0x54, 0x65, 0x61, 0x6d, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x12, 0x3a, 0x0a, 0x08, 0x6d, 0x75, 0x74, 0x65, 0x54, 0x79, 0x70, 0x65, 0x18, 0x0d,
	0x20, 0x01, 0x28, 0x0e, 0x32, 0x1e, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61,
	0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x74, 0x65, 0x61, 0x6d, 0x2e, 0x4d, 0x75, 0x74, 0x65,
	0x4d, 0x6f, 0x64, 0x65, 0x52, 0x08, 0x6d, 0x75, 0x74, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x40,
	0x0a, 0x0a, 0x69, 0x6e, 0x76, 0x69, 0x74, 0x65, 0x4d, 0x6f, 0x64, 0x65, 0x18, 0x0e, 0x20, 0x01,
	0x28, 0x0e, 0x32, 0x20, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d,
	0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x74, 0x65, 0x61, 0x6d, 0x2e, 0x49, 0x6e, 0x76, 0x69, 0x74, 0x65,
	0x4d, 0x6f, 0x64, 0x65, 0x52, 0x0a, 0x69, 0x6e, 0x76, 0x69, 0x74, 0x65, 0x4d, 0x6f, 0x64, 0x65,
	0x12, 0x0e, 0x0a, 0x02, 0x65, 0x78, 0x18, 0x10, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x65, 0x78,
	0x12, 0x1a, 0x0a, 0x08, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x41, 0x74, 0x18, 0x11, 0x20, 0x01,
	0x28, 0x06, 0x52, 0x08, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x41, 0x74, 0x12, 0x1a, 0x0a, 0x08,
	0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x41, 0x74, 0x18, 0x12, 0x20, 0x01, 0x28, 0x06, 0x52, 0x08,
	0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x41, 0x74, 0x12, 0x40, 0x0a, 0x0a, 0x6e, 0x6f, 0x74, 0x69,
	0x66, 0x79, 0x54, 0x79, 0x70, 0x65, 0x18, 0x14, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x20, 0x2e, 0x63,
	0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x74,
	0x65, 0x61, 0x6d, 0x2e, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x54, 0x79, 0x70, 0x65, 0x52, 0x0a,
	0x6e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x54, 0x79, 0x70, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x69, 0x73,
	0x4d, 0x75, 0x74, 0x65, 0x18, 0x15, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x69, 0x73, 0x4d, 0x75,
	0x74, 0x65, 0x2a, 0x5a, 0x0a, 0x08, 0x54, 0x65, 0x61, 0x6d, 0x54, 0x79, 0x70, 0x65, 0x12, 0x10,
	0x0a, 0x0c, 0x54, 0x74, 0x5f, 0x55, 0x6e, 0x64, 0x65, 0x66, 0x69, 0x6e, 0x65, 0x64, 0x10, 0x00,
	0x12, 0x0d, 0x0a, 0x09, 0x54, 0x74, 0x5f, 0x4e, 0x6f, 0x72, 0x6d, 0x61, 0x6c, 0x10, 0x01, 0x12,
	0x0f, 0x0a, 0x0b, 0x54, 0x74, 0x5f, 0x41, 0x64, 0x76, 0x61, 0x6e, 0x63, 0x65, 0x64, 0x10, 0x02,
	0x12, 0x0a, 0x0a, 0x06, 0x54, 0x74, 0x5f, 0x56, 0x69, 0x70, 0x10, 0x03, 0x12, 0x10, 0x0a, 0x0c,
	0x54, 0x74, 0x5f, 0x54, 0x65, 0x6d, 0x70, 0x6f, 0x72, 0x61, 0x72, 0x79, 0x10, 0x04, 0x2a, 0x4f,
	0x0a, 0x09, 0x51, 0x75, 0x65, 0x72, 0x79, 0x54, 0x79, 0x70, 0x65, 0x12, 0x12, 0x0a, 0x0e, 0x54,
	0x6d, 0x71, 0x74, 0x5f, 0x55, 0x6e, 0x64, 0x65, 0x66, 0x69, 0x6e, 0x65, 0x64, 0x10, 0x00, 0x12,
	0x0c, 0x0a, 0x08, 0x54, 0x6d, 0x71, 0x74, 0x5f, 0x41, 0x6c, 0x6c, 0x10, 0x01, 0x12, 0x10, 0x0a,
	0x0c, 0x54, 0x6d, 0x71, 0x74, 0x5f, 0x4d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x10, 0x02, 0x12,
	0x0e, 0x0a, 0x0a, 0x54, 0x6d, 0x71, 0x74, 0x5f, 0x4d, 0x75, 0x74, 0x65, 0x64, 0x10, 0x03, 0x2a,
	0x5a, 0x0a, 0x0a, 0x54, 0x65, 0x61, 0x6d, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x0f, 0x0a,
	0x0b, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x5f, 0x49, 0x6e, 0x69, 0x74, 0x10, 0x00, 0x12, 0x11,
	0x0a, 0x0d, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x5f, 0x4e, 0x6f, 0x72, 0x6d, 0x61, 0x6c, 0x10,
	0x01, 0x12, 0x12, 0x0a, 0x0e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x5f, 0x42, 0x6c, 0x6f, 0x63,
	0x6b, 0x65, 0x64, 0x10, 0x02, 0x12, 0x14, 0x0a, 0x10, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x5f,
	0x44, 0x69, 0x73, 0x4d, 0x69, 0x73, 0x73, 0x65, 0x64, 0x10, 0x03, 0x2a, 0x49, 0x0a, 0x0a, 0x56,
	0x65, 0x72, 0x69, 0x66, 0x79, 0x54, 0x79, 0x70, 0x65, 0x12, 0x10, 0x0a, 0x0c, 0x56, 0x74, 0x5f,
	0x55, 0x6e, 0x64, 0x65, 0x66, 0x69, 0x6e, 0x65, 0x64, 0x10, 0x00, 0x12, 0x0b, 0x0a, 0x07, 0x56,
	0x74, 0x5f, 0x46, 0x72, 0x65, 0x65, 0x10, 0x01, 0x12, 0x0c, 0x0a, 0x08, 0x56, 0x74, 0x5f, 0x41,
	0x70, 0x70, 0x6c, 0x79, 0x10, 0x02, 0x12, 0x0e, 0x0a, 0x0a, 0x56, 0x74, 0x5f, 0x50, 0x72, 0x69,
	0x76, 0x61, 0x74, 0x65, 0x10, 0x03, 0x2a, 0x4c, 0x0a, 0x08, 0x4d, 0x75, 0x74, 0x65, 0x4d, 0x6f,
	0x64, 0x65, 0x12, 0x10, 0x0a, 0x0c, 0x4d, 0x6d, 0x5f, 0x55, 0x6e, 0x64, 0x65, 0x66, 0x69, 0x6e,
	0x65, 0x64, 0x10, 0x00, 0x12, 0x0b, 0x0a, 0x07, 0x4d, 0x6d, 0x5f, 0x4e, 0x6f, 0x6e, 0x65, 0x10,
	0x01, 0x12, 0x0e, 0x0a, 0x0a, 0x4d, 0x6d, 0x5f, 0x4d, 0x75, 0x74, 0x65, 0x41, 0x4c, 0x4c, 0x10,
	0x02, 0x12, 0x11, 0x0a, 0x0d, 0x4d, 0x6d, 0x5f, 0x4d, 0x75, 0x74, 0x65, 0x4e, 0x6f, 0x72, 0x6d,
	0x61, 0x6c, 0x10, 0x03, 0x2a, 0x43, 0x0a, 0x0c, 0x42, 0x65, 0x49, 0x6e, 0x76, 0x69, 0x74, 0x65,
	0x4d, 0x6f, 0x64, 0x65, 0x12, 0x11, 0x0a, 0x0d, 0x42, 0x69, 0x6d, 0x5f, 0x55, 0x6e, 0x64, 0x65,
	0x66, 0x69, 0x6e, 0x65, 0x64, 0x10, 0x00, 0x12, 0x10, 0x0a, 0x0c, 0x42, 0x69, 0x6d, 0x5f, 0x4e,
	0x65, 0x65, 0x64, 0x41, 0x75, 0x74, 0x68, 0x10, 0x01, 0x12, 0x0e, 0x0a, 0x0a, 0x42, 0x69, 0x6d,
	0x5f, 0x4e, 0x6f, 0x41, 0x75, 0x74, 0x68, 0x10, 0x02, 0x2a, 0x58, 0x0a, 0x0a, 0x49, 0x6e, 0x76,
	0x69, 0x74, 0x65, 0x4d, 0x6f, 0x64, 0x65, 0x12, 0x14, 0x0a, 0x10, 0x49, 0x6e, 0x76, 0x69, 0x74,
	0x65, 0x5f, 0x55, 0x6e, 0x64, 0x65, 0x66, 0x69, 0x6e, 0x65, 0x64, 0x10, 0x00, 0x12, 0x0e, 0x0a,
	0x0a, 0x49, 0x6e, 0x76, 0x69, 0x74, 0x65, 0x5f, 0x41, 0x6c, 0x6c, 0x10, 0x01, 0x12, 0x12, 0x0a,
	0x0e, 0x49, 0x6e, 0x76, 0x69, 0x74, 0x65, 0x5f, 0x4d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x10,
	0x02, 0x12, 0x10, 0x0a, 0x0c, 0x49, 0x6e, 0x76, 0x69, 0x74, 0x65, 0x5f, 0x43, 0x68, 0x65, 0x63,
	0x6b, 0x10, 0x03, 0x2a, 0x46, 0x0a, 0x0a, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x4d, 0x6f, 0x64,
	0x65, 0x12, 0x14, 0x0a, 0x10, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x5f, 0x55, 0x6e, 0x64, 0x65,
	0x66, 0x69, 0x6e, 0x65, 0x64, 0x10, 0x00, 0x12, 0x0e, 0x0a, 0x0a, 0x55, 0x70, 0x64, 0x61, 0x74,
	0x65, 0x5f, 0x41, 0x6c, 0x6c, 0x10, 0x01, 0x12, 0x12, 0x0a, 0x0e, 0x55, 0x70, 0x64, 0x61, 0x74,
	0x65, 0x5f, 0x4d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x10, 0x02, 0x2a, 0x6b, 0x0a, 0x0a, 0x4e,
	0x6f, 0x74, 0x69, 0x66, 0x79, 0x54, 0x79, 0x70, 0x65, 0x12, 0x14, 0x0a, 0x10, 0x4e, 0x6f, 0x74,
	0x69, 0x66, 0x79, 0x5f, 0x55, 0x6e, 0x64, 0x65, 0x66, 0x69, 0x6e, 0x65, 0x64, 0x10, 0x00, 0x12,
	0x0e, 0x0a, 0x0a, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x5f, 0x41, 0x6c, 0x6c, 0x10, 0x01, 0x12,
	0x12, 0x0a, 0x0e, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x5f, 0x4d, 0x61, 0x6e, 0x61, 0x67, 0x65,
	0x72, 0x10, 0x02, 0x12, 0x12, 0x0a, 0x0e, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x79, 0x5f, 0x43, 0x6f,
	0x6e, 0x74, 0x61, 0x63, 0x74, 0x10, 0x03, 0x12, 0x0f, 0x0a, 0x0b, 0x4e, 0x6f, 0x74, 0x69, 0x66,
	0x79, 0x5f, 0x4d, 0x75, 0x74, 0x65, 0x10, 0x04, 0x2a, 0x62, 0x0a, 0x0e, 0x54, 0x65, 0x61, 0x6d,
	0x4d, 0x65, 0x6d, 0x62, 0x65, 0x72, 0x54, 0x79, 0x70, 0x65, 0x12, 0x11, 0x0a, 0x0d, 0x54, 0x6d,
	0x74, 0x5f, 0x55, 0x6e, 0x64, 0x65, 0x66, 0x69, 0x6e, 0x65, 0x64, 0x10, 0x00, 0x12, 0x0d, 0x0a,
	0x09, 0x54, 0x6d, 0x74, 0x5f, 0x41, 0x70, 0x70, 0x6c, 0x79, 0x10, 0x01, 0x12, 0x0f, 0x0a, 0x0b,
	0x54, 0x6d, 0x74, 0x5f, 0x4d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x10, 0x02, 0x12, 0x0e, 0x0a,
	0x0a, 0x54, 0x6d, 0x74, 0x5f, 0x4e, 0x6f, 0x72, 0x6d, 0x61, 0x6c, 0x10, 0x03, 0x12, 0x0d, 0x0a,
	0x09, 0x54, 0x6d, 0x74, 0x5f, 0x4f, 0x77, 0x6e, 0x65, 0x72, 0x10, 0x04, 0x2a, 0xaf, 0x01, 0x0a,
	0x09, 0x54, 0x65, 0x61, 0x6d, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x12, 0x11, 0x0a, 0x0d, 0x54, 0x6d,
	0x66, 0x5f, 0x55, 0x6e, 0x64, 0x65, 0x66, 0x69, 0x6e, 0x65, 0x64, 0x10, 0x00, 0x12, 0x0c, 0x0a,
	0x08, 0x54, 0x6d, 0x66, 0x5f, 0x4e, 0x61, 0x6d, 0x65, 0x10, 0x01, 0x12, 0x0c, 0x0a, 0x08, 0x54,
	0x6d, 0x66, 0x5f, 0x49, 0x63, 0x6f, 0x6e, 0x10, 0x02, 0x12, 0x14, 0x0a, 0x10, 0x54, 0x6d, 0x66,
	0x5f, 0x41, 0x6e, 0x6e, 0x6f, 0x75, 0x6e, 0x63, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x10, 0x03, 0x12,
	0x11, 0x0a, 0x0d, 0x54, 0x6d, 0x66, 0x5f, 0x49, 0x6e, 0x74, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x65,
	0x10, 0x04, 0x12, 0x12, 0x0a, 0x0e, 0x54, 0x6d, 0x66, 0x5f, 0x56, 0x65, 0x72, 0x69, 0x66, 0x79,
	0x54, 0x79, 0x70, 0x65, 0x10, 0x05, 0x12, 0x12, 0x0a, 0x0e, 0x54, 0x6d, 0x66, 0x5f, 0x49, 0x6e,
	0x76, 0x69, 0x74, 0x65, 0x4d, 0x6f, 0x64, 0x65, 0x10, 0x06, 0x12, 0x16, 0x0a, 0x12, 0x54, 0x6d,
	0x66, 0x5f, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x54, 0x65, 0x61, 0x6d, 0x4d, 0x6f, 0x64, 0x65,
	0x10, 0x07, 0x12, 0x0a, 0x0a, 0x06, 0x54, 0x6d, 0x66, 0x5f, 0x45, 0x78, 0x10, 0x08, 0x42, 0x2a,
	0x5a, 0x28, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x69, 0x61,
	0x6e, 0x6d, 0x69, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x74, 0x65, 0x61, 0x6d, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_api_proto_team_TeamInfo_proto_rawDescOnce sync.Once
	file_api_proto_team_TeamInfo_proto_rawDescData = file_api_proto_team_TeamInfo_proto_rawDesc
)

func file_api_proto_team_TeamInfo_proto_rawDescGZIP() []byte {
	file_api_proto_team_TeamInfo_proto_rawDescOnce.Do(func() {
		file_api_proto_team_TeamInfo_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_proto_team_TeamInfo_proto_rawDescData)
	})
	return file_api_proto_team_TeamInfo_proto_rawDescData
}

var file_api_proto_team_TeamInfo_proto_enumTypes = make([]protoimpl.EnumInfo, 11)
var file_api_proto_team_TeamInfo_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_api_proto_team_TeamInfo_proto_goTypes = []interface{}{
	(TeamType)(0),       // 0: cloud.lianmi.im.team.TeamType
	(QueryType)(0),      // 1: cloud.lianmi.im.team.QueryType
	(TeamStatus)(0),     // 2: cloud.lianmi.im.team.TeamStatus
	(VerifyType)(0),     // 3: cloud.lianmi.im.team.VerifyType
	(MuteMode)(0),       // 4: cloud.lianmi.im.team.MuteMode
	(BeInviteMode)(0),   // 5: cloud.lianmi.im.team.BeInviteMode
	(InviteMode)(0),     // 6: cloud.lianmi.im.team.InviteMode
	(UpdateMode)(0),     // 7: cloud.lianmi.im.team.UpdateMode
	(NotifyType)(0),     // 8: cloud.lianmi.im.team.NotifyType
	(TeamMemberType)(0), // 9: cloud.lianmi.im.team.TeamMemberType
	(TeamField)(0),      // 10: cloud.lianmi.im.team.TeamField
	(*TeamInfo)(nil),    // 11: cloud.lianmi.im.team.TeamInfo
}
var file_api_proto_team_TeamInfo_proto_depIdxs = []int32{
	0, // 0: cloud.lianmi.im.team.TeamInfo.type:type_name -> cloud.lianmi.im.team.TeamType
	3, // 1: cloud.lianmi.im.team.TeamInfo.verifyType:type_name -> cloud.lianmi.im.team.VerifyType
	2, // 2: cloud.lianmi.im.team.TeamInfo.status:type_name -> cloud.lianmi.im.team.TeamStatus
	4, // 3: cloud.lianmi.im.team.TeamInfo.muteType:type_name -> cloud.lianmi.im.team.MuteMode
	6, // 4: cloud.lianmi.im.team.TeamInfo.inviteMode:type_name -> cloud.lianmi.im.team.InviteMode
	8, // 5: cloud.lianmi.im.team.TeamInfo.notifyType:type_name -> cloud.lianmi.im.team.NotifyType
	6, // [6:6] is the sub-list for method output_type
	6, // [6:6] is the sub-list for method input_type
	6, // [6:6] is the sub-list for extension type_name
	6, // [6:6] is the sub-list for extension extendee
	0, // [0:6] is the sub-list for field type_name
}

func init() { file_api_proto_team_TeamInfo_proto_init() }
func file_api_proto_team_TeamInfo_proto_init() {
	if File_api_proto_team_TeamInfo_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_proto_team_TeamInfo_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TeamInfo); i {
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
			RawDescriptor: file_api_proto_team_TeamInfo_proto_rawDesc,
			NumEnums:      11,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_proto_team_TeamInfo_proto_goTypes,
		DependencyIndexes: file_api_proto_team_TeamInfo_proto_depIdxs,
		EnumInfos:         file_api_proto_team_TeamInfo_proto_enumTypes,
		MessageInfos:      file_api_proto_team_TeamInfo_proto_msgTypes,
	}.Build()
	File_api_proto_team_TeamInfo_proto = out.File
	file_api_proto_team_TeamInfo_proto_rawDesc = nil
	file_api_proto_team_TeamInfo_proto_goTypes = nil
	file_api_proto_team_TeamInfo_proto_depIdxs = nil
}
