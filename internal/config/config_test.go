package config

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.HTTPPort != 9000 {
		t.Errorf("Expected HTTPPort 9000, got %d", cfg.HTTPPort)
	}

	if cfg.GRPCPort != 9001 {
		t.Errorf("Expected GRPCPort 9001, got %d", cfg.GRPCPort)
	}

	if cfg.ReplicationN != 3 {
		t.Errorf("Expected ReplicationN 3, got %d", cfg.ReplicationN)
	}
}

func TestEnvOverride(t *testing.T) {
	os.Setenv("HTTP_PORT", "8080")
	defer os.Unsetenv("HTTP_PORT")

	cfg := Load()

	if cfg.HTTPPort != 8080 {
		t.Errorf("Expected HTTPPort 8080 from env, got %d", cfg.HTTPPort)
	}
}
