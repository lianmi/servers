// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.7.0
// source: api/proto/user/User.proto

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

//性别枚举定义
type Gender int32

const (
	Gender_Sex_Unknown Gender = 0
	Gender_Sex_Male    Gender = 1
	Gender_Sex_Female  Gender = 2
)

// Enum value maps for Gender.
var (
	Gender_name = map[int32]string{
		0: "Sex_Unknown",
		1: "Sex_Male",
		2: "Sex_Female",
	}
	Gender_value = map[string]int32{
		"Sex_Unknown": 0,
		"Sex_Male":    1,
		"Sex_Female":  2,
	}
)

func (x Gender) Enum() *Gender {
	p := new(Gender)
	*p = x
	return p
}

func (x Gender) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Gender) Descriptor() protoreflect.EnumDescriptor {
	return file_api_proto_user_User_proto_enumTypes[0].Descriptor()
}

func (Gender) Type() protoreflect.EnumType {
	return &file_api_proto_user_User_proto_enumTypes[0]
}

func (x Gender) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Gender.Descriptor instead.
func (Gender) EnumDescriptor() ([]byte, []int) {
	return file_api_proto_user_User_proto_rawDescGZIP(), []int{0}
}

//账号类型
type UserType int32

const (
	UserType_Ut_Undefined UserType = 0
	//一般用户
	UserType_Ut_Normal UserType = 1
	//网点用户
	UserType_Ut_Business UserType = 2
	//操作员, 例如admin
	UserType_Ut_Operator UserType = 10086
)

// Enum value maps for UserType.
var (
	UserType_name = map[int32]string{
		0:     "Ut_Undefined",
		1:     "Ut_Normal",
		2:     "Ut_Business",
		10086: "Ut_Operator",
	}
	UserType_value = map[string]int32{
		"Ut_Undefined": 0,
		"Ut_Normal":    1,
		"Ut_Business":  2,
		"Ut_Operator":  10086,
	}
)

func (x UserType) Enum() *UserType {
	p := new(UserType)
	*p = x
	return p
}

