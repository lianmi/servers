// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.7.0
// source: api/proto/user/store.proto

package user

import (
	proto "github.com/golang/protobuf/proto"
	global "github.com/lianmi/servers/api/proto/global"
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

// 店铺信息
type Store struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	StoreUUID          string           `protobuf:"bytes,1,opt,name=storeUUID,proto3" json:"storeUUID,omitempty"`                                        //店铺的uuid
	BusinessUsername   string           `protobuf:"bytes,2,opt,name=businessUsername,proto3" json:"businessUsername,omitempty"`                          //店铺的商户注册号
	Avatar             string           `protobuf:"bytes,3,opt,name=avatar,proto3" json:"avatar,omitempty"`                                              //店铺的头像，与商户头像一致
	ImageUrl           string           `protobuf:"bytes,4,opt,name=imageUrl,proto3" json:"imageUrl,omitempty"`                                          //店铺的外景照片或产品图片
	StoreType          global.StoreType `protobuf:"varint,5,opt,name=storeType,proto3,enum=cloud.lianmi.im.global.StoreType" json:"storeType,omitempty"` //店铺类型,对应Global.proto里的StoreType枚举
	Introductory       string           `protobuf:"bytes,6,opt,name=introductory,proto3" json:"introductory,omitempty"`                                  //商店简介 Text文本类型
	Province           string           `protobuf:"bytes,7,opt,name=province,proto3" json:"province,omitempty"`                                          //省份, 如广东省
	City               string           `protobuf:"bytes,8,opt,name=city,proto3" json:"city,omitempty"`                                                  //城市，如广州市
	County             string           `protobuf:"bytes,9,opt,name=county,proto3" json:"county,omitempty"`                                              //区，如天河区
	Street             string           `protobuf:"bytes,10,opt,name=street,proto3" json:"street,omitempty"`                                             //街道
	Address            string           `protobuf:"bytes,11,opt,name=address,proto3" json:"address,omitempty"`                                           //地址
	Branchesname       string           `protobuf:"bytes,12,opt,name=branchesname,proto3" json:"branchesname,omitempty"`                                 //店铺名称
	Keys               string           `protobuf:"bytes,13,opt,name=keys,proto3" json:"keys,omitempty"`                                                 //店铺经营范围搜索关键字
	LegalPerson        string           `protobuf:"bytes,14,opt,name=legalPerson,proto3" json:"legalPerson,omitempty"`                                   //法人姓名
	LegalIdentityCard  string           `protobuf:"bytes,15,opt,name=legalIdentityCard,proto3" json:"legalIdentityCard,omitempty"`                       //法人身份证
	BusinessLicenseUrl string           `protobuf:"bytes,16,opt,name=businessLicenseUrl,proto3" json:"businessLicenseUrl,omitempty"`                     //营业执照阿里云url
	Wechat             string           `protobuf:"bytes,17,opt,name=wechat,proto3" json:"wechat,omitempty"`                                             //商户地址的纬度
	Longitude          float64          `protobuf:"fixed64,18,opt,name=longitude,proto3" json:"longitude,omitempty"`                                     //商户地址的经度
	Latitude           float64          `protobuf:"fixed64,19,opt,name=latitude,proto3" json:"latitude,omitempty"`                                       //商户地址的纬度
	AuditState         int32            `protobuf:"varint,20,opt,name=auditState,proto3" json:"auditState,omitempty"`                                    //商户审核状态， 0-预审核，1-已审核，2-占位
	CreatedAt          uint64           `protobuf:"fixed64,21,opt,name=createdAt,proto3" json:"createdAt,omitempty"`                                     //用户注册时间,Unix时间戳
	UpdatedAt          uint64           `protobuf:"fixed64,22,opt,name=updatedAt,proto3" json:"updatedAt,omitempty"`                                     //用户资料最后更新时间,Unix时间戳
	Likes              uint64           `protobuf:"fixed64,23,opt,name=likes,proto3" json:"likes,omitempty"`                                             //用户点赞数
}

func (x *Store) Reset() {
	*x = Store{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_user_store_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Store) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Store) ProtoMessage() {}

func (x *Store) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_user_store_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Store.ProtoReflect.Descriptor instead.
func (*Store) Descriptor() ([]byte, []int) {
	return file_api_proto_user_store_proto_rawDescGZIP(), []int{0}
}

