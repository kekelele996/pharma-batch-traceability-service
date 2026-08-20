package config

import (
	"testing"
	"time"
)

// TestConfigBulkTimeoutP14 BULK_VERIFY_TIMEOUT 环境变量必须真实生效。
func TestConfigBulkTimeoutR015D(t *testing.T) {
	t.Setenv("BULK_VERIFY_TIMEOUT", "5s")
	cfg := Load()
	if cfg.BulkTimeout != 5*time.Second {
		t.Fatalf("expected 5s bulk timeout, got %v", cfg.BulkTimeout)
	}
}
