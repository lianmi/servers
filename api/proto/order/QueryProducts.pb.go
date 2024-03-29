// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.7.0
// source: api/proto/order/QueryProducts.proto

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

type QueryProductsReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//商户用户账号 id
	//是否必须-是
	UserName string `protobuf:"bytes,1,opt,name=userName,proto3" json:"userName,omitempty"`
	//商品详情最大修改时间戳，对应timeAt字段，为0时获取全量商品
	//是否必须-是
	TimeAt uint64 `protobuf:"fixed64,2,opt,name=timeAt,proto3" json:"timeAt,omitempty"`
}

func (x *QueryProductsReq) Reset() {
	*x = QueryProductsReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_order_QueryProducts_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QueryProductsReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryProductsReq) ProtoMessage() {}

func (x *QueryProductsReq) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_order_QueryProducts_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QueryProductsReq.ProtoReflect.Descriptor instead.
func (*QueryProductsReq) Descriptor() ([]byte, []int) {
	return file_api_proto_order_QueryProducts_proto_rawDescGZIP(), []int{0}
}

func (x *QueryProductsReq) GetUserName() string {
	if x != nil {
		return x.UserName
	}
	return ""
}

func (x *QueryProductsReq) GetTimeAt() uint64 {
	if x != nil {
		return x.TimeAt
	}
	return 0
}

//
//获取商品信息-响应
type QueryProductsRsp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//商品列表
	//是否必须-是
	Products []*Product `protobuf:"bytes,1,rep,name=products,proto3" json:"products,omitempty"`
	//该商品下架后的商品id列表
	//是否必须-否
	SoldoutProducts []string `protobuf:"bytes,2,rep,name=soldoutProducts,proto3" json:"soldoutProducts,omitempty"`
	//本次同步后，服务器时间
	//是否必须-是
	TimeAt uint64 `protobuf:"fixed64,3,opt,name=timeAt,proto3" json:"timeAt,omitempty"`
}

func (x *QueryProductsRsp) Reset() {
	*x = QueryProductsRsp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_order_QueryProducts_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *QueryProductsRsp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*QueryProductsRsp) ProtoMessage() {}

func (x *QueryProductsRsp) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_order_QueryProducts_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use QueryProductsRsp.ProtoReflect.Descriptor instead.
func (*QueryProductsRsp) Descriptor() ([]byte, []int) {
	return file_api_proto_order_QueryProducts_proto_rawDescGZIP(), []int{1}
}

func (x *QueryProductsRsp) GetProducts() []*Product {
	if x != nil {
		return x.Products
	}
	return nil
}

func (x *QueryProductsRsp) GetSoldoutProducts() []string {
	if x != nil {
		return x.SoldoutProducts
	}
	return nil
}

func (x *QueryProductsRsp) GetTimeAt() uint64 {
	if x != nil {
		return x.TimeAt
	}
	return 0
}

var File_api_proto_order_QueryProducts_proto protoreflect.FileDescriptor

var file_api_proto_order_QueryProducts_proto_rawDesc = []byte{
	0x0a, 0x23, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6f, 0x72, 0x64, 0x65,
	0x72, 0x2f, 0x51, 0x75, 0x65, 0x72, 0x79, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x15, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61,
	0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x1a, 0x1d, 0x61, 0x70,
	0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x2f, 0x50, 0x72,
	0x6f, 0x64, 0x75, 0x63, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x46, 0x0a, 0x10, 0x51,
	0x75, 0x65, 0x72, 0x79, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x73, 0x52, 0x65, 0x71, 0x12,
	0x1a, 0x0a, 0x08, 0x75, 0x73, 0x65, 0x72, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x75, 0x73, 0x65, 0x72, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x74,
	0x69, 0x6d, 0x65, 0x41, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x06, 0x52, 0x06, 0x74, 0x69, 0x6d,
	0x65, 0x41, 0x74, 0x22, 0x90, 0x01, 0x0a, 0x10, 0x51, 0x75, 0x65, 0x72, 0x79, 0x50, 0x72, 0x6f,
	0x64, 0x75, 0x63, 0x74, 0x73, 0x52, 0x73, 0x70, 0x12, 0x3a, 0x0a, 0x08, 0x70, 0x72, 0x6f, 0x64,
	0x75, 0x63, 0x74, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x63, 0x6c, 0x6f,
	0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x6f, 0x72, 0x64,
	0x65, 0x72, 0x2e, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x52, 0x08, 0x70, 0x72, 0x6f, 0x64,
	0x75, 0x63, 0x74, 0x73, 0x12, 0x28, 0x0a, 0x0f, 0x73, 0x6f, 0x6c, 0x64, 0x6f, 0x75, 0x74, 0x50,
	0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0f, 0x73,
	0x6f, 0x6c, 0x64, 0x6f, 0x75, 0x74, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x73, 0x12, 0x16,
	0x0a, 0x06, 0x74, 0x69, 0x6d, 0x65, 0x41, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x06, 0x52, 0x06,
	0x74, 0x69, 0x6d, 0x65, 0x41, 0x74, 0x42, 0x2b, 0x5a, 0x29, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2f, 0x73, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6f, 0x72,
	0x64, 0x65, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_proto_order_QueryProducts_proto_rawDescOnce sync.Once
	file_api_proto_order_QueryProducts_proto_rawDescData = file_api_proto_order_QueryProducts_proto_rawDesc
)

func file_api_proto_order_QueryProducts_proto_rawDescGZIP() []byte {
	file_api_proto_order_QueryProducts_proto_rawDescOnce.Do(func() {
		file_api_proto_order_QueryProducts_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_proto_order_QueryProducts_proto_rawDescData)
	})
	return file_api_proto_order_QueryProducts_proto_rawDescData
}

var file_api_proto_order_QueryProducts_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_api_proto_order_QueryProducts_proto_goTypes = []interface{}{
	(*QueryProductsReq)(nil), // 0: cloud.lianmi.im.order.QueryProductsReq
	(*QueryProductsRsp)(nil), // 1: cloud.lianmi.im.order.QueryProductsRsp
	(*Product)(nil),          // 2: cloud.lianmi.im.order.Product
}
var file_api_proto_order_QueryProducts_proto_depIdxs = []int32{
	2, // 0: cloud.lianmi.im.order.QueryProductsRsp.products:type_name -> cloud.lianmi.im.order.Product
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_api_proto_order_QueryProducts_proto_init() }
func file_api_proto_order_QueryProducts_proto_init() {
	if File_api_proto_order_QueryProducts_proto != nil {
		return
	}
	file_api_proto_order_Product_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_api_proto_order_QueryProducts_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*QueryProductsReq); i {
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
		file_api_proto_order_QueryProducts_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*QueryProductsRsp); i {
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
			RawDescriptor: file_api_proto_order_QueryProducts_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_proto_order_QueryProducts_proto_goTypes,
		DependencyIndexes: file_api_proto_order_QueryProducts_proto_depIdxs,
		MessageInfos:      file_api_proto_order_QueryProducts_proto_msgTypes,
	}.Build()
	File_api_proto_order_QueryProducts_proto = out.File
	file_api_proto_order_QueryProducts_proto_rawDesc = nil
	file_api_proto_order_QueryProducts_proto_goTypes = nil
	file_api_proto_order_QueryProducts_proto_depIdxs = nil
}
