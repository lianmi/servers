// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.7.0
// source: api/proto/wallet/SyncWithdrawHistory.proto

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

type SyncWithdrawHistoryPageReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//开始时间
	StartAt uint64 `protobuf:"fixed64,1,opt,name=startAt,proto3" json:"startAt,omitempty"`
	//结束时间
	EndAt uint64 `protobuf:"fixed64,2,opt,name=endAt,proto3" json:"endAt,omitempty"`
	//页数,第几页
	Page int32 `protobuf:"varint,3,opt,name=page,proto3" json:"page,omitempty"`
	//每页记录数量
	PageSize int32 `protobuf:"varint,4,opt,name=pageSize,proto3" json:"pageSize,omitempty"`
}

func (x *SyncWithdrawHistoryPageReq) Reset() {
	*x = SyncWithdrawHistoryPageReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_wallet_SyncWithdrawHistory_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SyncWithdrawHistoryPageReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SyncWithdrawHistoryPageReq) ProtoMessage() {}

func (x *SyncWithdrawHistoryPageReq) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_wallet_SyncWithdrawHistory_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SyncWithdrawHistoryPageReq.ProtoReflect.Descriptor instead.
func (*SyncWithdrawHistoryPageReq) Descriptor() ([]byte, []int) {
	return file_api_proto_wallet_SyncWithdrawHistory_proto_rawDescGZIP(), []int{0}
}

func (x *SyncWithdrawHistoryPageReq) GetStartAt() uint64 {
	if x != nil {
		return x.StartAt
	}
	return 0
}

func (x *SyncWithdrawHistoryPageReq) GetEndAt() uint64 {
	if x != nil {
		return x.EndAt
	}
	return 0
}

func (x *SyncWithdrawHistoryPageReq) GetPage() int32 {
	if x != nil {
		return x.Page
	}
	return 0
}

func (x *SyncWithdrawHistoryPageReq) GetPageSize() int32 {
	if x != nil {
		return x.PageSize
	}
	return 0
}

type SyncWithdrawHistoryPageRsp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//提现记录列表
	Withdraws []*Withdraw `protobuf:"bytes,1,rep,name=withdraws,proto3" json:"withdraws,omitempty"`
	//按请求参数的pageSize计算出来的总页数
	Total int32 `protobuf:"varint,2,opt,name=total,proto3" json:"total,omitempty"`
}

func (x *SyncWithdrawHistoryPageRsp) Reset() {
	*x = SyncWithdrawHistoryPageRsp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_wallet_SyncWithdrawHistory_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SyncWithdrawHistoryPageRsp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SyncWithdrawHistoryPageRsp) ProtoMessage() {}

func (x *SyncWithdrawHistoryPageRsp) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_wallet_SyncWithdrawHistory_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SyncWithdrawHistoryPageRsp.ProtoReflect.Descriptor instead.
func (*SyncWithdrawHistoryPageRsp) Descriptor() ([]byte, []int) {
	return file_api_proto_wallet_SyncWithdrawHistory_proto_rawDescGZIP(), []int{1}
}

func (x *SyncWithdrawHistoryPageRsp) GetWithdraws() []*Withdraw {
	if x != nil {
		return x.Withdraws
	}
	return nil
}

func (x *SyncWithdrawHistoryPageRsp) GetTotal() int32 {
	if x != nil {
		return x.Total
	}
	return 0
}

var File_api_proto_wallet_SyncWithdrawHistory_proto protoreflect.FileDescriptor

