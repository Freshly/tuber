// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.13.0
// source: pkg/proto/tuber_service.proto

package proto

import (
	context "context"
	reflect "reflect"
	sync "sync"

	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
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

type CreateReviewAppRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Token   string `protobuf:"bytes,1,opt,name=token,proto3" json:"token,omitempty"`
	AppName string `protobuf:"bytes,2,opt,name=appName,proto3" json:"appName,omitempty"`
	Branch  string `protobuf:"bytes,3,opt,name=branch,proto3" json:"branch,omitempty"`
}

func (x *CreateReviewAppRequest) Reset() {
	*x = CreateReviewAppRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_proto_tuber_service_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateReviewAppRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateReviewAppRequest) ProtoMessage() {}

func (x *CreateReviewAppRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_proto_tuber_service_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateReviewAppRequest.ProtoReflect.Descriptor instead.
func (*CreateReviewAppRequest) Descriptor() ([]byte, []int) {
	return file_pkg_proto_tuber_service_proto_rawDescGZIP(), []int{0}
}

func (x *CreateReviewAppRequest) GetToken() string {
	if x != nil {
		return x.Token
	}
	return ""
}

func (x *CreateReviewAppRequest) GetAppName() string {
	if x != nil {
		return x.AppName
	}
	return ""
}

func (x *CreateReviewAppRequest) GetBranch() string {
	if x != nil {
		return x.Branch
	}
	return ""
}

type CreateReviewAppResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Hostname string `protobuf:"bytes,1,opt,name=hostname,proto3" json:"hostname,omitempty"`
	Error    string `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *CreateReviewAppResponse) Reset() {
	*x = CreateReviewAppResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_proto_tuber_service_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateReviewAppResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateReviewAppResponse) ProtoMessage() {}

func (x *CreateReviewAppResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_proto_tuber_service_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateReviewAppResponse.ProtoReflect.Descriptor instead.
func (*CreateReviewAppResponse) Descriptor() ([]byte, []int) {
	return file_pkg_proto_tuber_service_proto_rawDescGZIP(), []int{1}
}

func (x *CreateReviewAppResponse) GetHostname() string {
	if x != nil {
		return x.Hostname
	}
	return ""
}

func (x *CreateReviewAppResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

type DeleteReviewAppRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Token   string `protobuf:"bytes,1,opt,name=token,proto3" json:"token,omitempty"`
	AppName string `protobuf:"bytes,2,opt,name=appName,proto3" json:"appName,omitempty"`
}

func (x *DeleteReviewAppRequest) Reset() {
	*x = DeleteReviewAppRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_proto_tuber_service_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteReviewAppRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteReviewAppRequest) ProtoMessage() {}

func (x *DeleteReviewAppRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_proto_tuber_service_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteReviewAppRequest.ProtoReflect.Descriptor instead.
func (*DeleteReviewAppRequest) Descriptor() ([]byte, []int) {
	return file_pkg_proto_tuber_service_proto_rawDescGZIP(), []int{2}
}

func (x *DeleteReviewAppRequest) GetToken() string {
	if x != nil {
		return x.Token
	}
	return ""
}

func (x *DeleteReviewAppRequest) GetAppName() string {
	if x != nil {
		return x.AppName
	}
	return ""
}

type DeleteReviewAppResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	AppName string `protobuf:"bytes,1,opt,name=appName,proto3" json:"appName,omitempty"`
	Error   string `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
}

func (x *DeleteReviewAppResponse) Reset() {
	*x = DeleteReviewAppResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_proto_tuber_service_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DeleteReviewAppResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DeleteReviewAppResponse) ProtoMessage() {}

