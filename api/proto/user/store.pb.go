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
	Businessusername   string           `protobuf:"bytes,2,opt,name=businessusername,proto3" json:"businessusername,omitempty"`                          //店铺的商户注册号
	Avatar             string           `protobuf:"bytes,3,opt,name=avatar,proto3" json:"avatar,omitempty"`                                              //店铺的头像，与商户头像一致
	StoreType          global.StoreType `protobuf:"varint,4,opt,name=storeType,proto3,enum=cloud.lianmi.im.global.StoreType" json:"storeType,omitempty"` //店铺类型,对应Global.proto里的StoreType枚举
	Introductory       string           `protobuf:"bytes,5,opt,name=introductory,proto3" json:"introductory,omitempty"`                                  //商店简介 Text文本类型
	Province           string           `protobuf:"bytes,6,opt,name=province,proto3" json:"province,omitempty"`                                          //省份, 如广东省
	City               string           `protobuf:"bytes,7,opt,name=city,proto3" json:"city,omitempty"`                                                  //城市，如广州市
	County             string           `protobuf:"bytes,8,opt,name=county,proto3" json:"county,omitempty"`                                              //区，如天河区
	Street             string           `protobuf:"bytes,9,opt,name=street,proto3" json:"street,omitempty"`                                              //街道
	Address            string           `protobuf:"bytes,10,opt,name=address,proto3" json:"address,omitempty"`                                           //地址
	Branchesname       string           `protobuf:"bytes,11,opt,name=branchesname,proto3" json:"branchesname,omitempty"`                                 //店铺名称
	Keys               string           `protobuf:"bytes,12,opt,name=keys,proto3" json:"keys,omitempty"`                                                 //店铺经营范围搜索关键字
	LegalPerson        string           `protobuf:"bytes,13,opt,name=legalPerson,proto3" json:"legalPerson,omitempty"`                                   //法人姓名
	LegalIdentityCard  string           `protobuf:"bytes,14,opt,name=legalIdentityCard,proto3" json:"legalIdentityCard,omitempty"`                       //法人身份证
	BusinessLicenseUrl string           `protobuf:"bytes,15,opt,name=businessLicenseUrl,proto3" json:"businessLicenseUrl,omitempty"`                     //营业执照阿里云url
	Wechat             string           `protobuf:"bytes,16,opt,name=wechat,proto3" json:"wechat,omitempty"`                                             //商户地址的纬度
	Longitude          float64          `protobuf:"fixed64,17,opt,name=longitude,proto3" json:"longitude,omitempty"`                                     //商户地址的经度
	Latitude           float64          `protobuf:"fixed64,18,opt,name=latitude,proto3" json:"latitude,omitempty"`                                       //商户地址的纬度
	CreatedAt          uint64           `protobuf:"fixed64,19,opt,name=createdAt,proto3" json:"createdAt,omitempty"`                                     //用户注册时间,Unix时间戳
	UpdatedAt          uint64           `protobuf:"fixed64,20,opt,name=updatedAt,proto3" json:"updatedAt,omitempty"`                                     //用户资料最后更新时间,Unix时间戳
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

func (x *Store) GetBusinessusername() string {
	if x != nil {
		return x.Businessusername
	}
	return ""
}

