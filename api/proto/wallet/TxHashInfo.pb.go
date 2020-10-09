// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.7.0
// source: api/proto/wallet/TxHashInfo.proto

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

//查询交易
type TxHashInfoReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//交易哈希
	TxHash string `protobuf:"bytes,1,opt,name=txHash,proto3" json:"txHash,omitempty"`
}

func (x *TxHashInfoReq) Reset() {
	*x = TxHashInfoReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_wallet_TxHashInfo_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TxHashInfoReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TxHashInfoReq) ProtoMessage() {}

func (x *TxHashInfoReq) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_wallet_TxHashInfo_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TxHashInfoReq.ProtoReflect.Descriptor instead.
func (*TxHashInfoReq) Descriptor() ([]byte, []int) {
	return file_api_proto_wallet_TxHashInfo_proto_rawDescGZIP(), []int{0}
}

func (x *TxHashInfoReq) GetTxHash() string {
	if x != nil {
		return x.TxHash
	}
	return ""
}

type TxHashInfoRsp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//区块高度
	BlockNumber uint64 `protobuf:"fixed64,1,opt,name=blockNumber,proto3" json:"blockNumber,omitempty"`
	//区块打包时间
	BlockTime uint64 `protobuf:"fixed64,2,opt,name=blockTime,proto3" json:"blockTime,omitempty"`
	//ether
	Value uint64 `protobuf:"fixed64,3,opt,name=value,proto3" json:"value,omitempty"`
	//燃气值
	Gas uint64 `protobuf:"fixed64,4,opt,name=gas,proto3" json:"gas,omitempty"`
	//燃气价格
	GasPrice uint64 `protobuf:"fixed64,5,opt,name=gasPrice,proto3" json:"gasPrice,omitempty"`
	//随机数
	Nonce uint64 `protobuf:"fixed64,6,opt,name=nonce,proto3" json:"nonce,omitempty"`
	//数据，hex格式
	Input string `protobuf:"bytes,7,opt,name=input,proto3" json:"input,omitempty"`
	//发送者账号
	From string `protobuf:"bytes,8,opt,name=from,proto3" json:"from,omitempty"`
	//接收者账号
	To string `protobuf:"bytes,9,opt,name=to,proto3" json:"to,omitempty"`
}

func (x *TxHashInfoRsp) Reset() {
	*x = TxHashInfoRsp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_wallet_TxHashInfo_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TxHashInfoRsp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TxHashInfoRsp) ProtoMessage() {}

func (x *TxHashInfoRsp) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_wallet_TxHashInfo_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TxHashInfoRsp.ProtoReflect.Descriptor instead.
func (*TxHashInfoRsp) Descriptor() ([]byte, []int) {
	return file_api_proto_wallet_TxHashInfo_proto_rawDescGZIP(), []int{1}
}

func (x *TxHashInfoRsp) GetBlockNumber() uint64 {
	if x != nil {
		return x.BlockNumber
	}
	return 0
}

func (x *TxHashInfoRsp) GetBlockTime() uint64 {
	if x != nil {
		return x.BlockTime
	}
	return 0
}

func (x *TxHashInfoRsp) GetValue() uint64 {
	if x != nil {
		return x.Value
	}
	return 0
}

func (x *TxHashInfoRsp) GetGas() uint64 {
	if x != nil {
		return x.Gas
	}
	return 0
}

func (x *TxHashInfoRsp) GetGasPrice() uint64 {
	if x != nil {
		return x.GasPrice
	}
	return 0
}

func (x *TxHashInfoRsp) GetNonce() uint64 {
	if x != nil {
		return x.Nonce
	}
	return 0
}

func (x *TxHashInfoRsp) GetInput() string {
	if x != nil {
		return x.Input
	}
	return ""
}

func (x *TxHashInfoRsp) GetFrom() string {
	if x != nil {
		return x.From
	}
	return ""
}

func (x *TxHashInfoRsp) GetTo() string {
	if x != nil {
		return x.To
	}
	return ""
}

var File_api_proto_wallet_TxHashInfo_proto protoreflect.FileDescriptor