func (x UserType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (UserType) Descriptor() protoreflect.EnumDescriptor {
	return file_api_proto_user_User_proto_enumTypes[1].Descriptor()
}

func (UserType) Type() protoreflect.EnumType {
	return &file_api_proto_user_User_proto_enumTypes[1]
}

func (x UserType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use UserType.Descriptor instead.
func (UserType) EnumDescriptor() ([]byte, []int) {
	return file_api_proto_user_User_proto_rawDescGZIP(), []int{1}
}

type UserState int32

const (
	UserState_Ss_Unknow  UserState = 0
	UserState_Ss_Normal  UserState = 1
	UserState_Ss_Blocked UserState = 2
)

// Enum value maps for UserState.
var (
	UserState_name = map[int32]string{
		0: "Ss_Unknow",
		1: "Ss_Normal",
		2: "Ss_Blocked",
	}
	UserState_value = map[string]int32{
		"Ss_Unknow":  0,
		"Ss_Normal":  1,
		"Ss_Blocked": 2,
	}
)

func (x UserState) Enum() *UserState {
	p := new(UserState)
	*p = x
	return p
}

func (x UserState) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (UserState) Descriptor() protoreflect.EnumDescriptor {
	return file_api_proto_user_User_proto_enumTypes[2].Descriptor()
}

func (UserState) Type() protoreflect.EnumType {
	return &file_api_proto_user_User_proto_enumTypes[2]
}

func (x UserState) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use UserState.Descriptor instead.
func (UserState) EnumDescriptor() ([]byte, []int) {
	return file_api_proto_user_User_proto_rawDescGZIP(), []int{2}
}

//获取用户加好友的选项,3是默认
type AllowType int32

const (
	AllowType_UAT_Unknow      AllowType = 0
	AllowType_UAT_AllowAny    AllowType = 1
	AllowType_UAT_DenyAny     AllowType = 2
	AllowType_UAT_NeedConfirm AllowType = 3
)

// Enum value maps for AllowType.
var (
	AllowType_name = map[int32]string{
		0: "UAT_Unknow",
		1: "UAT_AllowAny",
		2: "UAT_DenyAny",
		3: "UAT_NeedConfirm",
	}
	AllowType_value = map[string]int32{
		"UAT_Unknow":      0,
		"UAT_AllowAny":    1,
		"UAT_DenyAny":     2,
		"UAT_NeedConfirm": 3,
	}
)

func (x AllowType) Enum() *AllowType {
	p := new(AllowType)
	*p = x
	return p
}

func (x AllowType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (AllowType) Descriptor() protoreflect.EnumDescriptor {
	return file_api_proto_user_User_proto_enumTypes[3].Descriptor()
}

func (AllowType) Type() protoreflect.EnumType {
	return &file_api_proto_user_User_proto_enumTypes[3]
}

func (x AllowType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use AllowType.Descriptor instead.
func (AllowType) EnumDescriptor() ([]byte, []int) {
	return file_api_proto_user_User_proto_rawDescGZIP(), []int{3}
}

//根据用户ID批量获取用户信息,登录后拉取其他用户资料,添加好友查询好友资料
type GetUsersReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Usernames []string `protobuf:"bytes,1,rep,name=usernames,proto3" json:"usernames,omitempty"`
}

func (x *GetUsersReq) Reset() {
	*x = GetUsersReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_user_User_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetUsersReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetUsersReq) ProtoMessage() {}

func (x *GetUsersReq) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_user_User_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetUsersReq.ProtoReflect.Descriptor instead.
func (*GetUsersReq) Descriptor() ([]byte, []int) {
	return file_api_proto_user_User_proto_rawDescGZIP(), []int{0}
}

func (x *GetUsersReq) GetUsernames() []string {
	if x != nil {
		return x.Usernames
	}
	return nil
}

type GetUsersResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Users []*User `protobuf:"bytes,1,rep,name=users,proto3" json:"users,omitempty"`
}

func (x *GetUsersResp) Reset() {
	*x = GetUsersResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_user_User_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetUsersResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetUsersResp) ProtoMessage() {}

func (x *GetUsersResp) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_user_User_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetUsersResp.ProtoReflect.Descriptor instead.
func (*GetUsersResp) Descriptor() ([]byte, []int) {
	return file_api_proto_user_User_proto_rawDescGZIP(), []int{1}
}

func (x *GetUsersResp) GetUsers() []*User {
	if x != nil {
		return x.Users
	}
	return nil
}

// 用户信息
type User struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id               uint64    `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"` //ID
	Smscode          string    `protobuf:"bytes,2,opt,name=smscode,proto3" json:"smscode,omitempty"`
	Username         string    `protobuf:"bytes,3,opt,name=username,proto3" json:"username,omitempty"`
	Password         string    `protobuf:"bytes,4,opt,name=password,proto3" json:"password,omitempty"` //密码 是否必填-是
	Gender           Gender    `protobuf:"varint,5,opt,name=gender,proto3,enum=cloud.lianmi.im.user.Gender" json:"gender,omitempty"`
	Nick             string    `protobuf:"bytes,6,opt,name=nick,proto3" json:"nick,omitempty"`
	Avatar           string    `protobuf:"bytes,7,opt,name=avatar,proto3" json:"avatar,omitempty"`
	Label            string    `protobuf:"bytes,8,opt,name=label,proto3" json:"label,omitempty"`
	Mobile           string    `protobuf:"bytes,9,opt,name=mobile,proto3" json:"mobile,omitempty"`
	Email            string    `protobuf:"bytes,10,opt,name=email,proto3" json:"email,omitempty"`
	UserType         UserType  `protobuf:"varint,11,opt,name=userType,proto3,enum=cloud.lianmi.im.user.UserType" json:"userType,omitempty"`
	State            UserState `protobuf:"varint,12,opt,name=state,proto3,enum=cloud.lianmi.im.user.UserState" json:"state,omitempty"`
	Extend           string    `protobuf:"bytes,13,opt,name=extend,proto3" json:"extend,omitempty"`
	ReferrerUsername string    `protobuf:"bytes,14,opt,name=referrerUsername,proto3" json:"referrerUsername,omitempty"`
	ContactPerson    string    `protobuf:"bytes,15,opt,name=contactPerson,proto3" json:"contactPerson,omitempty"`
	CreatedAt        uint64    `protobuf:"fixed64,16,opt,name=createdAt,proto3" json:"createdAt,omitempty"` //用户注册时间,Unix时间戳
	UpdatedAt        uint64    `protobuf:"fixed64,17,opt,name=updatedAt,proto3" json:"updatedAt,omitempty"` //用户资料最后更新时间,Unix时间戳
}

func (x *User) Reset() {
	*x = User{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_user_User_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *User) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*User) ProtoMessage() {}

func (x *User) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_user_User_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use User.ProtoReflect.Descriptor instead.
func (*User) Descriptor() ([]byte, []int) {
	return file_api_proto_user_User_proto_rawDescGZIP(), []int{2}
}

func (x *User) GetId() uint64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *User) GetSmscode() string {
	if x != nil {
		return x.Smscode
	}
	return ""
}