func (x *Store) GetStoreUUID() string {
	if x != nil {
		return x.StoreUUID
	}
	return ""
}

func (x *Store) GetBusinessUsername() string {
	if x != nil {
		return x.BusinessUsername
	}
	return ""
}

func (x *Store) GetAvatar() string {
	if x != nil {
		return x.Avatar
	}
	return ""
}

func (x *Store) GetImageUrl() string {
	if x != nil {
		return x.ImageUrl
	}
	return ""
}

func (x *Store) GetStoreType() global.StoreType {
	if x != nil {
		return x.StoreType
	}
	return global.StoreType_ST_Undefined
}

func (x *Store) GetIntroductory() string {
	if x != nil {
		return x.Introductory
	}
	return ""
}

func (x *Store) GetProvince() string {
	if x != nil {
		return x.Province
	}
	return ""
}

func (x *Store) GetCity() string {
	if x != nil {
		return x.City
	}
	return ""
}

func (x *Store) GetCounty() string {
	if x != nil {
		return x.County
	}
	return ""
}

func (x *Store) GetStreet() string {
	if x != nil {
		return x.Street
	}
	return ""
}

func (x *Store) GetAddress() string {
	if x != nil {
		return x.Address
	}
	return ""
}

func (x *Store) GetBranchesname() string {
	if x != nil {
		return x.Branchesname
	}
	return ""
}

func (x *Store) GetKeys() string {
	if x != nil {
		return x.Keys
	}
	return ""
}

func (x *Store) GetLegalPerson() string {
	if x != nil {
		return x.LegalPerson
	}
	return ""
}

func (x *Store) GetLegalIdentityCard() string {
	if x != nil {
		return x.LegalIdentityCard
	}
	return ""
}

func (x *Store) GetBusinessLicenseUrl() string {
	if x != nil {
		return x.BusinessLicenseUrl
	}
	return ""
}

func (x *Store) GetWechat() string {
	if x != nil {
		return x.Wechat
	}
	return ""
}

func (x *Store) GetLongitude() float64 {
	if x != nil {
		return x.Longitude
	}
	return 0
}

func (x *Store) GetLatitude() float64 {
	if x != nil {
		return x.Latitude
	}
	return 0
}

func (x *Store) GetAuditState() int32 {
	if x != nil {
		return x.AuditState
	}
	return 0
}

func (x *Store) GetCreatedAt() uint64 {
	if x != nil {
		return x.CreatedAt
	}
	return 0
}

func (x *Store) GetUpdatedAt() uint64 {
	if x != nil {
		return x.UpdatedAt
	}
	return 0
}

func (x *Store) GetLikes() uint64 {
	if x != nil {
		return x.Likes
	}
	return 0
}

type BusinessUserUploadLicenseReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Businessusername   string `protobuf:"bytes,1,opt,name=businessusername,proto3" json:"businessusername,omitempty"`     //店铺的商户注册号
	BusinessLicenseUrl string `protobuf:"bytes,2,opt,name=businessLicenseUrl,proto3" json:"businessLicenseUrl,omitempty"` //营业执照阿里云url
}

func (x *BusinessUserUploadLicenseReq) Reset() {
	*x = BusinessUserUploadLicenseReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_user_store_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BusinessUserUploadLicenseReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BusinessUserUploadLicenseReq) ProtoMessage() {}

func (x *BusinessUserUploadLicenseReq) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_user_store_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BusinessUserUploadLicenseReq.ProtoReflect.Descriptor instead.
func (*BusinessUserUploadLicenseReq) Descriptor() ([]byte, []int) {
	return file_api_proto_user_store_proto_rawDescGZIP(), []int{1}
}

func (x *BusinessUserUploadLicenseReq) GetBusinessusername() string {
	if x != nil {
		return x.Businessusername
	}
	return ""
}

func (x *BusinessUserUploadLicenseReq) GetBusinessLicenseUrl() string {
	if x != nil {
		return x.BusinessLicenseUrl
	}
	return ""
}