func (x *Store) GetAvatar() string {
	if x != nil {
		return x.Avatar
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

//======查询经纬度范围内的商户列表=====//
type QueryStoresNearbyReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//店铺类型,对应Global.proto里的StoreType枚举
	StoreType global.StoreType `protobuf:"varint,1,opt,name=storeType,proto3,enum=cloud.lianmi.im.global.StoreType" json:"storeType,omitempty"`
	//商户经营范围搜索关键字, 用半角的逗号隔开
	Keys string `protobuf:"bytes,2,opt,name=keys,proto3" json:"keys,omitempty"`
	//用户当前位置的经度
	Longitude float64 `protobuf:"fixed64,3,opt,name=longitude,proto3" json:"longitude,omitempty"`
	//用户当前位置的经度
	Latitude float64 `protobuf:"fixed64,4,opt,name=latitude,proto3" json:"latitude,omitempty"`
	//半径范围, 默认10km
	Radius float64 `protobuf:"fixed64,5,opt,name=radius,proto3" json:"radius,omitempty"`
	//省份, 可选
	Province string `protobuf:"bytes,6,opt,name=province,proto3" json:"province,omitempty"`
	//城市, 可选
	City string `protobuf:"bytes,7,opt,name=city,proto3" json:"city,omitempty"`
	//区, 可选
	County string `protobuf:"bytes,8,opt,name=county,proto3" json:"county,omitempty"`
	//街道, 可选
	Street string `protobuf:"bytes,9,opt,name=street,proto3" json:"street,omitempty"`
	//页数,第几页
	//默认1
	//是否必填-否
	Page int32 `protobuf:"varint,10,opt,name=page,proto3" json:"page,omitempty"` // [default=1];
	//每页成员数量
	//默认20,最大只允许100
	//是否必填-否
	Limit int32 `protobuf:"varint,11,opt,name=limit,proto3" json:"limit,omitempty"` // [default=20];
}

func (x *QueryStoresNearbyReq) Reset() {
	*x = QueryStoresNearbyReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_user_store_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QueryStoresNearbyReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryStoresNearbyReq) ProtoMessage() {}

func (x *QueryStoresNearbyReq) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use QueryStoresNearbyReq.ProtoReflect.Descriptor instead.
func (*QueryStoresNearbyReq) Descriptor() ([]byte, []int) {
	return file_api_proto_user_store_proto_rawDescGZIP(), []int{2}
}

func (x *QueryStoresNearbyReq) GetStoreType() global.StoreType {
	if x != nil {
		return x.StoreType
	}
	return global.StoreType_ST_Undefined
}

func (x *QueryStoresNearbyReq) GetKeys() string {
	if x != nil {
		return x.Keys
	}
	return ""
}

func (x *QueryStoresNearbyReq) GetLongitude() float64 {
	if x != nil {
		return x.Longitude
	}
	return 0
}

func (x *QueryStoresNearbyReq) GetLatitude() float64 {
	if x != nil {
		return x.Latitude
	}
	return 0
}

func (x *QueryStoresNearbyReq) GetRadius() float64 {
	if x != nil {
		return x.Radius
	}
	return 0
}

func (x *QueryStoresNearbyReq) GetProvince() string {
	if x != nil {
		return x.Province
	}
	return ""
}

func (x *QueryStoresNearbyReq) GetCity() string {
	if x != nil {
		return x.City
	}
	return ""
}

func (x *QueryStoresNearbyReq) GetCounty() string {
	if x != nil {
		return x.County
	}
	return ""
}

func (x *QueryStoresNearbyReq) GetStreet() string {
	if x != nil {
		return x.Street
	}
	return ""
}

func (x *QueryStoresNearbyReq) GetPage() int32 {
	if x != nil {
		return x.Page
	}
	return 0
}

func (x *QueryStoresNearbyReq) GetLimit() int32 {
	if x != nil {
		return x.Limit
	}
	return 0
}

type QueryStoresNearbyResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//总页数
	TotalPage uint64 `protobuf:"fixed64,1,opt,name=totalPage,proto3" json:"totalPage,omitempty"`
	//搜索结果列表
	Stores []*Store `protobuf:"bytes,2,rep,name=stores,proto3" json:"stores,omitempty"`
}

func (x *QueryStoresNearbyResp) Reset() {
	*x = QueryStoresNearbyResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_user_store_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QueryStoresNearbyResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryStoresNearbyResp) ProtoMessage() {}