func (x *DeleteReviewAppResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_proto_tuber_service_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DeleteReviewAppResponse.ProtoReflect.Descriptor instead.
func (*DeleteReviewAppResponse) Descriptor() ([]byte, []int) {
	return file_pkg_proto_tuber_service_proto_rawDescGZIP(), []int{3}
}

func (x *DeleteReviewAppResponse) GetAppName() string {
	if x != nil {
		return x.AppName
	}
	return ""
}

func (x *DeleteReviewAppResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

var File_pkg_proto_tuber_service_proto protoreflect.FileDescriptor

var file_pkg_proto_tuber_service_proto_rawDesc = []byte{
	0x0a, 0x1d, 0x70, 0x6b, 0x67, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x74, 0x75, 0x62, 0x65,
	0x72, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x05, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x60, 0x0a, 0x16, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65,
	0x52, 0x65, 0x76, 0x69, 0x65, 0x77, 0x41, 0x70, 0x70, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x12, 0x14, 0x0a, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x70, 0x70, 0x4e, 0x61, 0x6d,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x70, 0x70, 0x4e, 0x61, 0x6d, 0x65,
	0x12, 0x16, 0x0a, 0x06, 0x62, 0x72, 0x61, 0x6e, 0x63, 0x68, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x06, 0x62, 0x72, 0x61, 0x6e, 0x63, 0x68, 0x22, 0x4b, 0x0a, 0x17, 0x43, 0x72, 0x65, 0x61,
	0x74, 0x65, 0x52, 0x65, 0x76, 0x69, 0x65, 0x77, 0x41, 0x70, 0x70, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x68, 0x6f, 0x73, 0x74, 0x6e, 0x61, 0x6d, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x68, 0x6f, 0x73, 0x74, 0x6e, 0x61, 0x6d, 0x65, 0x12,
	0x14, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x65, 0x72, 0x72, 0x6f, 0x72, 0x22, 0x48, 0x0a, 0x16, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x52,
	0x65, 0x76, 0x69, 0x65, 0x77, 0x41, 0x70, 0x70, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x14, 0x0a, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05,
	0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x70, 0x70, 0x4e, 0x61, 0x6d, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x70, 0x70, 0x4e, 0x61, 0x6d, 0x65, 0x22,
	0x49, 0x0a, 0x17, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x52, 0x65, 0x76, 0x69, 0x65, 0x77, 0x41,
	0x70, 0x70, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x61, 0x70,
	0x70, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x61, 0x70, 0x70,
	0x4e, 0x61, 0x6d, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x72, 0x72, 0x6f, 0x72, 0x32, 0xaf, 0x01, 0x0a, 0x05, 0x54,
	0x75, 0x62, 0x65, 0x72, 0x12, 0x52, 0x0a, 0x0f, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x52, 0x65,
	0x76, 0x69, 0x65, 0x77, 0x41, 0x70, 0x70, 0x12, 0x1d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e,
	0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x52, 0x65, 0x76, 0x69, 0x65, 0x77, 0x41, 0x70, 0x70, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x43,
	0x72, 0x65, 0x61, 0x74, 0x65, 0x52, 0x65, 0x76, 0x69, 0x65, 0x77, 0x41, 0x70, 0x70, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x52, 0x0a, 0x0f, 0x44, 0x65, 0x6c, 0x65,
	0x74, 0x65, 0x52, 0x65, 0x76, 0x69, 0x65, 0x77, 0x41, 0x70, 0x70, 0x12, 0x1d, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x52, 0x65, 0x76, 0x69, 0x65, 0x77,
	0x41, 0x70, 0x70, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1e, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x2e, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x52, 0x65, 0x76, 0x69, 0x65, 0x77, 0x41,
	0x70, 0x70, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pkg_proto_tuber_service_proto_rawDescOnce sync.Once
	file_pkg_proto_tuber_service_proto_rawDescData = file_pkg_proto_tuber_service_proto_rawDesc
)

func file_pkg_proto_tuber_service_proto_rawDescGZIP() []byte {
	file_pkg_proto_tuber_service_proto_rawDescOnce.Do(func() {
		file_pkg_proto_tuber_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_proto_tuber_service_proto_rawDescData)
	})
	return file_pkg_proto_tuber_service_proto_rawDescData
}

var file_pkg_proto_tuber_service_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_pkg_proto_tuber_service_proto_goTypes = []interface{}{
	(*CreateReviewAppRequest)(nil),  // 0: proto.CreateReviewAppRequest
	(*CreateReviewAppResponse)(nil), // 1: proto.CreateReviewAppResponse
	(*DeleteReviewAppRequest)(nil),  // 2: proto.DeleteReviewAppRequest
	(*DeleteReviewAppResponse)(nil), // 3: proto.DeleteReviewAppResponse
}
var file_pkg_proto_tuber_service_proto_depIdxs = []int32{
	0, // 0: proto.Tuber.CreateReviewApp:input_type -> proto.CreateReviewAppRequest
	2, // 1: proto.Tuber.DeleteReviewApp:input_type -> proto.DeleteReviewAppRequest
	1, // 2: proto.Tuber.CreateReviewApp:output_type -> proto.CreateReviewAppResponse
	3, // 3: proto.Tuber.DeleteReviewApp:output_type -> proto.DeleteReviewAppResponse
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_pkg_proto_tuber_service_proto_init() }
func file_pkg_proto_tuber_service_proto_init() {
	if File_pkg_proto_tuber_service_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pkg_proto_tuber_service_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateReviewAppRequest); i {
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
		file_pkg_proto_tuber_service_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateReviewAppResponse); i {
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
		file_pkg_proto_tuber_service_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteReviewAppRequest); i {
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
		file_pkg_proto_tuber_service_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DeleteReviewAppResponse); i {
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
			RawDescriptor: file_pkg_proto_tuber_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_pkg_proto_tuber_service_proto_goTypes,
		DependencyIndexes: file_pkg_proto_tuber_service_proto_depIdxs,
		MessageInfos:      file_pkg_proto_tuber_service_proto_msgTypes,
	}.Build()
	File_pkg_proto_tuber_service_proto = out.File
	file_pkg_proto_tuber_service_proto_rawDesc = nil
	file_pkg_proto_tuber_service_proto_goTypes = nil
	file_pkg_proto_tuber_service_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// TuberClient is the client API for Tuber service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type TuberClient interface {
	CreateReviewApp(ctx context.Context, in *CreateReviewAppRequest, opts ...grpc.CallOption) (*CreateReviewAppResponse, error)
	DeleteReviewApp(ctx context.Context, in *DeleteReviewAppRequest, opts ...grpc.CallOption) (*DeleteReviewAppResponse, error)
}

