// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.7.0
// source: api/proto/msg/GetOssToken.proto

package msg

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

//向服务端发送获取阿里云OSS上传Token的请求，用业务号及子号即可
type GetOssTokenReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//新增，如果true表示私有, false表示公有目录
	IsPrivate bool `protobuf:"varint,1,opt,name=isPrivate,proto3" json:"isPrivate,omitempty"`
}

func (x *GetOssTokenReq) Reset() {
	*x = GetOssTokenReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_msg_GetOssToken_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetOssTokenReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetOssTokenReq) ProtoMessage() {}

func (x *GetOssTokenReq) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_msg_GetOssToken_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetOssTokenReq.ProtoReflect.Descriptor instead.
func (*GetOssTokenReq) Descriptor() ([]byte, []int) {
	return file_api_proto_msg_GetOssToken_proto_rawDescGZIP(), []int{0}
}

func (x *GetOssTokenReq) GetIsPrivate() bool {
	if x != nil {
		return x.IsPrivate
	}
	return false
}

//响应参数
type GetOssTokenRsp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	//资源服务器地址
	EndPoint string `protobuf:"bytes,1,opt,name=endPoint,proto3" json:"endPoint,omitempty"`
	//空间名称
	BucketName string `protobuf:"bytes,2,opt,name=bucketName,proto3" json:"bucketName,omitempty"`
	// Bucket访问凭证
	AccessKeyId string `protobuf:"bytes,3,opt,name=accessKeyId,proto3" json:"accessKeyId,omitempty"`
	// Bucket访问密钥
	AccessKeySecret string `protobuf:"bytes,4,opt,name=accessKeySecret,proto3" json:"accessKeySecret,omitempty"`
	// 安全凭证
	SecurityToken string `protobuf:"bytes,5,opt,name=securityToken,proto3" json:"securityToken,omitempty"`
	// oss的文件目录，日期为目录名, 如：  2020/8/28， 客户端需要拼接为完整的上传文件名
	Directory string `protobuf:"bytes,6,opt,name=directory,proto3" json:"directory,omitempty"`
	//token有效时长(单位S)
	Expire uint64 `protobuf:"fixed64,7,opt,name=expire,proto3" json:"expire,omitempty"`
	//服务器按json格式组装
	Callback string `protobuf:"bytes,8,opt,name=callback,proto3" json:"callback,omitempty"`
}

func (x *GetOssTokenRsp) Reset() {
	*x = GetOssTokenRsp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_proto_msg_GetOssToken_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetOssTokenRsp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetOssTokenRsp) ProtoMessage() {}

