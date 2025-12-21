package node

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/AuraReaper/strangedb/internal/config"
	"github.com/AuraReaper/strangedb/internal/coordinator"
	"github.com/AuraReaper/strangedb/internal/gossip"
	"github.com/AuraReaper/strangedb/internal/hlc"
	"github.com/AuraReaper/strangedb/internal/ring"
	"github.com/AuraReaper/strangedb/internal/storage"
	"github.com/AuraReaper/strangedb/internal/transport/grpc"
	grpcTransport "github.com/AuraReaper/strangedb/internal/transport/grpc"
	httpTransport "github.com/AuraReaper/strangedb/internal/transport/http"
	"github.com/rs/zerolog/log"
)

type Node struct {
	cfg         *config.Config
	storage     storage.Storage
	clock       *hlc.Clock
	ring        *ring.ConsistentHashRing
	gossiper    *gossip.Gossiper
	coordinator *coordinator.Coordinator
	grpcServer  *grpc.Server
	grpcClient  *grpc.Client
	httpServer  *httpTransport.Server
}

func New(cfg *config.Config) (*Node, error) {
	store := storage.NewBadgerStorage(cfg.DataDir)
	clock := hlc.NewClock(cfg.NodeID)

	hashring := ring.New(cfg.VNodes)
	nodeURL := fmt.Sprintf("localhost:%d", cfg.GRPCPort)
	hashring.AddNode(nodeURL)

	for _, seed := range cfg.Seeds {
		if seed != "" && seed != nodeURL {
			hashring.AddNode(seed)
		}
	}

	grpcClient := grpcTransport.NewClient()
	gossiper := gossip.New(nodeURL, cfg.Seeds, cfg.GossipInterval)

	gossiper.SetMembershipChangeCallback(func(members []string) {
		for _, member := range members {
			hashring.AddNode(member)
		}
	})

	coordLogger := log.With().Str("component", "coordinator").Logger()
	coord := coordinator.New(
		nodeURL,
		hashring,
		store,
		clock,
		grpcClient,
		cfg.ReplicationN,
		cfg.ReadQuorum,
		cfg.WriteQuorum,
		coordLogger,
	)

	grpcServer := grpcTransport.NewServer(cfg.GRPCPort, store, clock)
	handler := httpTransport.NewHandler(coord, clock, cfg.NodeID, gossiper, hashring)
	httpServer := httpTransport.NewServer(handler, cfg.HTTPPort)

	return &Node{
		cfg:         cfg,
		storage:     store,
		clock:       clock,
		ring:        hashring,
		gossiper:    gossiper,
		coordinator: coord,
		grpcServer:  grpcServer,
		grpcClient:  grpcClient,
		httpServer:  httpServer,
	}, nil
}

func (n *Node) Start(ctx context.Context) error {
	if err := n.storage.Open(); err != nil {
		return fmt.Errorf("failed to open storage: %w", err)
	}

	go func() {
		fmt.Printf("Starting gRPC server on port %d\n", n.cfg.GRPCPort)
		if err := n.grpcServer.Start(); err != nil {
			fmt.Printf("gRPC server error: %v\n", err)
		}
	}()

	n.gossiper.Start()
	fmt.Println("Gossiper started")

	errCh := make(chan error, 1)
	go func() {
		errCh <- n.httpServer.Start()
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return n.Shutdown()
	}
}

func (n *Node) Shutdown() error {
	n.gossiper.Stop()
	n.grpcServer.Stop()
	n.grpcClient.Close()

	if err := n.httpServer.Shutdown(); err != nil {
		return err
	}

	return n.storage.Close()
}

// starts node with graceful shutdown
func Run(cfg *config.Config) error {
	node, err := New(cfg)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		cancel()
	}()

	return node.Start(ctx)
}