func (x *QueryStoresNearbyResp) ProtoReflect() protoreflect.Message {
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

// Deprecated: Use QueryStoresNearbyResp.ProtoReflect.Descriptor instead.
func (*QueryStoresNearbyResp) Descriptor() ([]byte, []int) {
	return file_api_proto_user_store_proto_rawDescGZIP(), []int{3}
}

func (x *QueryStoresNearbyResp) GetTotalPage() uint64 {
	if x != nil {
		return x.TotalPage
	}
	return 0
}

func (x *QueryStoresNearbyResp) GetStores() []*Store {
	if x != nil {
		return x.Stores
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
	0x6f, 0x22, 0x8e, 0x05, 0x0a, 0x05, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x73,
	0x74, 0x6f, 0x72, 0x65, 0x55, 0x55, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09,
	0x73, 0x74, 0x6f, 0x72, 0x65, 0x55, 0x55, 0x49, 0x44, 0x12, 0x2a, 0x0a, 0x10, 0x62, 0x75, 0x73,
	0x69, 0x6e, 0x65, 0x73, 0x73, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x10, 0x62, 0x75, 0x73, 0x69, 0x6e, 0x65, 0x73, 0x73, 0x75, 0x73, 0x65,
	0x72, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x61, 0x76, 0x61, 0x74, 0x61, 0x72, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x61, 0x76, 0x61, 0x74, 0x61, 0x72, 0x12, 0x3f, 0x0a,
	0x09, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x54, 0x79, 0x70, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0e,
	0x32, 0x21, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e,
	0x69, 0x6d, 0x2e, 0x67, 0x6c, 0x6f, 0x62, 0x61, 0x6c, 0x2e, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x54,
	0x79, 0x70, 0x65, 0x52, 0x09, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x22,
	0x0a, 0x0c, 0x69, 0x6e, 0x74, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x6f, 0x72, 0x79, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x69, 0x6e, 0x74, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x6f,
	0x72, 0x79, 0x12, 0x1a, 0x0a, 0x08, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x6e, 0x63, 0x65, 0x18, 0x06,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x6e, 0x63, 0x65, 0x12, 0x12,
	0x0a, 0x04, 0x63, 0x69, 0x74, 0x79, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x63, 0x69,
	0x74, 0x79, 0x12, 0x16, 0x0a, 0x06, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x79, 0x18, 0x08, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x06, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x79, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74,
	0x72, 0x65, 0x65, 0x74, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x74, 0x72, 0x65,
	0x65, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x0a, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x22, 0x0a, 0x0c,
	0x62, 0x72, 0x61, 0x6e, 0x63, 0x68, 0x65, 0x73, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x0b, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0c, 0x62, 0x72, 0x61, 0x6e, 0x63, 0x68, 0x65, 0x73, 0x6e, 0x61, 0x6d, 0x65,
	0x12, 0x12, 0x0a, 0x04, 0x6b, 0x65, 0x79, 0x73, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x6b, 0x65, 0x79, 0x73, 0x12, 0x20, 0x0a, 0x0b, 0x6c, 0x65, 0x67, 0x61, 0x6c, 0x50, 0x65, 0x72,
	0x73, 0x6f, 0x6e, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x6c, 0x65, 0x67, 0x61, 0x6c,
	0x50, 0x65, 0x72, 0x73, 0x6f, 0x6e, 0x12, 0x2c, 0x0a, 0x11, 0x6c, 0x65, 0x67, 0x61, 0x6c, 0x49,
	0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79, 0x43, 0x61, 0x72, 0x64, 0x18, 0x0e, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x11, 0x6c, 0x65, 0x67, 0x61, 0x6c, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x74, 0x79,
	0x43, 0x61, 0x72, 0x64, 0x12, 0x2e, 0x0a, 0x12, 0x62, 0x75, 0x73, 0x69, 0x6e, 0x65, 0x73, 0x73,
	0x4c, 0x69, 0x63, 0x65, 0x6e, 0x73, 0x65, 0x55, 0x72, 0x6c, 0x18, 0x0f, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x12, 0x62, 0x75, 0x73, 0x69, 0x6e, 0x65, 0x73, 0x73, 0x4c, 0x69, 0x63, 0x65, 0x6e, 0x73,
	0x65, 0x55, 0x72, 0x6c, 0x12, 0x16, 0x0a, 0x06, 0x77, 0x65, 0x63, 0x68, 0x61, 0x74, 0x18, 0x10,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x77, 0x65, 0x63, 0x68, 0x61, 0x74, 0x12, 0x1c, 0x0a, 0x09,
	0x6c, 0x6f, 0x6e, 0x67, 0x69, 0x74, 0x75, 0x64, 0x65, 0x18, 0x11, 0x20, 0x01, 0x28, 0x01, 0x52,
	0x09, 0x6c, 0x6f, 0x6e, 0x67, 0x69, 0x74, 0x75, 0x64, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x6c, 0x61,
	0x74, 0x69, 0x74, 0x75, 0x64, 0x65, 0x18, 0x12, 0x20, 0x01, 0x28, 0x01, 0x52, 0x08, 0x6c, 0x61,
	0x74, 0x69, 0x74, 0x75, 0x64, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65,
	0x64, 0x41, 0x74, 0x18, 0x13, 0x20, 0x01, 0x28, 0x06, 0x52, 0x09, 0x63, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x64, 0x41, 0x74, 0x12, 0x1c, 0x0a, 0x09, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x41,
	0x74, 0x18, 0x14, 0x20, 0x01, 0x28, 0x06, 0x52, 0x09, 0x75, 0x70, 0x64, 0x61, 0x74, 0x65, 0x64,
	0x41, 0x74, 0x22, 0x7a, 0x0a, 0x1c, 0x42, 0x75, 0x73, 0x69, 0x6e, 0x65, 0x73, 0x73, 0x55, 0x73,
	0x65, 0x72, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x4c, 0x69, 0x63, 0x65, 0x6e, 0x73, 0x65, 0x52,
	0x65, 0x71, 0x12, 0x2a, 0x0a, 0x10, 0x62, 0x75, 0x73, 0x69, 0x6e, 0x65, 0x73, 0x73, 0x75, 0x73,
	0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x10, 0x62, 0x75,
	0x73, 0x69, 0x6e, 0x65, 0x73, 0x73, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x2e,
	0x0a, 0x12, 0x62, 0x75, 0x73, 0x69, 0x6e, 0x65, 0x73, 0x73, 0x4c, 0x69, 0x63, 0x65, 0x6e, 0x73,
	0x65, 0x55, 0x72, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x12, 0x62, 0x75, 0x73, 0x69,
	0x6e, 0x65, 0x73, 0x73, 0x4c, 0x69, 0x63, 0x65, 0x6e, 0x73, 0x65, 0x55, 0x72, 0x6c, 0x22, 0xc7,
	0x02, 0x0a, 0x14, 0x51, 0x75, 0x65, 0x72, 0x79, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x73, 0x4e, 0x65,
	0x61, 0x72, 0x62, 0x79, 0x52, 0x65, 0x71, 0x12, 0x3f, 0x0a, 0x09, 0x73, 0x74, 0x6f, 0x72, 0x65,
	0x54, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x21, 0x2e, 0x63, 0x6c, 0x6f,
	0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x67, 0x6c, 0x6f,
	0x62, 0x61, 0x6c, 0x2e, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x54, 0x79, 0x70, 0x65, 0x52, 0x09, 0x73,
	0x74, 0x6f, 0x72, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6b, 0x65, 0x79, 0x73,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6b, 0x65, 0x79, 0x73, 0x12, 0x1c, 0x0a, 0x09,
	0x6c, 0x6f, 0x6e, 0x67, 0x69, 0x74, 0x75, 0x64, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x01, 0x52,
	0x09, 0x6c, 0x6f, 0x6e, 0x67, 0x69, 0x74, 0x75, 0x64, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x6c, 0x61,
	0x74, 0x69, 0x74, 0x75, 0x64, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x01, 0x52, 0x08, 0x6c, 0x61,
	0x74, 0x69, 0x74, 0x75, 0x64, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x72, 0x61, 0x64, 0x69, 0x75, 0x73,
	0x18, 0x05, 0x20, 0x01, 0x28, 0x01, 0x52, 0x06, 0x72, 0x61, 0x64, 0x69, 0x75, 0x73, 0x12, 0x1a,
	0x0a, 0x08, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x6e, 0x63, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x08, 0x70, 0x72, 0x6f, 0x76, 0x69, 0x6e, 0x63, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x63, 0x69,
	0x74, 0x79, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x63, 0x69, 0x74, 0x79, 0x12, 0x16,
	0x0a, 0x06, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x79, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06,
	0x63, 0x6f, 0x75, 0x6e, 0x74, 0x79, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x72, 0x65, 0x65, 0x74,
	0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x74, 0x72, 0x65, 0x65, 0x74, 0x12, 0x12,
	0x0a, 0x04, 0x70, 0x61, 0x67, 0x65, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x70, 0x61,
	0x67, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x18, 0x0b, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x05, 0x6c, 0x69, 0x6d, 0x69, 0x74, 0x22, 0x6a, 0x0a, 0x15, 0x51, 0x75, 0x65, 0x72,
	0x79, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x73, 0x4e, 0x65, 0x61, 0x72, 0x62, 0x79, 0x52, 0x65, 0x73,
	0x70, 0x12, 0x1c, 0x0a, 0x09, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x50, 0x61, 0x67, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x06, 0x52, 0x09, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x50, 0x61, 0x67, 0x65, 0x12,
	0x33, 0x0a, 0x06, 0x73, 0x74, 0x6f, 0x72, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x1b, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69,
	0x6d, 0x2e, 0x75, 0x73, 0x65, 0x72, 0x2e, 0x53, 0x74, 0x6f, 0x72, 0x65, 0x52, 0x06, 0x73, 0x74,
	0x6f, 0x72, 0x65, 0x73, 0x42, 0x2a, 0x5a, 0x28, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x75, 0x73, 0x65, 0x72,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
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
	(*QueryStoresNearbyReq)(nil),         // 2: cloud.lianmi.im.user.QueryStoresNearbyReq
	(*QueryStoresNearbyResp)(nil),        // 3: cloud.lianmi.im.user.QueryStoresNearbyResp
	(global.StoreType)(0),                // 4: cloud.lianmi.im.global.StoreType
}
var file_api_proto_user_store_proto_depIdxs = []int32{
	4, // 0: cloud.lianmi.im.user.Store.storeType:type_name -> cloud.lianmi.im.global.StoreType
	4, // 1: cloud.lianmi.im.user.QueryStoresNearbyReq.storeType:type_name -> cloud.lianmi.im.global.StoreType
	0, // 2: cloud.lianmi.im.user.QueryStoresNearbyResp.stores:type_name -> cloud.lianmi.im.user.Store
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
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
			switch v := v.(*QueryStoresNearbyReq); i {
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
			switch v := v.(*QueryStoresNearbyResp); i {
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