func (x *GetOssTokenRsp) ProtoReflect() protoreflect.Message {
	mi := &file_api_proto_msg_GetOssToken_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetOssTokenRsp.ProtoReflect.Descriptor instead.
func (*GetOssTokenRsp) Descriptor() ([]byte, []int) {
	return file_api_proto_msg_GetOssToken_proto_rawDescGZIP(), []int{1}
}

func (x *GetOssTokenRsp) GetEndPoint() string {
	if x != nil {
		return x.EndPoint
	}
	return ""
}

func (x *GetOssTokenRsp) GetBucketName() string {
	if x != nil {
		return x.BucketName
	}
	return ""
}

func (x *GetOssTokenRsp) GetAccessKeyId() string {
	if x != nil {
		return x.AccessKeyId
	}
	return ""
}

func (x *GetOssTokenRsp) GetAccessKeySecret() string {
	if x != nil {
		return x.AccessKeySecret
	}
	return ""
}

func (x *GetOssTokenRsp) GetSecurityToken() string {
	if x != nil {
		return x.SecurityToken
	}
	return ""
}

func (x *GetOssTokenRsp) GetDirectory() string {
	if x != nil {
		return x.Directory
	}
	return ""
}

func (x *GetOssTokenRsp) GetExpire() uint64 {
	if x != nil {
		return x.Expire
	}
	return 0
}

func (x *GetOssTokenRsp) GetCallback() string {
	if x != nil {
		return x.Callback
	}
	return ""
}

var File_api_proto_msg_GetOssToken_proto protoreflect.FileDescriptor

var file_api_proto_msg_GetOssToken_proto_rawDesc = []byte{
	0x0a, 0x1f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6d, 0x73, 0x67, 0x2f,
	0x47, 0x65, 0x74, 0x4f, 0x73, 0x73, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x12, 0x13, 0x63, 0x6c, 0x6f, 0x75, 0x64, 0x2e, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2e,
	0x69, 0x6d, 0x2e, 0x6d, 0x73, 0x67, 0x22, 0x2e, 0x0a, 0x0e, 0x47, 0x65, 0x74, 0x4f, 0x73, 0x73,
	0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x52, 0x65, 0x71, 0x12, 0x1c, 0x0a, 0x09, 0x69, 0x73, 0x50, 0x72,
	0x69, 0x76, 0x61, 0x74, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x09, 0x69, 0x73, 0x50,
	0x72, 0x69, 0x76, 0x61, 0x74, 0x65, 0x22, 0x90, 0x02, 0x0a, 0x0e, 0x47, 0x65, 0x74, 0x4f, 0x73,
	0x73, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x52, 0x73, 0x70, 0x12, 0x1a, 0x0a, 0x08, 0x65, 0x6e, 0x64,
	0x50, 0x6f, 0x69, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x65, 0x6e, 0x64,
	0x50, 0x6f, 0x69, 0x6e, 0x74, 0x12, 0x1e, 0x0a, 0x0a, 0x62, 0x75, 0x63, 0x6b, 0x65, 0x74, 0x4e,
	0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x62, 0x75, 0x63, 0x6b, 0x65,
	0x74, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x20, 0x0a, 0x0b, 0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x4b,
	0x65, 0x79, 0x49, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x61, 0x63, 0x63, 0x65,
	0x73, 0x73, 0x4b, 0x65, 0x79, 0x49, 0x64, 0x12, 0x28, 0x0a, 0x0f, 0x61, 0x63, 0x63, 0x65, 0x73,
	0x73, 0x4b, 0x65, 0x79, 0x53, 0x65, 0x63, 0x72, 0x65, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0f, 0x61, 0x63, 0x63, 0x65, 0x73, 0x73, 0x4b, 0x65, 0x79, 0x53, 0x65, 0x63, 0x72, 0x65,
	0x74, 0x12, 0x24, 0x0a, 0x0d, 0x73, 0x65, 0x63, 0x75, 0x72, 0x69, 0x74, 0x79, 0x54, 0x6f, 0x6b,
	0x65, 0x6e, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x73, 0x65, 0x63, 0x75, 0x72, 0x69,
	0x74, 0x79, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x1c, 0x0a, 0x09, 0x64, 0x69, 0x72, 0x65, 0x63,
	0x74, 0x6f, 0x72, 0x79, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x64, 0x69, 0x72, 0x65,
	0x63, 0x74, 0x6f, 0x72, 0x79, 0x12, 0x16, 0x0a, 0x06, 0x65, 0x78, 0x70, 0x69, 0x72, 0x65, 0x18,
	0x07, 0x20, 0x01, 0x28, 0x06, 0x52, 0x06, 0x65, 0x78, 0x70, 0x69, 0x72, 0x65, 0x12, 0x1a, 0x0a,
	0x08, 0x63, 0x61, 0x6c, 0x6c, 0x62, 0x61, 0x63, 0x6b, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x08, 0x63, 0x61, 0x6c, 0x6c, 0x62, 0x61, 0x63, 0x6b, 0x42, 0x29, 0x5a, 0x27, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6c, 0x69, 0x61, 0x6e, 0x6d, 0x69, 0x2f, 0x73,
	0x65, 0x72, 0x76, 0x65, 0x72, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2f, 0x6d, 0x73, 0x67, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_proto_msg_GetOssToken_proto_rawDescOnce sync.Once
	file_api_proto_msg_GetOssToken_proto_rawDescData = file_api_proto_msg_GetOssToken_proto_rawDesc
)

func file_api_proto_msg_GetOssToken_proto_rawDescGZIP() []byte {
	file_api_proto_msg_GetOssToken_proto_rawDescOnce.Do(func() {
		file_api_proto_msg_GetOssToken_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_proto_msg_GetOssToken_proto_rawDescData)
	})
	return file_api_proto_msg_GetOssToken_proto_rawDescData
}

var file_api_proto_msg_GetOssToken_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_api_proto_msg_GetOssToken_proto_goTypes = []interface{}{
	(*GetOssTokenReq)(nil), // 0: cloud.lianmi.im.msg.GetOssTokenReq
	(*GetOssTokenRsp)(nil), // 1: cloud.lianmi.im.msg.GetOssTokenRsp
}
var file_api_proto_msg_GetOssToken_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_api_proto_msg_GetOssToken_proto_init() }
func file_api_proto_msg_GetOssToken_proto_init() {
	if File_api_proto_msg_GetOssToken_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_proto_msg_GetOssToken_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetOssTokenReq); i {
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
		file_api_proto_msg_GetOssToken_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetOssTokenRsp); i {
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
			RawDescriptor: file_api_proto_msg_GetOssToken_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_proto_msg_GetOssToken_proto_goTypes,
		DependencyIndexes: file_api_proto_msg_GetOssToken_proto_depIdxs,
		MessageInfos:      file_api_proto_msg_GetOssToken_proto_msgTypes,
	}.Build()
	File_api_proto_msg_GetOssToken_proto = out.File
	file_api_proto_msg_GetOssToken_proto_rawDesc = nil
	file_api_proto_msg_GetOssToken_proto_goTypes = nil
	file_api_proto_msg_GetOssToken_proto_depIdxs = nil
}