func (x *User) GetUsername() string {
	if x != nil {
		return x.Username
	}
	return ""
}

func (x *User) GetPassword() string {
	if x != nil {
		return x.Password
	}
	return ""
}

func (x *User) GetGender() Gender {
	if x != nil {
		return x.Gender
	}
	return Gender_Sex_Unknown
}

func (x *User) GetNick() string {
	if x != nil {
		return x.Nick
	}
	return ""
}

func (x *User) GetAvatar() string {
	if x != nil {
		return x.Avatar
	}
	return ""
}

func (x *User) GetLabel() string {
	if x != nil {
		return x.Label
	}
	return ""
}

func (x *User) GetMobile() string {
	if x != nil {
		return x.Mobile
	}
	return ""
}

func (x *User) GetEmail() string {
	if x != nil {
		return x.Email
	}
	return ""
}

func (x *User) GetUserType() UserType {
	if x != nil {
		return x.UserType
	}
	return UserType_Ut_Undefined
}

func (x *User) GetState() UserState {
	if x != nil {
		return x.State
	}
	return UserState_Ss_Unknow
}

func (x *User) GetExtend() string {
	if x != nil {
		return x.Extend
	}
	return ""
}

func (x *User) GetReferrerUsername() string {
	if x != nil {
		return x.ReferrerUsername
	}
	return ""
}

func (x *User) GetContactPerson() string {
	if x != nil {
		return x.ContactPerson
	}
	return ""
}

func (x *User) GetCreatedAt() uint64 {
	if x != nil {
		return x.CreatedAt
	}
	return 0
}

func (x *User) GetUpdatedAt() uint64 {
	if x != nil {
		return x.UpdatedAt
	}
	return 0
}

