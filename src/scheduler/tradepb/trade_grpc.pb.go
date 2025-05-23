// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.0
// source: tradepb/trade.proto

package tradepb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	TradeService_SendTrade_FullMethodName = "/trade.TradeService/SendTrade"
)

// TradeServiceClient is the client API for TradeService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
//
// The TradeService definition
type TradeServiceClient interface {
	SendTrade(ctx context.Context, in *Trade, opts ...grpc.CallOption) (*TradeResponse, error)
}

type tradeServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewTradeServiceClient(cc grpc.ClientConnInterface) TradeServiceClient {
	return &tradeServiceClient{cc}
}

func (c *tradeServiceClient) SendTrade(ctx context.Context, in *Trade, opts ...grpc.CallOption) (*TradeResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(TradeResponse)
	err := c.cc.Invoke(ctx, TradeService_SendTrade_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TradeServiceServer is the server API for TradeService service.
// All implementations must embed UnimplementedTradeServiceServer
// for forward compatibility.
//
// The TradeService definition
type TradeServiceServer interface {
	SendTrade(context.Context, *Trade) (*TradeResponse, error)
	mustEmbedUnimplementedTradeServiceServer()
}

// UnimplementedTradeServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedTradeServiceServer struct{}

func (UnimplementedTradeServiceServer) SendTrade(context.Context, *Trade) (*TradeResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendTrade not implemented")
}
func (UnimplementedTradeServiceServer) mustEmbedUnimplementedTradeServiceServer() {}
func (UnimplementedTradeServiceServer) testEmbeddedByValue()                      {}

// UnsafeTradeServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TradeServiceServer will
// result in compilation errors.
type UnsafeTradeServiceServer interface {
	mustEmbedUnimplementedTradeServiceServer()
}

func RegisterTradeServiceServer(s grpc.ServiceRegistrar, srv TradeServiceServer) {
	// If the following call pancis, it indicates UnimplementedTradeServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&TradeService_ServiceDesc, srv)
}

func _TradeService_SendTrade_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Trade)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TradeServiceServer).SendTrade(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TradeService_SendTrade_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TradeServiceServer).SendTrade(ctx, req.(*Trade))
	}
	return interceptor(ctx, in, info, handler)
}

// TradeService_ServiceDesc is the grpc.ServiceDesc for TradeService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var TradeService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "trade.TradeService",
	HandlerType: (*TradeServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendTrade",
			Handler:    _TradeService_SendTrade_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "tradepb/trade.proto",
}
