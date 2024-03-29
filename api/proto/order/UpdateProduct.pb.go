// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.7.0
// source: api/proto/order/UpdateProduct.proto

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

//请求参数
type UpdateProductReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//商品详情
	Product *Product `protobuf:"bytes,1,opt,name=product,proto3" json:"product,omitempty"`
}

func (x *UpdateProductReq) Reset() {
	*x = UpdateProductReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_order_UpdateProduct_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateProductReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateProductReq) ProtoMessage() {}

func (x *UpdateProductReq) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_order_UpdateProduct_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateProductReq.ProtoReflect.Descriptor instead.
func (*UpdateProductReq) Descriptor() ([]byte, []int) {
	return file_api_proto_order_UpdateProduct_proto_rawDescGZIP(), []int{0}
}

func (x *UpdateProductReq) GetProduct() *Product {
	if x != nil {
		return x.Product
	}
	return nil
}

//响应参数
type UpdateProductRsp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//商品详情
	Product *Product `protobuf:"bytes,1,opt,name=product,proto3" json:"product,omitempty"`
	//更新时间
	TimeAt uint64 `protobuf:"fixed64,2,opt,name=timeAt,proto3" json:"timeAt,omitempty"`
}

func (x *UpdateProductRsp) Reset() {
	*x = UpdateProductRsp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_order_UpdateProduct_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UpdateProductRsp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UpdateProductRsp) ProtoMessage() {}

func (x *UpdateProductRsp) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_order_UpdateProduct_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UpdateProductRsp.ProtoReflect.Descriptor instead.
func (*UpdateProductRsp) Descriptor() ([]byte, []int) {
	return file_api_proto_order_UpdateProduct_proto_rawDescGZIP(), []int{1}
}

func (x *UpdateProductRsp) GetProduct() *Product {
	if x != nil {
		return x.Product
	}
	return nil
}

func (x *UpdateProductRsp) GetTimeAt() uint64 {
	if x != nil {
		return x.TimeAt
	}
	return 0
}

var File_api_proto_order_UpdateProduct_proto protoreflect.FileDescriptor

var file_api_proto_order_UpdateProduct_proto_rawDesc = []byte{
	0x0a, 0x23, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6f, 0x72, 0x64, 0x65,
	0x72, 0x2f, 0x55, 0x70, 0x64, 0x61, 0x74, 0x65, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x15, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61,
	0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x1a, 0x1d, 0x61, 0x70,
	0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x2f, 0x50, 0x72,
	0x6f, 0x64, 0x75, 0x63, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x4c, 0x0a, 0x10, 0x55,
	0x70, 0x64, 0x61, 0x74, 0x65, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x52, 0x65, 0x71, 0x12,
	0x38, 0x0a, 0x07, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x1e, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e,
	0x69, 0x6d, 0x2e, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x2e, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74,
	0x52, 0x07, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x22, 0x64, 0x0a, 0x10, 0x55, 0x70, 0x64,
	0x61, 0x74, 0x65, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x52, 0x73, 0x70, 0x12, 0x38, 0x0a,
	0x07, 0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1e,
	0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d,
	0x2e, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x2e, 0x50, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x52, 0x07,
	0x70, 0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x12, 0x16, 0x0a, 0x06, 0x74, 0x69, 0x6d, 0x65, 0x41,
	0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x06, 0x52, 0x06, 0x74, 0x69, 0x6d, 0x65, 0x41, 0x74, 0x42,
	0x2b, 0x5a, 0x29, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x69,
	0x61, 0x6e, 0x6d, 0x69, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x2f, 0x61, 0x70, 0x69,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_proto_order_UpdateProduct_proto_rawDescOnce sync.Once
	file_api_proto_order_UpdateProduct_proto_rawDescData = file_api_proto_order_UpdateProduct_proto_rawDesc
)

func file_api_proto_order_UpdateProduct_proto_rawDescGZIP() []byte {
	file_api_proto_order_UpdateProduct_proto_rawDescOnce.Do(func() {
		file_api_proto_order_UpdateProduct_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_proto_order_UpdateProduct_proto_rawDescData)
	})
	return file_api_proto_order_UpdateProduct_proto_rawDescData
}

var file_api_proto_order_UpdateProduct_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_api_proto_order_UpdateProduct_proto_goTypes = []interface{}{
	(*UpdateProductReq)(nil), // 0: cloud.lianmi.im.order.UpdateProductReq
	(*UpdateProductRsp)(nil), // 1: cloud.lianmi.im.order.UpdateProductRsp
	(*Product)(nil),          // 2: cloud.lianmi.im.order.Product
}
var file_api_proto_order_UpdateProduct_proto_depIdxs = []int32{
	2, // 0: cloud.lianmi.im.order.UpdateProductReq.product:type_name -> cloud.lianmi.im.order.Product
	2, // 1: cloud.lianmi.im.order.UpdateProductRsp.product:type_name -> cloud.lianmi.im.order.Product
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_api_proto_order_UpdateProduct_proto_init() }
func file_api_proto_order_UpdateProduct_proto_init() {
	if File_api_proto_order_UpdateProduct_proto != nil {
		return
	}
	file_api_proto_order_Product_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_api_proto_order_UpdateProduct_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateProductReq); i {
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
		file_api_proto_order_UpdateProduct_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UpdateProductRsp); i {
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
			RawDescriptor: file_api_proto_order_UpdateProduct_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_proto_order_UpdateProduct_proto_goTypes,
		DependencyIndexes: file_api_proto_order_UpdateProduct_proto_depIdxs,
		MessageInfos:      file_api_proto_order_UpdateProduct_proto_msgTypes,
	}.Build()
	File_api_proto_order_UpdateProduct_proto = out.File
	file_api_proto_order_UpdateProduct_proto_rawDesc = nil
	file_api_proto_order_UpdateProduct_proto_goTypes = nil
	file_api_proto_order_UpdateProduct_proto_depIdxs = nil
}
