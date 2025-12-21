package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/AuraReaper/strangedb/internal/hlc"
	"github.com/AuraReaper/strangedb/internal/storage"
	pb "github.com/AuraReaper/strangedb/internal/transport/grpc/proto"
	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedNodeServiceServer
	storage storage.Storage
	clock   *hlc.Clock
	server  *grpc.Server
	port    int
}

func NewServer(port int, storage storage.Storage, clock *hlc.Clock) *Server {
	return &Server{
		storage: storage,
		clock:   clock,
		port:    port,
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return err
	}

	s.server = grpc.NewServer()
	pb.RegisterNodeServiceServer(s.server, s)

	return s.server.Serve(listener)
}

func (s *Server) Stop() {
	if s.server != nil {
		s.server.GracefulStop()
	}
}

func (s *Server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	record, err := s.storage.Get(req.Key)
	if err == storage.ErrKeyNotFound || err == storage.ErrKeyDeleted {
		return &pb.GetResponse{
			Found: false,
		}, nil
	}
	if err != nil {
		return nil, err
	}

	return &pb.GetResponse{
		Found: true,
		Record: &pb.Record{
			Key:   record.Key,
			Value: record.Value,
			Timestamp: &pb.Timestamp{
				WallTime: record.Timestamp.WallTime,
				Logical:  record.Timestamp.Logical,
				NodeId:   record.Timestamp.NodeID,
			},
			Tombstone: record.Tombstone,
		},
	}, nil
}

func (s *Server) Set(ctx context.Context, req *pb.SetRequest) (*pb.SetResponse, error) {
	record := &storage.Record{
		Key:   req.Record.Key,
		Value: req.Record.Value,
		Timestamp: hlc.Timestamp{
			WallTime: req.Record.Timestamp.WallTime,
			Logical:  req.Record.Timestamp.Logical,
			NodeID:   req.Record.Timestamp.NodeId,
		},
		Tombstone: req.Record.Tombstone,
	}

	if err := s.storage.Set(record); err != nil {
		return nil, err
	}

	return &pb.SetResponse{
		Success:   true,
		Timestamp: req.Record.Timestamp,
	}, nil
}

func (s *Server) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	ts := hlc.Timestamp{
		WallTime: req.Timestamp.WallTime,
		Logical:  req.Timestamp.Logical,
		NodeID:   req.Timestamp.NodeId,
	}

	if err := s.storage.Delete(req.Key, ts); err != nil {
		return nil, err
	}

	return &pb.DeleteResponse{
		Success: true,
	}, nil
}
