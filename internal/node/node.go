package node

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/AuraReaper/strangedb/internal/config"
	"github.com/AuraReaper/strangedb/internal/hlc"
	"github.com/AuraReaper/strangedb/internal/storage"
	httpTransport "github.com/AuraReaper/strangedb/internal/transport/http"
)

type Node struct {
	cfg     *config.Config
	storage storage.Storage
	clock   *hlc.Clock
	server  *httpTransport.Server
}

func New(cfg *config.Config) (*Node, error) {
	store := storage.NewBadgerStorage(cfg.DataDir)
	clock := hlc.NewClock(cfg.NodeID)
	handler := httpTransport.NewHandler(store, clock, cfg.NodeID)
	server := httpTransport.NewServer(handler, cfg.HTTPPort)

	return &Node{
		cfg:     cfg,
		storage: store,
		clock:   clock,
		server:  server,
	}, nil
}

func (n *Node) Start(ctx context.Context) error {
	if err := n.storage.Open(); err != nil {
		return err
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- n.server.Start()
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return n.server.Shutdown()
	}
}

func (n *Node) Shutdown() error {
	if err := n.server.Shutdown(); err != nil {
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
