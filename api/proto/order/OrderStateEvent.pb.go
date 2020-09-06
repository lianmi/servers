// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.7.0
// source: api/proto/order/OrderStateEvent.proto

package order

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

type OrderStateEventRsp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//订单数据体
	OrderBody *Product `protobuf:"bytes,1,opt,name=orderBody,proto3" json:"orderBody,omitempty"`
	//新状态
	State global.OrderState `protobuf:"varint,2,opt,name=state,proto3,enum=cloud.lianmi.im.global.OrderState" json:"state,omitempty"`
	//状态变化的Unix时间戳
	TimeAt int64 `protobuf:"varint,3,opt,name=timeAt,proto3" json:"timeAt,omitempty"`
}

func (x *OrderStateEventRsp) Reset() {
	*x = OrderStateEventRsp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_order_OrderStateEvent_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *OrderStateEventRsp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OrderStateEventRsp) ProtoMessage() {}

func (x *OrderStateEventRsp) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_order_OrderStateEvent_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OrderStateEventRsp.ProtoReflect.Descriptor instead.
func (*OrderStateEventRsp) Descriptor() ([]byte, []int) {
	return file_api_proto_order_OrderStateEvent_proto_rawDescGZIP(), []int{0}
}

func (x *OrderStateEventRsp) GetOrderBody() *Product {
	if x != nil {
		return x.OrderBody
	}
	return nil
}

func (x *OrderStateEventRsp) GetState() global.OrderState {
	if x != nil {
		return x.State
	}
	return global.OrderState_OS_Undefined
}

func (x *OrderStateEventRsp) GetTimeAt() int64 {
	if x != nil {
		return x.TimeAt
	}
	return 0
}

var File_api_proto_order_OrderStateEvent_proto protoreflect.FileDescriptor

var file_api_proto_order_OrderStateEvent_proto_rawDesc = []byte{
	0x0a, 0x25, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6f, 0x72, 0x64, 0x65,
	0x72, 0x2f, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x53, 0x74, 0x61, 0x74, 0x65, 0x45, 0x76, 0x65, 0x6e,
	0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x15, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c,
	0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x1a, 0x1d,
	0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x67, 0x6c, 0x6f, 0x62, 0x61, 0x6c,
	0x2f, 0x47, 0x6c, 0x6f, 0x62, 0x61, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1d, 0x61,
	0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x2f, 0x50,
	0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xa4, 0x01, 0x0a,
	0x12, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x53, 0x74, 0x61, 0x74, 0x65, 0x45, 0x76, 0x65, 0x6e, 0x74,
	0x52, 0x73, 0x70, 0x12, 0x3c, 0x0a, 0x09, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x42, 0x6f, 0x64, 0x79,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c,
	0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x2e, 0x50,
	0x72, 0x6f, 0x64, 0x75, 0x63, 0x74, 0x52, 0x09, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x42, 0x6f, 0x64,
	0x79, 0x12, 0x38, 0x0a, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e,
	0x32, 0x22, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e,
	0x69, 0x6d, 0x2e, 0x67, 0x6c, 0x6f, 0x62, 0x61, 0x6c, 0x2e, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x53,
	0x74, 0x61, 0x74, 0x65, 0x52, 0x05, 0x73, 0x74, 0x61, 0x74, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x74,
	0x69, 0x6d, 0x65, 0x41, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x74, 0x69, 0x6d,
	0x65, 0x41, 0x74, 0x42, 0x2b, 0x5a, 0x29, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73,
	0x2f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6f, 0x72, 0x64, 0x65, 0x72,
	0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_proto_order_OrderStateEvent_proto_rawDescOnce sync.Once
	file_api_proto_order_OrderStateEvent_proto_rawDescData = file_api_proto_order_OrderStateEvent_proto_rawDesc
)

func file_api_proto_order_OrderStateEvent_proto_rawDescGZIP() []byte {
	file_api_proto_order_OrderStateEvent_proto_rawDescOnce.Do(func() {
		file_api_proto_order_OrderStateEvent_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_proto_order_OrderStateEvent_proto_rawDescData)
	})
	return file_api_proto_order_OrderStateEvent_proto_rawDescData
}

var file_api_proto_order_OrderStateEvent_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_api_proto_order_OrderStateEvent_proto_goTypes = []interface{}{
	(*OrderStateEventRsp)(nil), // 0: cloud.lianmi.im.order.OrderStateEventRsp
	(*Product)(nil),            // 1: cloud.lianmi.im.order.Product
	(global.OrderState)(0),     // 2: cloud.lianmi.im.global.OrderState
}
var file_api_proto_order_OrderStateEvent_proto_depIdxs = []int32{
	1, // 0: cloud.lianmi.im.order.OrderStateEventRsp.orderBody:type_name -> cloud.lianmi.im.order.Product
	2, // 1: cloud.lianmi.im.order.OrderStateEventRsp.state:type_name -> cloud.lianmi.im.global.OrderState
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_api_proto_order_OrderStateEvent_proto_init() }
func file_api_proto_order_OrderStateEvent_proto_init() {
	if File_api_proto_order_OrderStateEvent_proto != nil {
		return
	}
	file_api_proto_order_Product_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_api_proto_order_OrderStateEvent_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*OrderStateEventRsp); i {
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
			RawDescriptor: file_api_proto_order_OrderStateEvent_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_proto_order_OrderStateEvent_proto_goTypes,
		DependencyIndexes: file_api_proto_order_OrderStateEvent_proto_depIdxs,
		MessageInfos:      file_api_proto_order_OrderStateEvent_proto_msgTypes,
	}.Build()
	File_api_proto_order_OrderStateEvent_proto = out.File
	file_api_proto_order_OrderStateEvent_proto_rawDesc = nil
	file_api_proto_order_OrderStateEvent_proto_goTypes = nil
	file_api_proto_order_OrderStateEvent_proto_depIdxs = nil
}