var file_api_proto_wallet_TxHashInfo_proto_rawDesc = []byte{
	0x0a, 0x21, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x77, 0x61, 0x6c, 0x6c,
	0x65, 0x74, 0x2f, 0x54, 0x78, 0x48, 0x61, 0x73, 0x68, 0x49, 0x6e, 0x66, 0x6f, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x16, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d,
	0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x77, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x22, 0x27, 0x0a, 0x0d, 0x54,
	0x78, 0x48, 0x61, 0x73, 0x68, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x71, 0x12, 0x16, 0x0a, 0x06,
	0x74, 0x78, 0x48, 0x61, 0x73, 0x68, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x74, 0x78,
	0x48, 0x61, 0x73, 0x68, 0x22, 0xe3, 0x01, 0x0a, 0x0d, 0x54, 0x78, 0x48, 0x61, 0x73, 0x68, 0x49,
	0x6e, 0x66, 0x6f, 0x52, 0x73, 0x70, 0x12, 0x20, 0x0a, 0x0b, 0x62, 0x6c, 0x6f, 0x63, 0x6b, 0x4e,
	0x75, 0x6d, 0x62, 0x65, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x06, 0x52, 0x0b, 0x62, 0x6c, 0x6f,
	0x63, 0x6b, 0x4e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x12, 0x1c, 0x0a, 0x09, 0x62, 0x6c, 0x6f, 0x63,
	0x6b, 0x54, 0x69, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x06, 0x52, 0x09, 0x62, 0x6c, 0x6f,
	0x63, 0x6b, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x06, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x10, 0x0a, 0x03,
	0x67, 0x61, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x06, 0x52, 0x03, 0x67, 0x61, 0x73, 0x12, 0x1a,
	0x0a, 0x08, 0x67, 0x61, 0x73, 0x50, 0x72, 0x69, 0x63, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x06,
	0x52, 0x08, 0x67, 0x61, 0x73, 0x50, 0x72, 0x69, 0x63, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x6e, 0x6f,
	0x6e, 0x63, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x06, 0x52, 0x05, 0x6e, 0x6f, 0x6e, 0x63, 0x65,
	0x12, 0x14, 0x0a, 0x05, 0x69, 0x6e, 0x70, 0x75, 0x74, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x05, 0x69, 0x6e, 0x70, 0x75, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x66, 0x72, 0x6f, 0x6d, 0x18, 0x08,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x66, 0x72, 0x6f, 0x6d, 0x12, 0x0e, 0x0a, 0x02, 0x74, 0x6f,
	0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x74, 0x6f, 0x42, 0x2c, 0x5a, 0x2a, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2f,
	0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x2f, 0x77, 0x61, 0x6c, 0x6c, 0x65, 0x74, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_proto_wallet_TxHashInfo_proto_rawDescOnce sync.Once
	file_api_proto_wallet_TxHashInfo_proto_rawDescData = file_api_proto_wallet_TxHashInfo_proto_rawDesc
)

func file_api_proto_wallet_TxHashInfo_proto_rawDescGZIP() []byte {
	file_api_proto_wallet_TxHashInfo_proto_rawDescOnce.Do(func() {
		file_api_proto_wallet_TxHashInfo_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_proto_wallet_TxHashInfo_proto_rawDescData)
	})
	return file_api_proto_wallet_TxHashInfo_proto_rawDescData
}

var file_api_proto_wallet_TxHashInfo_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_api_proto_wallet_TxHashInfo_proto_goTypes = []interface{}{
	(*TxHashInfoReq)(nil), // 0: cloud.lianmi.im.wallet.TxHashInfoReq
	(*TxHashInfoRsp)(nil), // 1: cloud.lianmi.im.wallet.TxHashInfoRsp
}
var file_api_proto_wallet_TxHashInfo_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_api_proto_wallet_TxHashInfo_proto_init() }
func file_api_proto_wallet_TxHashInfo_proto_init() {
	if File_api_proto_wallet_TxHashInfo_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_proto_wallet_TxHashInfo_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TxHashInfoReq); i {
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
		file_api_proto_wallet_TxHashInfo_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TxHashInfoRsp); i {
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
			RawDescriptor: file_api_proto_wallet_TxHashInfo_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_proto_wallet_TxHashInfo_proto_goTypes,
		DependencyIndexes: file_api_proto_wallet_TxHashInfo_proto_depIdxs,
		MessageInfos:      file_api_proto_wallet_TxHashInfo_proto_msgTypes,
	}.Build()
	File_api_proto_wallet_TxHashInfo_proto = out.File
	file_api_proto_wallet_TxHashInfo_proto_rawDesc = nil
	file_api_proto_wallet_TxHashInfo_proto_goTypes = nil
	file_api_proto_wallet_TxHashInfo_proto_depIdxs = nil
}