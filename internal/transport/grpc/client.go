package grpc

import (
	"context"
	"sync"
	"time"

	pb "github.com/AuraReaper/strangedb/internal/transport/grpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	mu    sync.RWMutex
	conns map[string]*grpc.ClientConn
}

func NewClient() *Client {
	return &Client{
		conns: make(map[string]*grpc.ClientConn),
	}
}

func (c *Client) getConn(address string) (*grpc.ClientConn, error) {
	c.mu.RLock()
	conn, ok := c.conns[address]
	c.mu.RUnlock()

	if ok {
		return conn, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if conn, ok := c.conns[address]; ok {
		return conn, nil
	}

	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(16*1024*1024)),
	)
	if err != nil {
		return nil, err
	}

	c.conns[address] = conn
	return conn, nil
}

func (c *Client) Get(ctx context.Context, address string, key string) (*pb.GetResponse, error) {
	conn, err := c.getConn(address)
	if err != nil {
		return nil, err
	}

	client := pb.NewNodeServiceClient(conn)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return client.Get(ctx, &pb.GetRequest{
		Key: key,
	})
}

func (c *Client) Set(ctx context.Context, address string, record *pb.Record) (*pb.SetResponse, error) {
	conn, err := c.getConn(address)
	if err != nil {
		return nil, err
	}

	client := pb.NewNodeServiceClient(conn)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return client.Set(ctx, &pb.SetRequest{
		Record: record,
	})
}

func (c *Client) Delete(ctx context.Context, address string, key string, timestamp *pb.Timestamp) (*pb.DeleteResponse, error) {
	conn, err := c.getConn(address)
	if err != nil {
		return nil, err
	}

	client := pb.NewNodeServiceClient(conn)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return client.Delete(ctx, &pb.DeleteRequest{
		Key:       key,
		Timestamp: timestamp,
	})
}

func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, conn := range c.conns {
		conn.Close()
	}

	c.conns = make(map[string]*grpc.ClientConn)
}
