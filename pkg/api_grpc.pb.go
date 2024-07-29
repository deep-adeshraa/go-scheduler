// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v5.26.1
// source: api.proto

package pkg

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

const (
	ScheduleJobService_GetUpcomingJobs_FullMethodName = "/pkg.ScheduleJobService/GetUpcomingJobs"
)

// ScheduleJobServiceClient is the client API for ScheduleJobService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ScheduleJobServiceClient interface {
	// rpc ScheduleJob(google.protobuf.Empty) returns (stream UpcomingJobs);
	GetUpcomingJobs(ctx context.Context, in *Worker, opts ...grpc.CallOption) (ScheduleJobService_GetUpcomingJobsClient, error)
}

type scheduleJobServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewScheduleJobServiceClient(cc grpc.ClientConnInterface) ScheduleJobServiceClient {
	return &scheduleJobServiceClient{cc}
}

func (c *scheduleJobServiceClient) GetUpcomingJobs(ctx context.Context, in *Worker, opts ...grpc.CallOption) (ScheduleJobService_GetUpcomingJobsClient, error) {
	stream, err := c.cc.NewStream(ctx, &ScheduleJobService_ServiceDesc.Streams[0], ScheduleJobService_GetUpcomingJobs_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &scheduleJobServiceGetUpcomingJobsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type ScheduleJobService_GetUpcomingJobsClient interface {
	Recv() (*UpcomingJobs, error)
	grpc.ClientStream
}

type scheduleJobServiceGetUpcomingJobsClient struct {
	grpc.ClientStream
}

func (x *scheduleJobServiceGetUpcomingJobsClient) Recv() (*UpcomingJobs, error) {
	m := new(UpcomingJobs)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// ScheduleJobServiceServer is the server API for ScheduleJobService service.
// All implementations must embed UnimplementedScheduleJobServiceServer
// for forward compatibility
type ScheduleJobServiceServer interface {
	// rpc ScheduleJob(google.protobuf.Empty) returns (stream UpcomingJobs);
	GetUpcomingJobs(*Worker, ScheduleJobService_GetUpcomingJobsServer) error
	mustEmbedUnimplementedScheduleJobServiceServer()
}

// UnimplementedScheduleJobServiceServer must be embedded to have forward compatible implementations.
type UnimplementedScheduleJobServiceServer struct {
}

func (UnimplementedScheduleJobServiceServer) GetUpcomingJobs(*Worker, ScheduleJobService_GetUpcomingJobsServer) error {
	return status.Errorf(codes.Unimplemented, "method GetUpcomingJobs not implemented")
}
func (UnimplementedScheduleJobServiceServer) mustEmbedUnimplementedScheduleJobServiceServer() {}

// UnsafeScheduleJobServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ScheduleJobServiceServer will
// result in compilation errors.
type UnsafeScheduleJobServiceServer interface {
	mustEmbedUnimplementedScheduleJobServiceServer()
}

func RegisterScheduleJobServiceServer(s grpc.ServiceRegistrar, srv ScheduleJobServiceServer) {
	s.RegisterService(&ScheduleJobService_ServiceDesc, srv)
}

func _ScheduleJobService_GetUpcomingJobs_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Worker)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(ScheduleJobServiceServer).GetUpcomingJobs(m, &scheduleJobServiceGetUpcomingJobsServer{stream})
}

type ScheduleJobService_GetUpcomingJobsServer interface {
	Send(*UpcomingJobs) error
	grpc.ServerStream
}

type scheduleJobServiceGetUpcomingJobsServer struct {
	grpc.ServerStream
}

func (x *scheduleJobServiceGetUpcomingJobsServer) Send(m *UpcomingJobs) error {
	return x.ServerStream.SendMsg(m)
}

// ScheduleJobService_ServiceDesc is the grpc.ServiceDesc for ScheduleJobService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ScheduleJobService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "pkg.ScheduleJobService",
	HandlerType: (*ScheduleJobServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GetUpcomingJobs",
			Handler:       _ScheduleJobService_GetUpcomingJobs_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "api.proto",
}