package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

func writeTestEnvFile(t *testing.T, contents string) string {
	t.Helper()
	dir := t.TempDir()
	testDirPath := filepath.Join(dir, "config.env")
	if err := os.WriteFile(testDirPath, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}

	return testDirPath
}

func TestLoad_InfuraURLAssembly(t *testing.T) {
	tests := []struct {
		name   string
		rpc    string
		key    string
		expect string
	}{
		{"base-no-v3", "https://sepolia.infura.io", "abc", "https://sepolia.infura.io/v3/abc"},
		{"base-with-v3", "https://sepolia.infura.io/v3", "abc", "https://sepolia.infura.io/v3/abc"},
		{"already-has-key", "https://sepolia.infura.io/v3/abc", "abc", "https://sepolia.infura.io/v3/abc"},
		{"trailing-slash", "https://sepolia.infura.io/", "abc", "https://sepolia.infura.io/v3/abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := `
HTTP_ADDR=:8080
RPC=` + tt.rpc + `
KEY=` + tt.key + `
USDC_ADDRESS=0x1c7D4B196Cb0C7B01d743Fbc6116a902379C7238
VERIFY_TIMEOUT=3s
SETTLE_TIMEOUT=10s
`
			path := writeTestEnvFile(t, env)
			cfg, err := Load(path)
			if err != nil {
				t.Fatalf("Load error: %v", err)
			}
			if cfg.RPCURL != tt.expect {
				t.Fatalf("RPCURL = %q, want %q", cfg.RPCURL, tt.expect)
			}
		})
	}
}

func TestLoad_RequiredFields(t *testing.T) {
	// Missing KEY
	path := writeTestEnvFile(t, `
HTTP_ADDR=:8080
RPC=https://sepolia.infura.io
USDC_ADDRESS=0x1c7D4B196Cb0C7B01d743Fbc6116a902379C7238
`)
	if _, err := Load(path); err == nil {
		t.Fatal("expected error for missing KEY")
	}

	// Missing RPC
	path = writeTestEnvFile(t, `
HTTP_ADDR=:8080
KEY=abc
USDC_ADDRESS=0x1c7D4B196Cb0C7B01d743Fbc6116a902379C7238
`)
	if _, err := Load(path); err == nil {
		t.Fatal("expected error for missing RPC")
	}
}

func TestLoad_DurationsAndAddress(t *testing.T) {
	path := writeTestEnvFile(t, `
HTTP_ADDR=:9090
RPC=https://sepolia.infura.io
KEY=abc
USDC_ADDRESS=0x1c7D4B196Cb0C7B01d743Fbc6116a902379C7238
VERIFY_TIMEOUT=250ms
SETTLE_TIMEOUT=1m
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.HTTPAddr != ":9090" {
		t.Fatalf("HTTPAddr = %q, want :9090", cfg.HTTPAddr)
	}
	if cfg.VerifyTimeout != 250*time.Millisecond {
		t.Fatalf("VerifyTimeout = %v, want 250ms", cfg.VerifyTimeout)
	}
	if cfg.SettleTimeout != time.Minute {
		t.Fatalf("SettleTimeout = %v, want 1m", cfg.SettleTimeout)
	}
	if !common.IsHexAddress(cfg.USDCAddress) {
		t.Fatalf("USDC address not hex: %q", cfg.USDCAddress)
	}
}

func TestParseEnvFile_CommentsAndMalformed(t *testing.T) {
	// Comments & blanks should be ignored
	path := writeTestEnvFile(t, `
# comment
RPC = https://sepolia.infura.io
KEY = abc

# comment 2
USDC_ADDRESS = 0x1c7D4B196Cb0C7B01d743Fbc6116a902379C7238
`)
	m, err := parseEnvFile(path)
	if err != nil {
		t.Fatalf("parseEnvFile error: %v", err)
	}
	if got := strings.TrimSpace(m["RPC"]); got != "https://sepolia.infura.io" {
		t.Fatalf("RPC = %q", got)
	}
	if _, ok := m["# comment"]; ok {
		t.Fatal("parsed comment line unexpectedly")
	}

	// Malformed line → error
	path = writeTestEnvFile(t, `
RPC https://sepolia.infura.io
`)
	if _, err := parseEnvFile(path); err == nil {
		t.Fatal("expected error for malformed line")
	}
}

func TestLoad_DefaultsOnInvalidDuration(t *testing.T) {
	// verify invalid durations fall back to defaults (3s / 10s)
	path := writeTestEnvFile(t, `
HTTP_ADDR=:8080
RPC=https://sepolia.infura.io
KEY=abc
USDC_ADDRESS=0x1c7D4B196Cb0C7B01d743Fbc6116a902379C7238
VERIFY_TIMEOUT=not-a-duration
SETTLE_TIMEOUT=also-bad
`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.VerifyTimeout != 3*time.Second {
		t.Fatalf("VerifyTimeout = %v, want 3s default", cfg.VerifyTimeout)
	}
	if cfg.SettleTimeout != 10*time.Second {
		t.Fatalf("SettleTimeout = %v, want 10s default", cfg.SettleTimeout)
	}
}