//获取当前用户对所有店铺点赞情况, UI会保存在本地表里,  UI主动发起同步
type UserLikesResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//用户注册账号
	Username string `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
	//某用户的所有店铺点赞列表
	Businessusernames []string `protobuf:"bytes,2,rep,name=businessusernames,proto3" json:"businessusernames,omitempty"`
}

func (x *UserLikesResp) Reset() {
	*x = UserLikesResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_user_store_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UserLikesResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UserLikesResp) ProtoMessage() {}

func (x *UserLikesResp) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_user_store_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UserLikesResp.ProtoReflect.Descriptor instead.
func (*UserLikesResp) Descriptor() ([]byte, []int) {
	return file_api_proto_user_store_proto_rawDescGZIP(), []int{2}
}

func (x *UserLikesResp) GetUsername() string {
	if x != nil {
		return x.Username
	}
	return ""
}

func (x *UserLikesResp) GetBusinessusernames() []string {
	if x != nil {
		return x.Businessusernames
	}
	return nil
}

//获取店铺的所有点赞的用户列表
type StoreLikesResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//店铺商户注册账号
	BusinessUsername string `protobuf:"bytes,1,opt,name=businessUsername,proto3" json:"businessUsername,omitempty"`
	//点赞用户列表
	Usernames []string `protobuf:"bytes,2,rep,name=usernames,proto3" json:"usernames,omitempty"`
}

func (x *StoreLikesResp) Reset() {
	*x = StoreLikesResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_user_store_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StoreLikesResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StoreLikesResp) ProtoMessage() {}

func (x *StoreLikesResp) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_user_store_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StoreLikesResp.ProtoReflect.Descriptor instead.
func (*StoreLikesResp) Descriptor() ([]byte, []int) {
	return file_api_proto_user_store_proto_rawDescGZIP(), []int{3}
}

func (x *StoreLikesResp) GetBusinessUsername() string {
	if x != nil {
		return x.BusinessUsername
	}
	return ""
}

func (x *StoreLikesResp) GetUsernames() []string {
	if x != nil {
		return x.Usernames
	}
	return nil
}

var File_api_proto_user_store_proto protoreflect.FileDescriptor

var file_api_proto_user_store_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x75, 0x73, 0x65, 0x72,
	0x2f, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x14, 0x63, 0x6c,
	0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x75, 0x73,
	0x65, 0x72, 0x1a, 0x1d, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x6c,
	0x6f, 0x62, 0x61, 0x6c, 0x2f, 0x47, 0x6c, 0x6f, 0x62, 0x61, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x22, 0xe0, 0x05, 0x0a, 0x05, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x73,
	0x74, 0x6f, 0x72, 0x65, 0x55, 0x55, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09,
	0x73, 0x74, 0x6f, 0x72, 0x65, 0x55, 0x55, 0x49, 0x44, 0x12, 0x2a, 0x0a, 0x10, 0x62, 0x75, 0x73,
	0x69, 0x6e, 0x65, 0x73, 0x73, 0x55, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x10, 0x62, 0x75, 0x73, 0x69, 0x6e, 0x65, 0x73, 0x73, 0x55, 0x73, 0x65,
	0x72, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x76, 0x61, 0x74, 0x61, 0x72, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x61, 0x76, 0x61, 0x74, 0x61, 0x72, 0x12, 0x1a, 0x0a,
	0x08, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x55, 0x72, 0x6c, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x08, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x55, 0x72, 0x6c, 0x12, 0x3f, 0x0a, 0x09, 0x73, 0x74, 0x6f,
	0x72, 0x65, 0x54, 0x79, 0x70, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x21, 0x2e, 0x63,
	0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x67,
	0x6c, 0x6f, 0x62, 0x61, 0x6c, 0x2e, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x54, 0x79, 0x70, 0x65, 0x52,
	0x09, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x22, 0x0a, 0x0c, 0x69, 0x6e,
	0x74, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x6f, 0x72, 0x79, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0c, 0x69, 0x6e, 0x74, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x6f, 0x72, 0x79, 0x12, 0x1a,
	0x0a, 0x08, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x6e, 0x63, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x08, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x6e, 0x63, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x63, 0x69,
	0x74, 0x79, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x63, 0x69, 0x74, 0x79, 0x12, 0x16,
	0x0a, 0x06, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x79, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06,
	0x63, 0x6f, 0x75, 0x6e, 0x74, 0x79, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x72, 0x65, 0x65, 0x74,
	0x18, 0x0a, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x74, 0x72, 0x65, 0x65, 0x74, 0x12, 0x18,
	0x0a, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x22, 0x0a, 0x0c, 0x62, 0x72, 0x61, 0x6e,
	0x63, 0x68, 0x65, 0x73, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c,
	0x62, 0x72, 0x61, 0x6e, 0x63, 0x68, 0x65, 0x73, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04,
	0x6b, 0x65, 0x79, 0x73, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6b, 0x65, 0x79, 0x73,
	0x12, 0x20, 0x0a, 0x0b, 0x6c, 0x65, 0x67, 0x61, 0x6c, 0x50, 0x65, 0x72, 0x73, 0x6f, 0x6e, 0x18,
	0x0e, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x6c, 0x65, 0x67, 0x61, 0x6c, 0x50, 0x65, 0x72, 0x73,
	0x6f, 0x6e, 0x12, 0x2c, 0x0a, 0x11, 0x6c, 0x65, 0x67, 0x61, 0x6c, 0x49, 0x64, 0x65, 0x6e, 0x74,
	0x69, 0x74, 0x79, 0x43, 0x61, 0x72, 0x64, 0x18, 0x0f, 0x20, 0x01, 0x28, 0x09, 0x52, 0x11, 0x6c,
	0x65, 0x67, 0x61, 0x6c, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x43, 0x61, 0x72, 0x64,
	0x12, 0x2e, 0x0a, 0x12, 0x62, 0x75, 0x73, 0x69, 0x6e, 0x65, 0x73, 0x73, 0x4c, 0x69, 0x63, 0x65,
	0x6e, 0x73, 0x65, 0x55, 0x72, 0x6c, 0x18, 0x10, 0x20, 0x01, 0x28, 0x09, 0x52, 0x12, 0x62, 0x75,
	0x73, 0x69, 0x6e, 0x65, 0x73, 0x73, 0x4c, 0x69, 0x63, 0x65, 0x6e, 0x73, 0x65, 0x55, 0x72, 0x6c,
	0x12, 0x16, 0x0a, 0x06, 0x77, 0x65, 0x63, 0x68, 0x61, 0x74, 0x18, 0x11, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x06, 0x77, 0x65, 0x63, 0x68, 0x61, 0x74, 0x12, 0x1c, 0x0a, 0x09, 0x6c, 0x6f, 0x6e, 0x67,
	0x69, 0x74, 0x75, 0x64, 0x65, 0x18, 0x12, 0x20, 0x01, 0x28, 0x01, 0x52, 0x09, 0x6c, 0x6f, 0x6e,
	0x67, 0x69, 0x74, 0x75, 0x64, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x6c, 0x61, 0x74, 0x69, 0x74, 0x75,
	0x64, 0x65, 0x18, 0x13, 0x20, 0x01, 0x28, 0x01, 0x52, 0x08, 0x6c, 0x61, 0x74, 0x69, 0x74, 0x75,
	0x64, 0x65, 0x12, 0x1e, 0x0a, 0x0a, 0x61, 0x75, 0x64, 0x69, 0x74, 0x53, 0x74, 0x61, 0x74, 0x65,
	0x18, 0x14, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0a, 0x61, 0x75, 0x64, 0x69, 0x74, 0x53, 0x74, 0x61,
	0x74, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x18,
	0x15, 0x20, 0x01, 0x28, 0x06, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74,
	0x12, 0x1c, 0x0a, 0x09, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x18, 0x16, 0x20,
	0x01, 0x28, 0x06, 0x52, 0x09, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x12, 0x14,
	0x0a, 0x05, 0x6c, 0x69, 0x6b, 0x65, 0x73, 0x18, 0x17, 0x20, 0x01, 0x28, 0x06, 0x52, 0x05, 0x6c,
	0x69, 0x6b, 0x65, 0x73, 0x22, 0x7a, 0x0a, 0x1c, 0x42, 0x75, 0x73, 0x69, 0x6e, 0x65, 0x73, 0x73,
	0x55, 0x73, 0x65, 0x72, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x4c, 0x69, 0x63, 0x65, 0x6e, 0x73,
	0x65, 0x52, 0x65, 0x71, 0x12, 0x2a, 0x0a, 0x10, 0x62, 0x75, 0x73, 0x69, 0x6e, 0x65, 0x73, 0x73,
	0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x10,
	0x62, 0x75, 0x73, 0x69, 0x6e, 0x65, 0x73, 0x73, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65,
	0x12, 0x2e, 0x0a, 0x12, 0x62, 0x75, 0x73, 0x69, 0x6e, 0x65, 0x73, 0x73, 0x4c, 0x69, 0x63, 0x65,
	0x6e, 0x73, 0x65, 0x55, 0x72, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x12, 0x62, 0x75,
	0x73, 0x69, 0x6e, 0x65, 0x73, 0x73, 0x4c, 0x69, 0x63, 0x65, 0x6e, 0x73, 0x65, 0x55, 0x72, 0x6c,
	0x22, 0x59, 0x0a, 0x0d, 0x55, 0x73, 0x65, 0x72, 0x4c, 0x69, 0x6b, 0x65, 0x73, 0x52, 0x65, 0x73,
	0x70, 0x12, 0x1a, 0x0a, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x08, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x2c, 0x0a,
	0x11, 0x62, 0x75, 0x73, 0x69, 0x6e, 0x65, 0x73, 0x73, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d,
	0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x11, 0x62, 0x75, 0x73, 0x69, 0x6e, 0x65,
	0x73, 0x73, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x22, 0x5a, 0x0a, 0x0e, 0x53,
	0x74, 0x6f, 0x72, 0x65, 0x4c, 0x69, 0x6b, 0x65, 0x73, 0x52, 0x65, 0x73, 0x70, 0x12, 0x2a, 0x0a,
	0x10, 0x62, 0x75, 0x73, 0x69, 0x6e, 0x65, 0x73, 0x73, 0x55, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x10, 0x62, 0x75, 0x73, 0x69, 0x6e, 0x65, 0x73,
	0x73, 0x55, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x75, 0x73, 0x65,
	0x72, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x09, 0x75, 0x73,
	0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x42, 0x2a, 0x5a, 0x28, 0x67, 0x69, 0x74, 0x68, 0x75,
	0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2f, 0x73, 0x65, 0x72,
	0x76, 0x65, 0x72, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x75,
	0x73, 0x65, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_proto_user_store_proto_rawDescOnce sync.Once
	file_api_proto_user_store_proto_rawDescData = file_api_proto_user_store_proto_rawDesc
)

func file_api_proto_user_store_proto_rawDescGZIP() []byte {
	file_api_proto_user_store_proto_rawDescOnce.Do(func() {
		file_api_proto_user_store_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_proto_user_store_proto_rawDescData)
	})
	return file_api_proto_user_store_proto_rawDescData
}

var file_api_proto_user_store_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_api_proto_user_store_proto_goTypes = []interface{}{
	(*Store)(nil),                        // 0: cloud.lianmi.im.user.Store
	(*BusinessUserUploadLicenseReq)(nil), // 1: cloud.lianmi.im.user.BusinessUserUploadLicenseReq
	(*UserLikesResp)(nil),                // 2: cloud.lianmi.im.user.UserLikesResp
	(*StoreLikesResp)(nil),               // 3: cloud.lianmi.im.user.StoreLikesResp
	(global.StoreType)(0),                // 4: cloud.lianmi.im.global.StoreType
}
var file_api_proto_user_store_proto_depIdxs = []int32{
	4, // 0: cloud.lianmi.im.user.Store.storeType:type_name -> cloud.lianmi.im.global.StoreType
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_api_proto_user_store_proto_init() }
func file_api_proto_user_store_proto_init() {
	if File_api_proto_user_store_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_proto_user_store_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Store); i {
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
		file_api_proto_user_store_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BusinessUserUploadLicenseReq); i {
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
		file_api_proto_user_store_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UserLikesResp); i {
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
		file_api_proto_user_store_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StoreLikesResp); i {
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
			RawDescriptor: file_api_proto_user_store_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_proto_user_store_proto_goTypes,
		DependencyIndexes: file_api_proto_user_store_proto_depIdxs,
		MessageInfos:      file_api_proto_user_store_proto_msgTypes,
	}.Build()
	File_api_proto_user_store_proto = out.File
	file_api_proto_user_store_proto_rawDesc = nil
	file_api_proto_user_store_proto_goTypes = nil
	file_api_proto_user_store_proto_depIdxs = nil
}
