package config

import (
	"flag"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	// node
	NodeID string

	// server
	HTTPPort int
	GRPCPort int

	// storage
	DataDir string

	// cluster settings
	Seeds        []string
	ReplicationN int
	ReadQuorum   int
	WriteQuorum  int
	VNodes       int // virtual nodes

	// timing settings
	GossipInterval      time.Duration
	AntiEntropyInterval time.Duration
	TombstoneTTL        time.Duration

	// logging
	LogLevel string
}

func DefaultConfig() *Config {
	return &Config{
		NodeID:              generateNodeID(),
		HTTPPort:            9000,
		GRPCPort:            9001,
		DataDir:             "./data",
		Seeds:               []string{},
		ReplicationN:        3,
		ReadQuorum:          2,
		WriteQuorum:         2,
		VNodes:              150,
		GossipInterval:      time.Second,
		AntiEntropyInterval: 10 * time.Minute,
		TombstoneTTL:        24 * time.Hour,
		LogLevel:            "info",
	}
}

// loads configurations from env and flags
func Load() *Config {
	cfg := DefaultConfig()

	cfg.loadFromEnv()

	cfg.loadFromFlags()

	return cfg
}

func (c *Config) loadFromEnv() {
	if v := os.Getenv("NODE_ID"); v != "" {
		c.NodeID = v
	}

	if v := os.Getenv("HTTP_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			c.HTTPPort = port
		}
	}

	if v := os.Getenv("GRPC_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			c.GRPCPort = port
		}
	}

	if v := os.Getenv("DATA_DIR"); v != "" {
		c.DataDir = v
	}

	if v := os.Getenv("REPLICATION_N"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.ReplicationN = n
		}
	}

	if v := os.Getenv("READ_QUORUM"); v != "" {
		if r, err := strconv.Atoi(v); err == nil {
			c.ReadQuorum = r
		}
	}

	if v := os.Getenv("WRITE_QUORUM"); v != "" {
		if w, err := strconv.Atoi(v); err == nil {
			c.WriteQuorum = w
		}
	}
	if v := os.Getenv("VIRTUAL_NODES"); v != "" {
		if vn, err := strconv.Atoi(v); err == nil {
			c.VNodes = vn
		}
	}

	if v := os.Getenv("LOG_LEVEL"); v != "" {
		c.LogLevel = v
	}
}

func (c *Config) loadFromFlags() {
	flag.StringVar(&c.NodeID, "node-id", c.NodeID, "Unique node identifier")
	flag.IntVar(&c.HTTPPort, "http-port", c.HTTPPort, "HTTP API port")
	flag.IntVar(&c.GRPCPort, "grpc-port", c.GRPCPort, "gRPC inter-node port")
	flag.StringVar(&c.DataDir, "data-dir", c.DataDir, "Data directory")
	flag.IntVar(&c.ReplicationN, "n", c.ReplicationN, "repliaction factor, N")
	flag.IntVar(&c.ReadQuorum, "r", c.ReadQuorum, "read quorum")
	flag.IntVar(&c.WriteQuorum, "w", c.WriteQuorum, "write quorum")
	flag.IntVar(&c.VNodes, "v-nodes", c.VNodes, "virtual nodes")
	flag.StringVar(&c.LogLevel, "log-level", c.LogLevel, "Log level (debug/info/warn/error)")

	var seeds string
	flag.StringVar(&seeds, "seeds", "", "comma seperated seed node urls")

	flag.Parse()

	if seeds != "" {
		c.Seeds = strings.Split(seeds, ",")
	}
}

func generateNodeID() string {
	hostname, _ := os.Hostname()
	return hostname + "-" + strconv.FormatInt(time.Now().UnixNano()%10000, 10)
}