//多条件不定参数批量分页获取用户列表
type QueryUsersReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Mobile           string    `protobuf:"bytes,1,opt,name=mobile,proto3" json:"mobile,omitempty"`                                         //手机号
	UserType         UserType  `protobuf:"varint,2,opt,name=userType,proto3,enum=cloud.lianmi.im.user.UserType" json:"userType,omitempty"` //用户类型
	State            UserState `protobuf:"varint,3,opt,name=state,proto3,enum=cloud.lianmi.im.user.UserState" json:"state,omitempty"`      //状态
	ReferrerUsername string    `protobuf:"bytes,4,opt,name=referrerUsername,proto3" json:"referrerUsername,omitempty"`                     //推荐人
	ContactPerson    string    `protobuf:"bytes,5,opt,name=contactPerson,proto3" json:"contactPerson,omitempty"`                           //联系人
	StartAt          uint64    `protobuf:"fixed64,6,opt,name=startAt,proto3" json:"startAt,omitempty"`                                     //注册开始时间, 按时间段查询
	EndAt            uint64    `protobuf:"fixed64,7,opt,name=endAt,proto3" json:"endAt,omitempty"`                                         //注册结束时间
	Page             int32     `protobuf:"varint,8,opt,name=page,proto3" json:"page,omitempty"`                                            //页数,第几页
	PageSize         int32     `protobuf:"varint,9,opt,name=pageSize,proto3" json:"pageSize,omitempty"`                                    //每页记录数量
}

func (x *QueryUsersReq) Reset() {
	*x = QueryUsersReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_user_User_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QueryUsersReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryUsersReq) ProtoMessage() {}

func (x *QueryUsersReq) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_user_User_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QueryUsersReq.ProtoReflect.Descriptor instead.
func (*QueryUsersReq) Descriptor() ([]byte, []int) {
	return file_api_proto_user_User_proto_rawDescGZIP(), []int{3}
}

func (x *QueryUsersReq) GetMobile() string {
	if x != nil {
		return x.Mobile
	}
	return ""
}

func (x *QueryUsersReq) GetUserType() UserType {
	if x != nil {
		return x.UserType
	}
	return UserType_Ut_Undefined
}

func (x *QueryUsersReq) GetState() UserState {
	if x != nil {
		return x.State
	}
	return UserState_Ss_Unknow
}

func (x *QueryUsersReq) GetReferrerUsername() string {
	if x != nil {
		return x.ReferrerUsername
	}
	return ""
}

func (x *QueryUsersReq) GetContactPerson() string {
	if x != nil {
		return x.ContactPerson
	}
	return ""
}

func (x *QueryUsersReq) GetStartAt() uint64 {
	if x != nil {
		return x.StartAt
	}
	return 0
}

func (x *QueryUsersReq) GetEndAt() uint64 {
	if x != nil {
		return x.EndAt
	}
	return 0
}

func (x *QueryUsersReq) GetPage() int32 {
	if x != nil {
		return x.Page
	}
	return 0
}

func (x *QueryUsersReq) GetPageSize() int32 {
	if x != nil {
		return x.PageSize
	}
	return 0
}

type QueryUsersResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Users []*User `protobuf:"bytes,1,rep,name=users,proto3" json:"users,omitempty"`   //用户列表
	Total uint64  `protobuf:"fixed64,2,opt,name=total,proto3" json:"total,omitempty"` //按请求参数的pageSize计算出来的总页数
}

func (x *QueryUsersResp) Reset() {
	*x = QueryUsersResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_user_User_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QueryUsersResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryUsersResp) ProtoMessage() {}

func (x *QueryUsersResp) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_user_User_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QueryUsersResp.ProtoReflect.Descriptor instead.
func (*QueryUsersResp) Descriptor() ([]byte, []int) {
	return file_api_proto_user_User_proto_rawDescGZIP(), []int{4}
}

func (x *QueryUsersResp) GetUsers() []*User {
	if x != nil {
		return x.Users
	}
	return nil
}

func (x *QueryUsersResp) GetTotal() uint64 {
	if x != nil {
		return x.Total
	}
	return 0
}

var File_api_proto_user_User_proto protoreflect.FileDescriptor

