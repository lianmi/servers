// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.7.0
// source: api/proto/msg/grpc.proto

package msg

import (
	context "context"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
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

var File_api_proto_msg_grpc_proto protoreflect.FileDescriptor

var file_api_proto_msg_grpc_proto_rawDesc = []byte{
	0x0a, 0x18, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x73, 0x67, 0x2f,
	0x67, 0x72, 0x70, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x13, 0x63, 0x6c, 0x6f, 0x75,
	0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x6d, 0x73, 0x67, 0x32,
	0x0c, 0x0a, 0x0a, 0x4c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x43, 0x68, 0x61, 0x74, 0x42, 0x29, 0x5a,
	0x27, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x69, 0x61, 0x6e,
	0x6d, 0x69, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x73, 0x67, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var file_api_proto_msg_grpc_proto_goTypes = []interface{}{}
var file_api_proto_msg_grpc_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_api_proto_msg_grpc_proto_init() }
func file_api_proto_msg_grpc_proto_init() {
	if File_api_proto_msg_grpc_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_api_proto_msg_grpc_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_api_proto_msg_grpc_proto_goTypes,
		DependencyIndexes: file_api_proto_msg_grpc_proto_depIdxs,
	}.Build()
	File_api_proto_msg_grpc_proto = out.File
	file_api_proto_msg_grpc_proto_rawDesc = nil
	file_api_proto_msg_grpc_proto_goTypes = nil
	file_api_proto_msg_grpc_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// LianmiChatClient is the client API for LianmiChat service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type LianmiChatClient interface {
}

type lianmiChatClient struct {
	cc grpc.ClientConnInterface
}

func NewLianmiChatClient(cc grpc.ClientConnInterface) LianmiChatClient {
	return &lianmiChatClient{cc}
}

// LianmiChatServer is the server API for LianmiChat service.
type LianmiChatServer interface {
}

// UnimplementedLianmiChatServer can be embedded to have forward compatible implementations.
type UnimplementedLianmiChatServer struct {
}

func RegisterLianmiChatServer(s *grpc.Server, srv LianmiChatServer) {
	s.RegisterService(&_LianmiChat_serviceDesc, srv)
}

var _LianmiChat_serviceDesc = grpc.ServiceDesc{
	ServiceName: "cloud.lianmi.im.msg.LianmiChat",
	HandlerType: (*LianmiChatServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams:     []grpc.StreamDesc{},
	Metadata:    "api/proto/msg/grpc.proto",
}