type tuberClient struct {
	cc grpc.ClientConnInterface
}

func NewTuberClient(cc grpc.ClientConnInterface) TuberClient {
	return &tuberClient{cc}
}

func (c *tuberClient) CreateReviewApp(ctx context.Context, in *CreateReviewAppRequest, opts ...grpc.CallOption) (*CreateReviewAppResponse, error) {
	out := new(CreateReviewAppResponse)
	err := c.cc.Invoke(ctx, "/proto.Tuber/CreateReviewApp", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tuberClient) DeleteReviewApp(ctx context.Context, in *DeleteReviewAppRequest, opts ...grpc.CallOption) (*DeleteReviewAppResponse, error) {
	out := new(DeleteReviewAppResponse)
	err := c.cc.Invoke(ctx, "/proto.Tuber/DeleteReviewApp", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TuberServer is the server API for Tuber service.
type TuberServer interface {
	CreateReviewApp(context.Context, *CreateReviewAppRequest) (*CreateReviewAppResponse, error)
	DeleteReviewApp(context.Context, *DeleteReviewAppRequest) (*DeleteReviewAppResponse, error)
}

// UnimplementedTuberServer can be embedded to have forward compatible implementations.
type UnimplementedTuberServer struct {
}

func (*UnimplementedTuberServer) CreateReviewApp(context.Context, *CreateReviewAppRequest) (*CreateReviewAppResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateReviewApp not implemented")
}
func (*UnimplementedTuberServer) DeleteReviewApp(context.Context, *DeleteReviewAppRequest) (*DeleteReviewAppResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteReviewApp not implemented")
}

func RegisterTuberServer(s *grpc.Server, srv TuberServer) {
	s.RegisterService(&_Tuber_serviceDesc, srv)
}

func _Tuber_CreateReviewApp_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateReviewAppRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TuberServer).CreateReviewApp(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Tuber/CreateReviewApp",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TuberServer).CreateReviewApp(ctx, req.(*CreateReviewAppRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Tuber_DeleteReviewApp_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteReviewAppRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TuberServer).DeleteReviewApp(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Tuber/DeleteReviewApp",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TuberServer).DeleteReviewApp(ctx, req.(*DeleteReviewAppRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Tuber_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proto.Tuber",
	HandlerType: (*TuberServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateReviewApp",
			Handler:    _Tuber_CreateReviewApp_Handler,
		},
		{
			MethodName: "DeleteReviewApp",
			Handler:    _Tuber_DeleteReviewApp_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pkg/proto/tuber_service.proto",
}
