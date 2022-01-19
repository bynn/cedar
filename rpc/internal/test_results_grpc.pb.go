// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.19.3
// source: test_results.proto

package internal

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// CedarTestResultsClient is the client API for CedarTestResults service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CedarTestResultsClient interface {
	CreateTestResultsRecord(ctx context.Context, in *TestResultsInfo, opts ...grpc.CallOption) (*TestResultsResponse, error)
	AddTestResults(ctx context.Context, in *TestResults, opts ...grpc.CallOption) (*TestResultsResponse, error)
	StreamTestResults(ctx context.Context, opts ...grpc.CallOption) (CedarTestResults_StreamTestResultsClient, error)
	CloseTestResultsRecord(ctx context.Context, in *TestResultsEndInfo, opts ...grpc.CallOption) (*TestResultsResponse, error)
}

type cedarTestResultsClient struct {
	cc grpc.ClientConnInterface
}

func NewCedarTestResultsClient(cc grpc.ClientConnInterface) CedarTestResultsClient {
	return &cedarTestResultsClient{cc}
}

func (c *cedarTestResultsClient) CreateTestResultsRecord(ctx context.Context, in *TestResultsInfo, opts ...grpc.CallOption) (*TestResultsResponse, error) {
	out := new(TestResultsResponse)
	err := c.cc.Invoke(ctx, "/cedar.CedarTestResults/CreateTestResultsRecord", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cedarTestResultsClient) AddTestResults(ctx context.Context, in *TestResults, opts ...grpc.CallOption) (*TestResultsResponse, error) {
	out := new(TestResultsResponse)
	err := c.cc.Invoke(ctx, "/cedar.CedarTestResults/AddTestResults", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cedarTestResultsClient) StreamTestResults(ctx context.Context, opts ...grpc.CallOption) (CedarTestResults_StreamTestResultsClient, error) {
	stream, err := c.cc.NewStream(ctx, &CedarTestResults_ServiceDesc.Streams[0], "/cedar.CedarTestResults/StreamTestResults", opts...)
	if err != nil {
		return nil, err
	}
	x := &cedarTestResultsStreamTestResultsClient{stream}
	return x, nil
}

type CedarTestResults_StreamTestResultsClient interface {
	Send(*TestResults) error
	CloseAndRecv() (*TestResultsResponse, error)
	grpc.ClientStream
}

type cedarTestResultsStreamTestResultsClient struct {
	grpc.ClientStream
}

func (x *cedarTestResultsStreamTestResultsClient) Send(m *TestResults) error {
	return x.ClientStream.SendMsg(m)
}

func (x *cedarTestResultsStreamTestResultsClient) CloseAndRecv() (*TestResultsResponse, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(TestResultsResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *cedarTestResultsClient) CloseTestResultsRecord(ctx context.Context, in *TestResultsEndInfo, opts ...grpc.CallOption) (*TestResultsResponse, error) {
	out := new(TestResultsResponse)
	err := c.cc.Invoke(ctx, "/cedar.CedarTestResults/CloseTestResultsRecord", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CedarTestResultsServer is the server API for CedarTestResults service.
// All implementations must embed UnimplementedCedarTestResultsServer
// for forward compatibility
type CedarTestResultsServer interface {
	CreateTestResultsRecord(context.Context, *TestResultsInfo) (*TestResultsResponse, error)
	AddTestResults(context.Context, *TestResults) (*TestResultsResponse, error)
	StreamTestResults(CedarTestResults_StreamTestResultsServer) error
	CloseTestResultsRecord(context.Context, *TestResultsEndInfo) (*TestResultsResponse, error)
	mustEmbedUnimplementedCedarTestResultsServer()
}

// UnimplementedCedarTestResultsServer must be embedded to have forward compatible implementations.
type UnimplementedCedarTestResultsServer struct {
}

func (UnimplementedCedarTestResultsServer) CreateTestResultsRecord(context.Context, *TestResultsInfo) (*TestResultsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateTestResultsRecord not implemented")
}
func (UnimplementedCedarTestResultsServer) AddTestResults(context.Context, *TestResults) (*TestResultsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddTestResults not implemented")
}
func (UnimplementedCedarTestResultsServer) StreamTestResults(CedarTestResults_StreamTestResultsServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamTestResults not implemented")
}
func (UnimplementedCedarTestResultsServer) CloseTestResultsRecord(context.Context, *TestResultsEndInfo) (*TestResultsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CloseTestResultsRecord not implemented")
}
func (UnimplementedCedarTestResultsServer) mustEmbedUnimplementedCedarTestResultsServer() {}

// UnsafeCedarTestResultsServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CedarTestResultsServer will
// result in compilation errors.
type UnsafeCedarTestResultsServer interface {
	mustEmbedUnimplementedCedarTestResultsServer()
}

func RegisterCedarTestResultsServer(s grpc.ServiceRegistrar, srv CedarTestResultsServer) {
	s.RegisterService(&CedarTestResults_ServiceDesc, srv)
}

func _CedarTestResults_CreateTestResultsRecord_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TestResultsInfo)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CedarTestResultsServer).CreateTestResultsRecord(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cedar.CedarTestResults/CreateTestResultsRecord",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CedarTestResultsServer).CreateTestResultsRecord(ctx, req.(*TestResultsInfo))
	}
	return interceptor(ctx, in, info, handler)
}

func _CedarTestResults_AddTestResults_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TestResults)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CedarTestResultsServer).AddTestResults(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cedar.CedarTestResults/AddTestResults",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CedarTestResultsServer).AddTestResults(ctx, req.(*TestResults))
	}
	return interceptor(ctx, in, info, handler)
}

func _CedarTestResults_StreamTestResults_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(CedarTestResultsServer).StreamTestResults(&cedarTestResultsStreamTestResultsServer{stream})
}

type CedarTestResults_StreamTestResultsServer interface {
	SendAndClose(*TestResultsResponse) error
	Recv() (*TestResults, error)
	grpc.ServerStream
}

type cedarTestResultsStreamTestResultsServer struct {
	grpc.ServerStream
}

func (x *cedarTestResultsStreamTestResultsServer) SendAndClose(m *TestResultsResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *cedarTestResultsStreamTestResultsServer) Recv() (*TestResults, error) {
	m := new(TestResults)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _CedarTestResults_CloseTestResultsRecord_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TestResultsEndInfo)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CedarTestResultsServer).CloseTestResultsRecord(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cedar.CedarTestResults/CloseTestResultsRecord",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CedarTestResultsServer).CloseTestResultsRecord(ctx, req.(*TestResultsEndInfo))
	}
	return interceptor(ctx, in, info, handler)
}

// CedarTestResults_ServiceDesc is the grpc.ServiceDesc for CedarTestResults service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var CedarTestResults_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "cedar.CedarTestResults",
	HandlerType: (*CedarTestResultsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateTestResultsRecord",
			Handler:    _CedarTestResults_CreateTestResultsRecord_Handler,
		},
		{
			MethodName: "AddTestResults",
			Handler:    _CedarTestResults_AddTestResults_Handler,
		},
		{
			MethodName: "CloseTestResultsRecord",
			Handler:    _CedarTestResults_CloseTestResultsRecord_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamTestResults",
			Handler:       _CedarTestResults_StreamTestResults_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "test_results.proto",
}
