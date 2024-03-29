// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.7.0
// source: api/proto/auth/AuthorizeCode.proto

package auth

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

type AuthorizeCodeReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	AppKey          string     `protobuf:"bytes,1,opt,name=appKey,proto3" json:"appKey,omitempty"`
	ClientType      ClientType `protobuf:"varint,2,opt,name=clientType,proto3,enum=cloud.lianmi.im.auth.ClientType" json:"clientType,omitempty"`
	Os              string     `protobuf:"bytes,3,opt,name=os,proto3" json:"os,omitempty"`
	ProtocolVersion string     `protobuf:"bytes,4,opt,name=protocolVersion,proto3" json:"protocolVersion,omitempty"`
	SdkVersion      string     `protobuf:"bytes,5,opt,name=sdkVersion,proto3" json:"sdkVersion,omitempty"`
}

func (x *AuthorizeCodeReq) Reset() {
	*x = AuthorizeCodeReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_auth_AuthorizeCode_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AuthorizeCodeReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthorizeCodeReq) ProtoMessage() {}

func (x *AuthorizeCodeReq) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_auth_AuthorizeCode_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthorizeCodeReq.ProtoReflect.Descriptor instead.
func (*AuthorizeCodeReq) Descriptor() ([]byte, []int) {
	return file_api_proto_auth_AuthorizeCode_proto_rawDescGZIP(), []int{0}
}

func (x *AuthorizeCodeReq) GetAppKey() string {
	if x != nil {
		return x.AppKey
	}
	return ""
}

func (x *AuthorizeCodeReq) GetClientType() ClientType {
	if x != nil {
		return x.ClientType
	}
	return ClientType_Ct_UnKnow
}

func (x *AuthorizeCodeReq) GetOs() string {
	if x != nil {
		return x.Os
	}
	return ""
}

func (x *AuthorizeCodeReq) GetProtocolVersion() string {
	if x != nil {
		return x.ProtocolVersion
	}
	return ""
}

func (x *AuthorizeCodeReq) GetSdkVersion() string {
	if x != nil {
		return x.SdkVersion
	}
	return ""
}

type AuthorizeCodeRsp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Code string `protobuf:"bytes,1,opt,name=code,proto3" json:"code,omitempty"`
}

func (x *AuthorizeCodeRsp) Reset() {
	*x = AuthorizeCodeRsp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_auth_AuthorizeCode_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AuthorizeCodeRsp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthorizeCodeRsp) ProtoMessage() {}

func (x *AuthorizeCodeRsp) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_auth_AuthorizeCode_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthorizeCodeRsp.ProtoReflect.Descriptor instead.
func (*AuthorizeCodeRsp) Descriptor() ([]byte, []int) {
	return file_api_proto_auth_AuthorizeCode_proto_rawDescGZIP(), []int{1}
}

func (x *AuthorizeCodeRsp) GetCode() string {
	if x != nil {
		return x.Code
	}
	return ""
}

var File_api_proto_auth_AuthorizeCode_proto protoreflect.FileDescriptor

var file_api_proto_auth_AuthorizeCode_proto_rawDesc = []byte{
	0x0a, 0x22, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x61, 0x75, 0x74, 0x68,
	0x2f, 0x41, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x69, 0x7a, 0x65, 0x43, 0x6f, 0x64, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x14, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e,
	0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x1a, 0x1b, 0x61, 0x70, 0x69, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x61, 0x75, 0x74, 0x68, 0x2f, 0x53, 0x69, 0x67, 0x6e, 0x49,
	0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xc6, 0x01, 0x0a, 0x10, 0x41, 0x75, 0x74, 0x68,
	0x6f, 0x72, 0x69, 0x7a, 0x65, 0x43, 0x6f, 0x64, 0x65, 0x52, 0x65, 0x71, 0x12, 0x16, 0x0a, 0x06,
	0x61, 0x70, 0x70, 0x4b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x61, 0x70,
	0x70, 0x4b, 0x65, 0x79, 0x12, 0x40, 0x0a, 0x0a, 0x63, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x54, 0x79,
	0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x20, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64,
	0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e,
	0x43, 0x6c, 0x69, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x52, 0x0a, 0x63, 0x6c, 0x69, 0x65,
	0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x6f, 0x73, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x02, 0x6f, 0x73, 0x12, 0x28, 0x0a, 0x0f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63,
	0x6f, 0x6c, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x6f, 0x6c, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e,
	0x12, 0x1e, 0x0a, 0x0a, 0x73, 0x64, 0x6b, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x73, 0x64, 0x6b, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e,
	0x22, 0x26, 0x0a, 0x10, 0x41, 0x75, 0x74, 0x68, 0x6f, 0x72, 0x69, 0x7a, 0x65, 0x43, 0x6f, 0x64,
	0x65, 0x52, 0x73, 0x70, 0x12, 0x12, 0x0a, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x63, 0x6f, 0x64, 0x65, 0x42, 0x2a, 0x5a, 0x28, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2f, 0x73, 0x65,
	0x72, 0x76, 0x65, 0x72, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f,
	0x61, 0x75, 0x74, 0x68, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_proto_auth_AuthorizeCode_proto_rawDescOnce sync.Once
	file_api_proto_auth_AuthorizeCode_proto_rawDescData = file_api_proto_auth_AuthorizeCode_proto_rawDesc
)

func file_api_proto_auth_AuthorizeCode_proto_rawDescGZIP() []byte {
	file_api_proto_auth_AuthorizeCode_proto_rawDescOnce.Do(func() {
		file_api_proto_auth_AuthorizeCode_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_proto_auth_AuthorizeCode_proto_rawDescData)
	})
	return file_api_proto_auth_AuthorizeCode_proto_rawDescData
}

var file_api_proto_auth_AuthorizeCode_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_api_proto_auth_AuthorizeCode_proto_goTypes = []interface{}{
	(*AuthorizeCodeReq)(nil), // 0: cloud.lianmi.im.auth.AuthorizeCodeReq
	(*AuthorizeCodeRsp)(nil), // 1: cloud.lianmi.im.auth.AuthorizeCodeRsp
	(ClientType)(0),          // 2: cloud.lianmi.im.auth.ClientType
}
var file_api_proto_auth_AuthorizeCode_proto_depIdxs = []int32{
	2, // 0: cloud.lianmi.im.auth.AuthorizeCodeReq.clientType:type_name -> cloud.lianmi.im.auth.ClientType
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_api_proto_auth_AuthorizeCode_proto_init() }
func file_api_proto_auth_AuthorizeCode_proto_init() {
	if File_api_proto_auth_AuthorizeCode_proto != nil {
		return
	}
	file_api_proto_auth_SignIn_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_api_proto_auth_AuthorizeCode_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AuthorizeCodeReq); i {
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
		file_api_proto_auth_AuthorizeCode_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AuthorizeCodeRsp); i {
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
			RawDescriptor: file_api_proto_auth_AuthorizeCode_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_proto_auth_AuthorizeCode_proto_goTypes,
		DependencyIndexes: file_api_proto_auth_AuthorizeCode_proto_depIdxs,
		MessageInfos:      file_api_proto_auth_AuthorizeCode_proto_msgTypes,
	}.Build()
	File_api_proto_auth_AuthorizeCode_proto = out.File
	file_api_proto_auth_AuthorizeCode_proto_rawDesc = nil
	file_api_proto_auth_AuthorizeCode_proto_goTypes = nil
	file_api_proto_auth_AuthorizeCode_proto_depIdxs = nil
}