var file_api_proto_wallet_SyncWithdrawHistory_proto_rawDesc = []byte{
	0x0a, 0x2a, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x77, 0x61, 0x6c, 0x6c,
	0x65, 0x74, 0x2f, 0x53, 0x79, 0x6e, 0x63, 0x57, 0x69, 0x74, 0x68, 0x64, 0x72, 0x61, 0x77, 0x48,
	0x69, 0x73, 0x74, 0x6f, 0x72, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x16, 0x63, 0x6c,
	0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x77, 0x61,
	0x6c, 0x6c, 0x65, 0x74, 0x1a, 0x1d, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f,
	0x77, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x2f, 0x57, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x22, 0x7c, 0x0a, 0x1a, 0x53, 0x79, 0x6e, 0x63, 0x57, 0x69, 0x74, 0x68, 0x64,
	0x72, 0x61, 0x77, 0x48, 0x69, 0x73, 0x74, 0x6f, 0x72, 0x79, 0x50, 0x61, 0x67, 0x65, 0x52, 0x65,
	0x71, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x74, 0x61, 0x72, 0x74, 0x41, 0x74, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x06, 0x52, 0x07, 0x73, 0x74, 0x61, 0x72, 0x74, 0x41, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x65,
	0x6e, 0x64, 0x41, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x06, 0x52, 0x05, 0x65, 0x6e, 0x64, 0x41,
	0x74, 0x12, 0x12, 0x0a, 0x04, 0x70, 0x61, 0x67, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x04, 0x70, 0x61, 0x67, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x70, 0x61, 0x67, 0x65, 0x53, 0x69, 0x7a,
	0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08, 0x70, 0x61, 0x67, 0x65, 0x53, 0x69, 0x7a,
	0x65, 0x22, 0x72, 0x0a, 0x1a, 0x53, 0x79, 0x6e, 0x63, 0x57, 0x69, 0x74, 0x68, 0x64, 0x72, 0x61,
	0x77, 0x48, 0x69, 0x73, 0x74, 0x6f, 0x72, 0x79, 0x50, 0x61, 0x67, 0x65, 0x52, 0x73, 0x70, 0x12,
	0x3e, 0x0a, 0x09, 0x77, 0x69, 0x74, 0x68, 0x64, 0x72, 0x61, 0x77, 0x73, 0x18, 0x01, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x20, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d,
	0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x77, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x2e, 0x57, 0x69, 0x74, 0x68,
	0x64, 0x72, 0x61, 0x77, 0x52, 0x09, 0x77, 0x69, 0x74, 0x68, 0x64, 0x72, 0x61, 0x77, 0x73, 0x12,
	0x14, 0x0a, 0x05, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05,
	0x74, 0x6f, 0x74, 0x61, 0x6c, 0x42, 0x2c, 0x5a, 0x2a, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e,
	0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x65,
	0x72, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x77, 0x61, 0x6c,
	0x6c, 0x65, 0x74, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_proto_wallet_SyncWithdrawHistory_proto_rawDescOnce sync.Once
	file_api_proto_wallet_SyncWithdrawHistory_proto_rawDescData = file_api_proto_wallet_SyncWithdrawHistory_proto_rawDesc
)

func file_api_proto_wallet_SyncWithdrawHistory_proto_rawDescGZIP() []byte {
	file_api_proto_wallet_SyncWithdrawHistory_proto_rawDescOnce.Do(func() {
		file_api_proto_wallet_SyncWithdrawHistory_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_proto_wallet_SyncWithdrawHistory_proto_rawDescData)
	})
	return file_api_proto_wallet_SyncWithdrawHistory_proto_rawDescData
}

var file_api_proto_wallet_SyncWithdrawHistory_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_api_proto_wallet_SyncWithdrawHistory_proto_goTypes = []interface{}{
	(*SyncWithdrawHistoryPageReq)(nil), // 0: cloud.lianmi.im.wallet.SyncWithdrawHistoryPageReq
	(*SyncWithdrawHistoryPageRsp)(nil), // 1: cloud.lianmi.im.wallet.SyncWithdrawHistoryPageRsp
	(*Withdraw)(nil),                   // 2: cloud.lianmi.im.wallet.Withdraw
}
var file_api_proto_wallet_SyncWithdrawHistory_proto_depIdxs = []int32{
	2, // 0: cloud.lianmi.im.wallet.SyncWithdrawHistoryPageRsp.withdraws:type_name -> cloud.lianmi.im.wallet.Withdraw
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_api_proto_wallet_SyncWithdrawHistory_proto_init() }
func file_api_proto_wallet_SyncWithdrawHistory_proto_init() {
	if File_api_proto_wallet_SyncWithdrawHistory_proto != nil {
		return
	}
	file_api_proto_wallet_Wallet_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_api_proto_wallet_SyncWithdrawHistory_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SyncWithdrawHistoryPageReq); i {
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
		file_api_proto_wallet_SyncWithdrawHistory_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SyncWithdrawHistoryPageRsp); i {
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
			RawDescriptor: file_api_proto_wallet_SyncWithdrawHistory_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_proto_wallet_SyncWithdrawHistory_proto_goTypes,
		DependencyIndexes: file_api_proto_wallet_SyncWithdrawHistory_proto_depIdxs,
		MessageInfos:      file_api_proto_wallet_SyncWithdrawHistory_proto_msgTypes,
	}.Build()
	File_api_proto_wallet_SyncWithdrawHistory_proto = out.File
	file_api_proto_wallet_SyncWithdrawHistory_proto_rawDesc = nil
	file_api_proto_wallet_SyncWithdrawHistory_proto_goTypes = nil
	file_api_proto_wallet_SyncWithdrawHistory_proto_depIdxs = nil
}