// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.7.0
// source: api/proto/order/OrderPayDoneEvent.proto

package order

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

type OrderPayDoneEventRsp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//订单ID
	OrderID string `protobuf:"bytes,1,opt,name=orderID,proto3" json:"orderID,omitempty"`
	//多签合约地址
	ContractAddress string `protobuf:"bytes,2,opt,name=contractAddress,proto3" json:"contractAddress,omitempty"`
	//本次支付连米币数量
	Amount uint64 `protobuf:"fixed64,3,opt,name=amount,proto3" json:"amount,omitempty"`
	// 区块高度
	BlockNumber uint64 `protobuf:"fixed64,4,opt,name=blockNumber,proto3" json:"blockNumber,omitempty"`
	// 交易哈希hex
	Hash string `protobuf:"bytes,5,opt,name=hash,proto3" json:"hash,omitempty"`
	//时间
	Time uint64 `protobuf:"fixed64,6,opt,name=time,proto3" json:"time,omitempty"`
}

func (x *OrderPayDoneEventRsp) Reset() {
	*x = OrderPayDoneEventRsp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_order_OrderPayDoneEvent_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *OrderPayDoneEventRsp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OrderPayDoneEventRsp) ProtoMessage() {}

func (x *OrderPayDoneEventRsp) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_order_OrderPayDoneEvent_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OrderPayDoneEventRsp.ProtoReflect.Descriptor instead.
func (*OrderPayDoneEventRsp) Descriptor() ([]byte, []int) {
	return file_api_proto_order_OrderPayDoneEvent_proto_rawDescGZIP(), []int{0}
}

func (x *OrderPayDoneEventRsp) GetOrderID() string {
	if x != nil {
		return x.OrderID
	}
	return ""
}

func (x *OrderPayDoneEventRsp) GetContractAddress() string {
	if x != nil {
		return x.ContractAddress
	}
	return ""
}

func (x *OrderPayDoneEventRsp) GetAmount() uint64 {
	if x != nil {
		return x.Amount
	}
	return 0
}

func (x *OrderPayDoneEventRsp) GetBlockNumber() uint64 {
	if x != nil {
		return x.BlockNumber
	}
	return 0
}

func (x *OrderPayDoneEventRsp) GetHash() string {
	if x != nil {
		return x.Hash
	}
	return ""
}

func (x *OrderPayDoneEventRsp) GetTime() uint64 {
	if x != nil {
		return x.Time
	}
	return 0
}

var File_api_proto_order_OrderPayDoneEvent_proto protoreflect.FileDescriptor

var file_api_proto_order_OrderPayDoneEvent_proto_rawDesc = []byte{
	0x0a, 0x27, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6f, 0x72, 0x64, 0x65,
	0x72, 0x2f, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x50, 0x61, 0x79, 0x44, 0x6f, 0x6e, 0x65, 0x45, 0x76,
	0x65, 0x6e, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x15, 0x63, 0x6c, 0x6f, 0x75, 0x64,
	0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x6f, 0x72, 0x64, 0x65, 0x72,
	0x22, 0xbc, 0x01, 0x0a, 0x14, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x50, 0x61, 0x79, 0x44, 0x6f, 0x6e,
	0x65, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x52, 0x73, 0x70, 0x12, 0x18, 0x0a, 0x07, 0x6f, 0x72, 0x64,
	0x65, 0x72, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6f, 0x72, 0x64, 0x65,
	0x72, 0x49, 0x44, 0x12, 0x28, 0x0a, 0x0f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x41,
	0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x63, 0x6f,
	0x6e, 0x74, 0x72, 0x61, 0x63, 0x74, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x16, 0x0a,
	0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x06, 0x52, 0x06, 0x61,
	0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x20, 0x0a, 0x0b, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x4e, 0x75,
	0x6d, 0x62, 0x65, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x06, 0x52, 0x0b, 0x62, 0x6c, 0x6f, 0x63,
	0x6b, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x61, 0x73, 0x68, 0x18,
	0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x68, 0x61, 0x73, 0x68, 0x12, 0x12, 0x0a, 0x04, 0x74,
	0x69, 0x6d, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x06, 0x52, 0x04, 0x74, 0x69, 0x6d, 0x65, 0x42,
	0x2b, 0x5a, 0x29, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x69,
	0x61, 0x6e, 0x6d, 0x69, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x2f, 0x61, 0x70, 0x69,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_proto_order_OrderPayDoneEvent_proto_rawDescOnce sync.Once
	file_api_proto_order_OrderPayDoneEvent_proto_rawDescData = file_api_proto_order_OrderPayDoneEvent_proto_rawDesc
)

func file_api_proto_order_OrderPayDoneEvent_proto_rawDescGZIP() []byte {
	file_api_proto_order_OrderPayDoneEvent_proto_rawDescOnce.Do(func() {
		file_api_proto_order_OrderPayDoneEvent_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_proto_order_OrderPayDoneEvent_proto_rawDescData)
	})
	return file_api_proto_order_OrderPayDoneEvent_proto_rawDescData
}

var file_api_proto_order_OrderPayDoneEvent_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_api_proto_order_OrderPayDoneEvent_proto_goTypes = []interface{}{
	(*OrderPayDoneEventRsp)(nil), // 0: cloud.lianmi.im.order.OrderPayDoneEventRsp
}
var file_api_proto_order_OrderPayDoneEvent_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_api_proto_order_OrderPayDoneEvent_proto_init() }
func file_api_proto_order_OrderPayDoneEvent_proto_init() {
	if File_api_proto_order_OrderPayDoneEvent_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_proto_order_OrderPayDoneEvent_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*OrderPayDoneEventRsp); i {
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
			RawDescriptor: file_api_proto_order_OrderPayDoneEvent_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_proto_order_OrderPayDoneEvent_proto_goTypes,
		DependencyIndexes: file_api_proto_order_OrderPayDoneEvent_proto_depIdxs,
		MessageInfos:      file_api_proto_order_OrderPayDoneEvent_proto_msgTypes,
	}.Build()
	File_api_proto_order_OrderPayDoneEvent_proto = out.File
	file_api_proto_order_OrderPayDoneEvent_proto_rawDesc = nil
	file_api_proto_order_OrderPayDoneEvent_proto_goTypes = nil
	file_api_proto_order_OrderPayDoneEvent_proto_depIdxs = nil
}
