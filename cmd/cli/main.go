package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

var dbProcesses []*exec.Cmd

type NodeConfig struct {
	HTTPPort int
	GRPCPort int
	DataDir  string
	NodeID   string
}

func main() {
	// Flags
	clusterMode := flag.Bool("cluster", true, "Start 3-node cluster (default: true)")
	singleNode := flag.Bool("single", false, "Start single node only")
	noStart := flag.Bool("no-start", false, "Connect to existing cluster, don't start servers")
	adminMode := flag.Bool("admin", false, "Enable admin mode (metrics, diagnostics)")
	urls := flag.String("urls", "", "Comma-separated node URLs (for no-start mode)")
	flag.Parse()

	var nodeURLs []string
	var mode string

	if *noStart {
		// Connect to existing cluster - no splash
		if *urls == "" {
			nodeURLs = []string{"http://localhost:9000", "http://localhost:9010", "http://localhost:9020"}
		} else {
			nodeURLs = splitURLs(*urls)
		}
		mode = "existing"
	} else if *singleNode {
		mode = "single"
		nodeURLs = []string{"http://localhost:9000"}
	} else if *clusterMode {
		mode = "cluster"
		nodeURLs = []string{
			"http://localhost:9000",
			"http://localhost:9010",
			"http://localhost:9020",
		}
	}

	// Show animated splash screen (unless connecting to existing)
	if mode != "existing" {
		// Start nodes in background while showing splash
		if mode == "single" {
			go startSingleNode()
		} else if mode == "cluster" {
			go startCluster()
		}

		splash := NewSplashModel(mode)
		p := tea.NewProgram(splash, tea.WithAltScreen())
		finalModel, err := p.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Check if user quit during splash
		if sm, ok := finalModel.(SplashModel); ok && !sm.IsDone() {
			stopCluster()
			os.Exit(0)
		}

		// Brief wait for nodes to be ready
		time.Sleep(500 * time.Millisecond)
	}

	// Create and run the main TUI
	model := NewModel(nodeURLs, *adminMode)

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		stopCluster()
		os.Exit(1)
	}

	// Cleanup on exit
	stopCluster()
	fmt.Println("👋 Goodbye!")
}

func startSingleNode() error {
	return startNode(NodeConfig{
		HTTPPort: 9000,
		GRPCPort: 9001,
		DataDir:  "./data/cli-node",
		NodeID:   "node-1",
	}, "")
}

func startCluster() error {
	nodes := []NodeConfig{
		{HTTPPort: 9000, GRPCPort: 9001, DataDir: "./data/node1", NodeID: "node-1"},
		{HTTPPort: 9010, GRPCPort: 9011, DataDir: "./data/node2", NodeID: "node-2"},
		{HTTPPort: 9020, GRPCPort: 9021, DataDir: "./data/node3", NodeID: "node-3"},
	}

	seeds := "localhost:9001,localhost:9011,localhost:9021"

	for _, node := range nodes {
		if err := startNode(node, seeds); err != nil {
			return err
		}
		time.Sleep(800 * time.Millisecond)
	}

	return nil
}

func startNode(cfg NodeConfig, seeds string) error {
	binaryPath := "./build/strangedb"
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		binaryPath = "strangedb"
	}

	args := []string{
		"--http-port", fmt.Sprintf("%d", cfg.HTTPPort),
		"--grpc-port", fmt.Sprintf("%d", cfg.GRPCPort),
		"--data-dir", cfg.DataDir,
		"--node-id", cfg.NodeID,
	}

	if seeds != "" {
		args = append(args, "--seeds", seeds)
	}

	cmd := exec.Command(binaryPath, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return err
	}

	dbProcesses = append(dbProcesses, cmd)
	return nil
}

func stopCluster() {
	for _, proc := range dbProcesses {
		if proc != nil && proc.Process != nil {
			proc.Process.Kill()
		}
	}
}

func splitURLs(urls string) []string {
	var result []string
	current := ""
	for _, c := range urls {
		if c == ',' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
