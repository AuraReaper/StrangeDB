package main

import (
	"fmt"
	"os"

	"github.com/AuraReaper/strangedb/internal/config"
	"github.com/AuraReaper/strangedb/internal/node"
)

func main() {
	cfg := config.Load()

	fmt.Printf("Starting StrangeDB node %s on port %d\n", cfg.NodeID, cfg.HTTPPort)

	if err := node.Run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
