// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.7.0
// source: api/proto/wallet/QueryCommission.proto

package wallet

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

type QueryCommissionReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//用户账号
	UserName string `protobuf:"bytes,1,opt,name=userName,proto3" json:"userName,omitempty"`
}

func (x *QueryCommissionReq) Reset() {
	*x = QueryCommissionReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_wallet_QueryCommission_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QueryCommissionReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryCommissionReq) ProtoMessage() {}

func (x *QueryCommissionReq) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_wallet_QueryCommission_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QueryCommissionReq.ProtoReflect.Descriptor instead.
func (*QueryCommissionReq) Descriptor() ([]byte, []int) {
	return file_api_proto_wallet_QueryCommission_proto_rawDescGZIP(), []int{0}
}

func (x *QueryCommissionReq) GetUserName() string {
	if x != nil {
		return x.UserName
	}
	return ""
}

type Commission struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//年月 2020-08
	YearMonth string `protobuf:"bytes,1,opt,name=yearMonth,proto3" json:"yearMonth,omitempty"`
	//当月佣金总额
	Commission float64 `protobuf:"fixed64,2,opt,name=commission,proto3" json:"commission,omitempty"`
}

func (x *Commission) Reset() {
	*x = Commission{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_wallet_QueryCommission_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Commission) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Commission) ProtoMessage() {}

func (x *Commission) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_wallet_QueryCommission_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Commission.ProtoReflect.Descriptor instead.
func (*Commission) Descriptor() ([]byte, []int) {
	return file_api_proto_wallet_QueryCommission_proto_rawDescGZIP(), []int{1}
}

func (x *Commission) GetYearMonth() string {
	if x != nil {
		return x.YearMonth
	}
	return ""
}

func (x *Commission) GetCommission() float64 {
	if x != nil {
		return x.Commission
	}
	return 0
}

type QueryCommissionRsp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//是否成功
	Commissions []*Commission `protobuf:"bytes,1,rep,name=commissions,proto3" json:"commissions,omitempty"`
}

func (x *QueryCommissionRsp) Reset() {
	*x = QueryCommissionRsp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_wallet_QueryCommission_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QueryCommissionRsp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryCommissionRsp) ProtoMessage() {}

func (x *QueryCommissionRsp) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_wallet_QueryCommission_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QueryCommissionRsp.ProtoReflect.Descriptor instead.
func (*QueryCommissionRsp) Descriptor() ([]byte, []int) {
	return file_api_proto_wallet_QueryCommission_proto_rawDescGZIP(), []int{2}
}

func (x *QueryCommissionRsp) GetCommissions() []*Commission {
	if x != nil {
		return x.Commissions
	}
	return nil
}

var File_api_proto_wallet_QueryCommission_proto protoreflect.FileDescriptor

var file_api_proto_wallet_QueryCommission_proto_rawDesc = []byte{
	0x0a, 0x26, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x77, 0x61, 0x6c, 0x6c,
	0x65, 0x74, 0x2f, 0x51, 0x75, 0x65, 0x72, 0x79, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x73, 0x73, 0x69,
	0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x16, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e,
	0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x77, 0x61, 0x6c, 0x6c, 0x65, 0x74,
	0x22, 0x30, 0x0a, 0x12, 0x51, 0x75, 0x65, 0x72, 0x79, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x73, 0x73,
	0x69, 0x6f, 0x6e, 0x52, 0x65, 0x71, 0x12, 0x1a, 0x0a, 0x08, 0x75, 0x73, 0x65, 0x72, 0x4e, 0x61,
	0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x75, 0x73, 0x65, 0x72, 0x4e, 0x61,
	0x6d, 0x65, 0x22, 0x4a, 0x0a, 0x0a, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e,
	0x12, 0x1c, 0x0a, 0x09, 0x79, 0x65, 0x61, 0x72, 0x4d, 0x6f, 0x6e, 0x74, 0x68, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x09, 0x79, 0x65, 0x61, 0x72, 0x4d, 0x6f, 0x6e, 0x74, 0x68, 0x12, 0x1e,
	0x0a, 0x0a, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x01, 0x52, 0x0a, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x22, 0x5a,
	0x0a, 0x12, 0x51, 0x75, 0x65, 0x72, 0x79, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f,
	0x6e, 0x52, 0x73, 0x70, 0x12, 0x44, 0x0a, 0x0b, 0x63, 0x6f, 0x6d, 0x6d, 0x69, 0x73, 0x73, 0x69,
	0x6f, 0x6e, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x63, 0x6c, 0x6f, 0x75,
	0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x77, 0x61, 0x6c, 0x6c,
	0x65, 0x74, 0x2e, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x0b, 0x63,
	0x6f, 0x6d, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x42, 0x2c, 0x5a, 0x2a, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2f,
	0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x2f, 0x77, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_proto_wallet_QueryCommission_proto_rawDescOnce sync.Once
	file_api_proto_wallet_QueryCommission_proto_rawDescData = file_api_proto_wallet_QueryCommission_proto_rawDesc
)

func file_api_proto_wallet_QueryCommission_proto_rawDescGZIP() []byte {
	file_api_proto_wallet_QueryCommission_proto_rawDescOnce.Do(func() {
		file_api_proto_wallet_QueryCommission_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_proto_wallet_QueryCommission_proto_rawDescData)
	})
	return file_api_proto_wallet_QueryCommission_proto_rawDescData
}

var file_api_proto_wallet_QueryCommission_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_api_proto_wallet_QueryCommission_proto_goTypes = []interface{}{
	(*QueryCommissionReq)(nil), // 0: cloud.lianmi.im.wallet.QueryCommissionReq
	(*Commission)(nil),         // 1: cloud.lianmi.im.wallet.Commission
	(*QueryCommissionRsp)(nil), // 2: cloud.lianmi.im.wallet.QueryCommissionRsp
}
var file_api_proto_wallet_QueryCommission_proto_depIdxs = []int32{
	1, // 0: cloud.lianmi.im.wallet.QueryCommissionRsp.commissions:type_name -> cloud.lianmi.im.wallet.Commission
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_api_proto_wallet_QueryCommission_proto_init() }
func file_api_proto_wallet_QueryCommission_proto_init() {
	if File_api_proto_wallet_QueryCommission_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_proto_wallet_QueryCommission_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*QueryCommissionReq); i {
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
		file_api_proto_wallet_QueryCommission_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Commission); i {
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
		file_api_proto_wallet_QueryCommission_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*QueryCommissionRsp); i {
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
			RawDescriptor: file_api_proto_wallet_QueryCommission_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_proto_wallet_QueryCommission_proto_goTypes,
		DependencyIndexes: file_api_proto_wallet_QueryCommission_proto_depIdxs,
		MessageInfos:      file_api_proto_wallet_QueryCommission_proto_msgTypes,
	}.Build()
	File_api_proto_wallet_QueryCommission_proto = out.File
	file_api_proto_wallet_QueryCommission_proto_rawDesc = nil
	file_api_proto_wallet_QueryCommission_proto_goTypes = nil
	file_api_proto_wallet_QueryCommission_proto_depIdxs = nil
}