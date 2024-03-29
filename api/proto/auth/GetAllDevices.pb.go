// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.7.0
// source: api/proto/auth/GetAllDevices.proto

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

type GetAllDevicesRsp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	OnlineDevices  []*DeviceInfo `protobuf:"bytes,1,rep,name=onlineDevices,proto3" json:"onlineDevices,omitempty"`
	OfflineDevices []*DeviceInfo `protobuf:"bytes,2,rep,name=offlineDevices,proto3" json:"offlineDevices,omitempty"`
}

func (x *GetAllDevicesRsp) Reset() {
	*x = GetAllDevicesRsp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_auth_GetAllDevices_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetAllDevicesRsp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetAllDevicesRsp) ProtoMessage() {}

func (x *GetAllDevicesRsp) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_auth_GetAllDevices_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetAllDevicesRsp.ProtoReflect.Descriptor instead.
func (*GetAllDevicesRsp) Descriptor() ([]byte, []int) {
	return file_api_proto_auth_GetAllDevices_proto_rawDescGZIP(), []int{0}
}

func (x *GetAllDevicesRsp) GetOnlineDevices() []*DeviceInfo {
	if x != nil {
		return x.OnlineDevices
	}
	return nil
}

func (x *GetAllDevicesRsp) GetOfflineDevices() []*DeviceInfo {
	if x != nil {
		return x.OfflineDevices
	}
	return nil
}

var File_api_proto_auth_GetAllDevices_proto protoreflect.FileDescriptor

var file_api_proto_auth_GetAllDevices_proto_rawDesc = []byte{
	0x0a, 0x22, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x61, 0x75, 0x74, 0x68,
	0x2f, 0x47, 0x65, 0x74, 0x41, 0x6c, 0x6c, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x73, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x14, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e,
	0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x1a, 0x1b, 0x61, 0x70, 0x69, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x61, 0x75, 0x74, 0x68, 0x2f, 0x53, 0x69, 0x67, 0x6e, 0x49,
	0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xa4, 0x01, 0x0a, 0x10, 0x47, 0x65, 0x74, 0x41,
	0x6c, 0x6c, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x73, 0x52, 0x73, 0x70, 0x12, 0x46, 0x0a, 0x0d,
	0x6f, 0x6e, 0x6c, 0x69, 0x6e, 0x65, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x73, 0x18, 0x01, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e,
	0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x61, 0x75, 0x74, 0x68, 0x2e, 0x44, 0x65, 0x76, 0x69, 0x63,
	0x65, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x0d, 0x6f, 0x6e, 0x6c, 0x69, 0x6e, 0x65, 0x44, 0x65, 0x76,
	0x69, 0x63, 0x65, 0x73, 0x12, 0x48, 0x0a, 0x0e, 0x6f, 0x66, 0x66, 0x6c, 0x69, 0x6e, 0x65, 0x44,
	0x65, 0x76, 0x69, 0x63, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x63,
	0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e, 0x69, 0x6d, 0x2e, 0x61,
	0x75, 0x74, 0x68, 0x2e, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x0e,
	0x6f, 0x66, 0x66, 0x6c, 0x69, 0x6e, 0x65, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x73, 0x42, 0x2a,
	0x5a, 0x28, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x69, 0x61,
	0x6e, 0x6d, 0x69, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x61, 0x75, 0x74, 0x68, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_api_proto_auth_GetAllDevices_proto_rawDescOnce sync.Once
	file_api_proto_auth_GetAllDevices_proto_rawDescData = file_api_proto_auth_GetAllDevices_proto_rawDesc
)

func file_api_proto_auth_GetAllDevices_proto_rawDescGZIP() []byte {
	file_api_proto_auth_GetAllDevices_proto_rawDescOnce.Do(func() {
		file_api_proto_auth_GetAllDevices_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_proto_auth_GetAllDevices_proto_rawDescData)
	})
	return file_api_proto_auth_GetAllDevices_proto_rawDescData
}

var file_api_proto_auth_GetAllDevices_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_api_proto_auth_GetAllDevices_proto_goTypes = []interface{}{
	(*GetAllDevicesRsp)(nil), // 0: cloud.lianmi.im.auth.GetAllDevicesRsp
	(*DeviceInfo)(nil),       // 1: cloud.lianmi.im.auth.DeviceInfo
}
var file_api_proto_auth_GetAllDevices_proto_depIdxs = []int32{
	1, // 0: cloud.lianmi.im.auth.GetAllDevicesRsp.onlineDevices:type_name -> cloud.lianmi.im.auth.DeviceInfo
	1, // 1: cloud.lianmi.im.auth.GetAllDevicesRsp.offlineDevices:type_name -> cloud.lianmi.im.auth.DeviceInfo
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_api_proto_auth_GetAllDevices_proto_init() }
func file_api_proto_auth_GetAllDevices_proto_init() {
	if File_api_proto_auth_GetAllDevices_proto != nil {
		return
	}
	file_api_proto_auth_SignIn_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_api_proto_auth_GetAllDevices_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetAllDevicesRsp); i {
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
			RawDescriptor: file_api_proto_auth_GetAllDevices_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_proto_auth_GetAllDevices_proto_goTypes,
		DependencyIndexes: file_api_proto_auth_GetAllDevices_proto_depIdxs,
		MessageInfos:      file_api_proto_auth_GetAllDevices_proto_msgTypes,
	}.Build()
	File_api_proto_auth_GetAllDevices_proto = out.File
	file_api_proto_auth_GetAllDevices_proto_rawDesc = nil
	file_api_proto_auth_GetAllDevices_proto_goTypes = nil
	file_api_proto_auth_GetAllDevices_proto_depIdxs = nil
}