var file_api_proto_user_User_proto_rawDesc = []byte{
	0x0a, 0x19, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x75, 0x73, 0x65, 0x72,
	0x2f, 0x55, 0x73, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x14, 0x63, 0x6c, 0x6f,
	0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x75, 0x73, 0x65,
	0x72, 0x22, 0x2b, 0x0a, 0x0b, 0x47, 0x65, 0x74, 0x55, 0x73, 0x65, 0x72, 0x73, 0x52, 0x65, 0x71,
	0x12, 0x1c, 0x0a, 0x09, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x18, 0x01, 0x20,
	0x03, 0x28, 0x09, 0x52, 0x09, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x22, 0x40,
	0x0a, 0x0c, 0x47, 0x65, 0x74, 0x55, 0x73, 0x65, 0x72, 0x73, 0x52, 0x65, 0x73, 0x70, 0x12, 0x30,
	0x0a, 0x05, 0x75, 0x73, 0x65, 0x72, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1a, 0x2e,
	0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e,
	0x75, 0x73, 0x65, 0x72, 0x2e, 0x55, 0x73, 0x65, 0x72, 0x52, 0x05, 0x75, 0x73, 0x65, 0x72, 0x73,
	0x22, 0xa7, 0x04, 0x0a, 0x04, 0x55, 0x73, 0x65, 0x72, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x02, 0x69, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x6d, 0x73,
	0x63, 0x6f, 0x64, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x73, 0x6d, 0x73, 0x63,
	0x6f, 0x64, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x12,
	0x1a, 0x0a, 0x08, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x70, 0x61, 0x73, 0x73, 0x77, 0x6f, 0x72, 0x64, 0x12, 0x34, 0x0a, 0x06, 0x67,
	0x65, 0x6e, 0x64, 0x65, 0x72, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1c, 0x2e, 0x63, 0x6c,
	0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x75, 0x73,
	0x65, 0x72, 0x2e, 0x47, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x52, 0x06, 0x67, 0x65, 0x6e, 0x64, 0x65,
	0x72, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x69, 0x63, 0x6b, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x6e, 0x69, 0x63, 0x6b, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x76, 0x61, 0x74, 0x61, 0x72, 0x18,
	0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x61, 0x76, 0x61, 0x74, 0x61, 0x72, 0x12, 0x14, 0x0a,
	0x05, 0x6c, 0x61, 0x62, 0x65, 0x6c, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x6c, 0x61,
	0x62, 0x65, 0x6c, 0x12, 0x16, 0x0a, 0x06, 0x6d, 0x6f, 0x62, 0x69, 0x6c, 0x65, 0x18, 0x09, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x06, 0x6d, 0x6f, 0x62, 0x69, 0x6c, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x65,
	0x6d, 0x61, 0x69, 0x6c, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x6d, 0x61, 0x69,
	0x6c, 0x12, 0x3a, 0x0a, 0x08, 0x75, 0x73, 0x65, 0x72, 0x54, 0x79, 0x70, 0x65, 0x18, 0x0b, 0x20,
	0x01, 0x28, 0x0e, 0x32, 0x1e, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e,
	0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x75, 0x73, 0x65, 0x72, 0x2e, 0x55, 0x73, 0x65, 0x72, 0x54,
	0x79, 0x70, 0x65, 0x52, 0x08, 0x75, 0x73, 0x65, 0x72, 0x54, 0x79, 0x70, 0x65, 0x12, 0x35, 0x0a,
	0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1f, 0x2e, 0x63,
	0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x75,
	0x73, 0x65, 0x72, 0x2e, 0x55, 0x73, 0x65, 0x72, 0x53, 0x74, 0x61, 0x74, 0x65, 0x52, 0x05, 0x73,
	0x74, 0x61, 0x74, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x64, 0x18, 0x0d,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x65, 0x78, 0x74, 0x65, 0x6e, 0x64, 0x12, 0x2a, 0x0a, 0x10,
	0x72, 0x65, 0x66, 0x65, 0x72, 0x72, 0x65, 0x72, 0x55, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x0e, 0x20, 0x01, 0x28, 0x09, 0x52, 0x10, 0x72, 0x65, 0x66, 0x65, 0x72, 0x72, 0x65, 0x72,
	0x55, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x24, 0x0a, 0x0d, 0x63, 0x6f, 0x6e, 0x74,
	0x61, 0x63, 0x74, 0x50, 0x65, 0x72, 0x73, 0x6f, 0x6e, 0x18, 0x0f, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0d, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x50, 0x65, 0x72, 0x73, 0x6f, 0x6e, 0x12, 0x1c,
	0x0a, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x18, 0x10, 0x20, 0x01, 0x28,
	0x06, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x12, 0x1c, 0x0a, 0x09,
	0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x18, 0x11, 0x20, 0x01, 0x28, 0x06, 0x52,
	0x09, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x22, 0xcc, 0x02, 0x0a, 0x0d, 0x51,
	0x75, 0x65, 0x72, 0x79, 0x55, 0x73, 0x65, 0x72, 0x73, 0x52, 0x65, 0x71, 0x12, 0x16, 0x0a, 0x06,
	0x6d, 0x6f, 0x62, 0x69, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6d, 0x6f,
	0x62, 0x69, 0x6c, 0x65, 0x12, 0x3a, 0x0a, 0x08, 0x75, 0x73, 0x65, 0x72, 0x54, 0x79, 0x70, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1e, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c,
	0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x75, 0x73, 0x65, 0x72, 0x2e, 0x55, 0x73,
	0x65, 0x72, 0x54, 0x79, 0x70, 0x65, 0x52, 0x08, 0x75, 0x73, 0x65, 0x72, 0x54, 0x79, 0x70, 0x65,
	0x12, 0x35, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32,
	0x1f, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69,
	0x6d, 0x2e, 0x75, 0x73, 0x65, 0x72, 0x2e, 0x55, 0x73, 0x65, 0x72, 0x53, 0x74, 0x61, 0x74, 0x65,
	0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x12, 0x2a, 0x0a, 0x10, 0x72, 0x65, 0x66, 0x65, 0x72,
	0x72, 0x65, 0x72, 0x55, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x10, 0x72, 0x65, 0x66, 0x65, 0x72, 0x72, 0x65, 0x72, 0x55, 0x73, 0x65, 0x72, 0x6e,
	0x61, 0x6d, 0x65, 0x12, 0x24, 0x0a, 0x0d, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x63, 0x74, 0x50, 0x65,
	0x72, 0x73, 0x6f, 0x6e, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x63, 0x6f, 0x6e, 0x74,
	0x61, 0x63, 0x74, 0x50, 0x65, 0x72, 0x73, 0x6f, 0x6e, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x74, 0x61,
	0x72, 0x74, 0x41, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x06, 0x52, 0x07, 0x73, 0x74, 0x61, 0x72,
	0x74, 0x41, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x6e, 0x64, 0x41, 0x74, 0x18, 0x07, 0x20, 0x01,
	0x28, 0x06, 0x52, 0x05, 0x65, 0x6e, 0x64, 0x41, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x61, 0x67,
	0x65, 0x18, 0x08, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x70, 0x61, 0x67, 0x65, 0x12, 0x1a, 0x0a,
	0x08, 0x70, 0x61, 0x67, 0x65, 0x53, 0x69, 0x7a, 0x65, 0x18, 0x09, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x08, 0x70, 0x61, 0x67, 0x65, 0x53, 0x69, 0x7a, 0x65, 0x22, 0x58, 0x0a, 0x0e, 0x51, 0x75, 0x65,
	0x72, 0x79, 0x55, 0x73, 0x65, 0x72, 0x73, 0x52, 0x65, 0x73, 0x70, 0x12, 0x30, 0x0a, 0x05, 0x75,
	0x73, 0x65, 0x72, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x63, 0x6c, 0x6f,
	0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x75, 0x73, 0x65,
	0x72, 0x2e, 0x55, 0x73, 0x65, 0x72, 0x52, 0x05, 0x75, 0x73, 0x65, 0x72, 0x73, 0x12, 0x14, 0x0a,
	0x05, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x06, 0x52, 0x05, 0x74, 0x6f,
	0x74, 0x61, 0x6c, 0x2a, 0x37, 0x0a, 0x06, 0x47, 0x65, 0x6e, 0x64, 0x65, 0x72, 0x12, 0x0f, 0x0a,
	0x0b, 0x53, 0x65, 0x78, 0x5f, 0x55, 0x6e, 0x6b, 0x6e, 0x6f, 0x77, 0x6e, 0x10, 0x00, 0x12, 0x0c,
	0x0a, 0x08, 0x53, 0x65, 0x78, 0x5f, 0x4d, 0x61, 0x6c, 0x65, 0x10, 0x01, 0x12, 0x0e, 0x0a, 0x0a,
	0x53, 0x65, 0x78, 0x5f, 0x46, 0x65, 0x6d, 0x61, 0x6c, 0x65, 0x10, 0x02, 0x2a, 0x4e, 0x0a, 0x08,
	0x55, 0x73, 0x65, 0x72, 0x54, 0x79, 0x70, 0x65, 0x12, 0x10, 0x0a, 0x0c, 0x55, 0x74, 0x5f, 0x55,
	0x6e, 0x64, 0x65, 0x66, 0x69, 0x6e, 0x65, 0x64, 0x10, 0x00, 0x12, 0x0d, 0x0a, 0x09, 0x55, 0x74,
	0x5f, 0x4e, 0x6f, 0x72, 0x6d, 0x61, 0x6c, 0x10, 0x01, 0x12, 0x0f, 0x0a, 0x0b, 0x55, 0x74, 0x5f,
	0x42, 0x75, 0x73, 0x69, 0x6e, 0x65, 0x73, 0x73, 0x10, 0x02, 0x12, 0x10, 0x0a, 0x0b, 0x55, 0x74,
	0x5f, 0x4f, 0x70, 0x65, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x10, 0xe6, 0x4e, 0x2a, 0x39, 0x0a, 0x09,
	0x55, 0x73, 0x65, 0x72, 0x53, 0x74, 0x61, 0x74, 0x65, 0x12, 0x0d, 0x0a, 0x09, 0x53, 0x73, 0x5f,
	0x55, 0x6e, 0x6b, 0x6e, 0x6f, 0x77, 0x10, 0x00, 0x12, 0x0d, 0x0a, 0x09, 0x53, 0x73, 0x5f, 0x4e,
	0x6f, 0x72, 0x6d, 0x61, 0x6c, 0x10, 0x01, 0x12, 0x0e, 0x0a, 0x0a, 0x53, 0x73, 0x5f, 0x42, 0x6c,
	0x6f, 0x63, 0x6b, 0x65, 0x64, 0x10, 0x02, 0x2a, 0x53, 0x0a, 0x09, 0x41, 0x6c, 0x6c, 0x6f, 0x77,
	0x54, 0x79, 0x70, 0x65, 0x12, 0x0e, 0x0a, 0x0a, 0x55, 0x41, 0x54, 0x5f, 0x55, 0x6e, 0x6b, 0x6e,
	0x6f, 0x77, 0x10, 0x00, 0x12, 0x10, 0x0a, 0x0c, 0x55, 0x41, 0x54, 0x5f, 0x41, 0x6c, 0x6c, 0x6f,
	0x77, 0x41, 0x6e, 0x79, 0x10, 0x01, 0x12, 0x0f, 0x0a, 0x0b, 0x55, 0x41, 0x54, 0x5f, 0x44, 0x65,
	0x6e, 0x79, 0x41, 0x6e, 0x79, 0x10, 0x02, 0x12, 0x13, 0x0a, 0x0f, 0x55, 0x41, 0x54, 0x5f, 0x4e,
	0x65, 0x65, 0x64, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x72, 0x6d, 0x10, 0x03, 0x42, 0x2a, 0x5a, 0x28,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x69, 0x61, 0x6e, 0x6d,
	0x69, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2f, 0x75, 0x73, 0x65, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_proto_user_User_proto_rawDescOnce sync.Once
	file_api_proto_user_User_proto_rawDescData = file_api_proto_user_User_proto_rawDesc
)

