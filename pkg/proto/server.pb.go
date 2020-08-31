// Code generated by protoc-gen-go. DO NOT EDIT.
// source: server.proto

package proto

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type Request struct {
	Token                string   `protobuf:"bytes,1,opt,name=token,proto3" json:"token,omitempty"`
	AppName              string   `protobuf:"bytes,2,opt,name=appName,proto3" json:"appName,omitempty"`
	Branch               string   `protobuf:"bytes,3,opt,name=branch,proto3" json:"branch,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Request) Reset()         { *m = Request{} }
func (m *Request) String() string { return proto.CompactTextString(m) }
func (*Request) ProtoMessage()    {}
func (*Request) Descriptor() ([]byte, []int) {
	return fileDescriptor_ad098daeda4239f7, []int{0}
}

func (m *Request) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Request.Unmarshal(m, b)
}
func (m *Request) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Request.Marshal(b, m, deterministic)
}
func (m *Request) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Request.Merge(m, src)
}
func (m *Request) XXX_Size() int {
	return xxx_messageInfo_Request.Size(m)
}
func (m *Request) XXX_DiscardUnknown() {
	xxx_messageInfo_Request.DiscardUnknown(m)
}

var xxx_messageInfo_Request proto.InternalMessageInfo

func (m *Request) GetToken() string {
	if m != nil {
		return m.Token
	}
	return ""
}

func (m *Request) GetAppName() string {
	if m != nil {
		return m.AppName
	}
	return ""
}

func (m *Request) GetBranch() string {
	if m != nil {
		return m.Branch
	}
	return ""
}

type Response struct {
	Hostname             string   `protobuf:"bytes,1,opt,name=hostname,proto3" json:"hostname,omitempty"`
	Error                string   `protobuf:"bytes,2,opt,name=error,proto3" json:"error,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Response) Reset()         { *m = Response{} }
func (m *Response) String() string { return proto.CompactTextString(m) }
func (*Response) ProtoMessage()    {}
func (*Response) Descriptor() ([]byte, []int) {
	return fileDescriptor_ad098daeda4239f7, []int{1}
}

func (m *Response) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Response.Unmarshal(m, b)
}
func (m *Response) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Response.Marshal(b, m, deterministic)
}
func (m *Response) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Response.Merge(m, src)
}
func (m *Response) XXX_Size() int {
	return xxx_messageInfo_Response.Size(m)
}
func (m *Response) XXX_DiscardUnknown() {
	xxx_messageInfo_Response.DiscardUnknown(m)
}

var xxx_messageInfo_Response proto.InternalMessageInfo

func (m *Response) GetHostname() string {
	if m != nil {
		return m.Hostname
	}
	return ""
}

func (m *Response) GetError() string {
	if m != nil {
		return m.Error
	}
	return ""
}

func init() {
	proto.RegisterType((*Request)(nil), "proto.Request")
	proto.RegisterType((*Response)(nil), "proto.Response")
}

func init() {
	proto.RegisterFile("server.proto", fileDescriptor_ad098daeda4239f7)
}

var fileDescriptor_ad098daeda4239f7 = []byte{
	// 191 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x3c, 0x8f, 0xcf, 0x4a, 0xc6, 0x30,
	0x10, 0xc4, 0xad, 0xd2, 0x3f, 0x2e, 0xc5, 0x42, 0x10, 0x09, 0x3d, 0x49, 0x4f, 0x9e, 0x7a, 0x50,
	0x8f, 0x5e, 0x44, 0xcf, 0x82, 0xd1, 0x17, 0x48, 0xcb, 0x40, 0x8b, 0x98, 0xc4, 0x4d, 0x5a, 0x5f,
	0xff, 0xa3, 0x49, 0xbf, 0x9e, 0x96, 0xdf, 0x2c, 0x3b, 0xb3, 0x43, 0xb5, 0x07, 0xaf, 0xe0, 0xde,
	0xb1, 0x0d, 0x56, 0xe4, 0x71, 0x74, 0x9f, 0x54, 0x2a, 0xfc, 0x2d, 0xf0, 0x41, 0xdc, 0x52, 0x1e,
	0xec, 0x0f, 0x8c, 0xcc, 0xee, 0xb3, 0x87, 0x6b, 0x95, 0x40, 0x48, 0x2a, 0xb5, 0x73, 0x1f, 0xfa,
	0x17, 0xf2, 0x32, 0xea, 0x67, 0x14, 0x77, 0x54, 0x0c, 0xac, 0xcd, 0x38, 0xc9, 0xab, 0xb8, 0xd8,
	0xa9, 0x7b, 0xa1, 0x4a, 0xc1, 0x3b, 0x6b, 0x3c, 0x44, 0x4b, 0xd5, 0x64, 0x7d, 0x30, 0xdb, 0x79,
	0xb2, 0x3d, 0x78, 0xcb, 0x03, 0xb3, 0xe5, 0xdd, 0x37, 0xc1, 0xe3, 0x3b, 0xd5, 0xdf, 0xcb, 0x00,
	0xfe, 0x02, 0xaf, 0xf3, 0x08, 0xf1, 0x4c, 0xcd, 0x1b, 0x43, 0x07, 0x28, 0xac, 0x33, 0xfe, 0x5f,
	0x9d, 0x13, 0x37, 0xa9, 0x42, 0xbf, 0x3f, 0xde, 0x36, 0x07, 0xa7, 0xd4, 0xee, 0x62, 0x28, 0xa2,
	0xf2, 0x74, 0x0a, 0x00, 0x00, 0xff, 0xff, 0x57, 0xb1, 0x5c, 0xe7, 0xf4, 0x00, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// TuberServiceClient is the client API for TuberService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type TuberServiceClient interface {
	CreateReviewApp(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error)
}

type tuberServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewTuberServiceClient(cc grpc.ClientConnInterface) TuberServiceClient {
	return &tuberServiceClient{cc}
}

func (c *tuberServiceClient) CreateReviewApp(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/proto.TuberService/CreateReviewApp", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TuberServiceServer is the server API for TuberService service.
type TuberServiceServer interface {
	CreateReviewApp(context.Context, *Request) (*Response, error)
}

// UnimplementedTuberServiceServer can be embedded to have forward compatible implementations.
type UnimplementedTuberServiceServer struct {
}

func (*UnimplementedTuberServiceServer) CreateReviewApp(ctx context.Context, req *Request) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateReviewApp not implemented")
}

func RegisterTuberServiceServer(s *grpc.Server, srv TuberServiceServer) {
	s.RegisterService(&_TuberService_serviceDesc, srv)
}

func _TuberService_CreateReviewApp_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TuberServiceServer).CreateReviewApp(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.TuberService/CreateReviewApp",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TuberServiceServer).CreateReviewApp(ctx, req.(*Request))
	}
	return interceptor(ctx, in, info, handler)
}

var _TuberService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proto.TuberService",
	HandlerType: (*TuberServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateReviewApp",
			Handler:    _TuberService_CreateReviewApp_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "server.proto",
}