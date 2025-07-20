package server

import (
	"context"
	"time"

	pb "github.com/Unfield/AmpKV/pkg/client/rpc"
	"github.com/Unfield/AmpKV/pkg/embedded"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AmpKVGrpcServer struct {
	pb.UnimplementedAmpKVServiceServer
	store embedded.AmpKV
}

func NewAmpKVGrpcServer(store embedded.AmpKV) *AmpKVGrpcServer {
	return &AmpKVGrpcServer{
		store: store,
	}
}

func (s *AmpKVGrpcServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	if req.Key == "" {
		return nil, status.Errorf(codes.InvalidArgument, "GetRequest: key must not be empty")
	}

	val, found := s.store.Get(req.Key)
	if !found {
		return &pb.GetResponse{
			Found: false,
			Kv:    nil,
		}, nil
	}

	return &pb.GetResponse{
		Found: true,
		Kv: &pb.KeyValue{
			Key:   req.Key,
			Value: val,
		},
	}, nil
}

func (s *AmpKVGrpcServer) Set(ctx context.Context, req *pb.SetRequest) (*pb.OperationResponse, error) {
	if req.Kv == nil || req.Kv.Key == "" || req.Kv.Value == nil {
		return nil, status.Errorf(codes.InvalidArgument, "SetRequest: key, value and kv fields must be provided")
	}

	if req.Kv.Cost <= 0 {
		req.Kv.Cost = 1
	}

	err := s.store.Set(req.Kv.Key, req.Kv.Value, req.Kv.Cost)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to set key in store: %v", err)
	}

	return &pb.OperationResponse{
		Success: true,
		Message: "Key set successfully",
	}, nil
}

func (s *AmpKVGrpcServer) SetWithTTL(ctx context.Context, req *pb.SetWithTTLRequest) (*pb.OperationResponse, error) {
	if req.Kv == nil || req.Kv.Key == "" || req.Kv.Value == nil {
		return nil, status.Errorf(codes.InvalidArgument, "SetRequest: key, value and kv fields must be provided")
	}
	if req.TtlSeconds <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "SetWithTTLRequest: TTL in seconds must be positive")
	}

	ttl := time.Duration(req.TtlSeconds) * time.Second

	if req.Kv.Cost <= 0 {
		req.Kv.Cost = 1
	}

	err := s.store.SetWithTTL(req.Kv.Key, req.Kv.Value, req.Kv.Cost, ttl)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to set key in store: %v", err)
	}

	return &pb.OperationResponse{
		Success: true,
		Message: "Key with ttl set successfully",
	}, nil
}

func (s *AmpKVGrpcServer) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.OperationResponse, error) {
	if req.Key == "" {
		return nil, status.Errorf(codes.InvalidArgument, "DeleteRequest: key must be provided")
	}

	s.store.Delete(req.Key)

	return &pb.OperationResponse{
		Success: true,
		Message: "Key deleted successfully",
	}, nil
}