func file_api_proto_user_User_proto_rawDescGZIP() []byte {
	file_api_proto_user_User_proto_rawDescOnce.Do(func() {
		file_api_proto_user_User_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_proto_user_User_proto_rawDescData)
	})
	return file_api_proto_user_User_proto_rawDescData
}

var file_api_proto_user_User_proto_enumTypes = make([]protoimpl.EnumInfo, 4)
var file_api_proto_user_User_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_api_proto_user_User_proto_goTypes = []interface{}{
	(Gender)(0),            // 0: cloud.lianmi.im.user.Gender
	(UserType)(0),          // 1: cloud.lianmi.im.user.UserType
	(UserState)(0),         // 2: cloud.lianmi.im.user.UserState
	(AllowType)(0),         // 3: cloud.lianmi.im.user.AllowType
	(*GetUsersReq)(nil),    // 4: cloud.lianmi.im.user.GetUsersReq
	(*GetUsersResp)(nil),   // 5: cloud.lianmi.im.user.GetUsersResp
	(*User)(nil),           // 6: cloud.lianmi.im.user.User
	(*QueryUsersReq)(nil),  // 7: cloud.lianmi.im.user.QueryUsersReq
	(*QueryUsersResp)(nil), // 8: cloud.lianmi.im.user.QueryUsersResp
}
var file_api_proto_user_User_proto_depIdxs = []int32{
	6, // 0: cloud.lianmi.im.user.GetUsersResp.users:type_name -> cloud.lianmi.im.user.User
	0, // 1: cloud.lianmi.im.user.User.gender:type_name -> cloud.lianmi.im.user.Gender
	1, // 2: cloud.lianmi.im.user.User.userType:type_name -> cloud.lianmi.im.user.UserType
	2, // 3: cloud.lianmi.im.user.User.state:type_name -> cloud.lianmi.im.user.UserState
	1, // 4: cloud.lianmi.im.user.QueryUsersReq.userType:type_name -> cloud.lianmi.im.user.UserType
	2, // 5: cloud.lianmi.im.user.QueryUsersReq.state:type_name -> cloud.lianmi.im.user.UserState
	6, // 6: cloud.lianmi.im.user.QueryUsersResp.users:type_name -> cloud.lianmi.im.user.User
	7, // [7:7] is the sub-list for method output_type
	7, // [7:7] is the sub-list for method input_type
	7, // [7:7] is the sub-list for extension type_name
	7, // [7:7] is the sub-list for extension extendee
	0, // [0:7] is the sub-list for field type_name
}

func init() { file_api_proto_user_User_proto_init() }
func file_api_proto_user_User_proto_init() {
	if File_api_proto_user_User_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_proto_user_User_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetUsersReq); i {
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
		file_api_proto_user_User_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetUsersResp); i {
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
		file_api_proto_user_User_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*User); i {
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
		file_api_proto_user_User_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*QueryUsersReq); i {
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
		file_api_proto_user_User_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*QueryUsersResp); i {
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
			RawDescriptor: file_api_proto_user_User_proto_rawDesc,
			NumEnums:      4,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_proto_user_User_proto_goTypes,
		DependencyIndexes: file_api_proto_user_User_proto_depIdxs,
		EnumInfos:         file_api_proto_user_User_proto_enumTypes,
		MessageInfos:      file_api_proto_user_User_proto_msgTypes,
	}.Build()
	File_api_proto_user_User_proto = out.File
	file_api_proto_user_User_proto_rawDesc = nil
	file_api_proto_user_User_proto_goTypes = nil
	file_api_proto_user_User_proto_depIdxs = nil
}
