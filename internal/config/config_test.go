package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_Success(t *testing.T) {
	yaml := `
marketstack:
  key: "test-ms-key"
  stocks:
    - AAPL
    - MSFT

fixer:
  key: "test-fixer-key"
  pairs:
    - { from: "EUR", to: "USD" }
    - { from: "GBP", to: "USD" }

ledger:
  price_db: "/tmp/prices.db"
`
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0o600); err != nil {
		t.Fatalf("could not create temp config: %v", err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// marketstack
	if cfg.Marketstack.Key != "test-ms-key" {
		t.Errorf("expected Marketstack.Key 'test-ms-key', got %q", cfg.Marketstack.Key)
	}
	if len(cfg.Marketstack.Stocks) != 2 || cfg.Marketstack.Stocks[0] != "AAPL" || cfg.Marketstack.Stocks[1] != "MSFT" {
		t.Errorf("unexpected Marketstack.Stocks: %v", cfg.Marketstack.Stocks)
	}

	// fixer
	if cfg.Fixer.Key != "test-fixer-key" {
		t.Errorf("expected Fixer.Key 'test-fixer-key', got %q", cfg.Fixer.Key)
	}
	if len(cfg.Fixer.Pairs) != 2 {
		t.Errorf("expected 2 currency pairs, got %d", len(cfg.Fixer.Pairs))
	}
	if cfg.Fixer.Pairs[0] != (Pair{From: "EUR", To: "USD"}) || cfg.Fixer.Pairs[1] != (Pair{From: "GBP", To: "USD"}) {
		t.Errorf("unexpected Fixer.Pairs: %v", cfg.Fixer.Pairs)
	}

	if cfg.Ledger.PriceDB != "/tmp/prices.db" {
		t.Errorf("expected Ledger.PriceDB '/tmp/prices.db', got %q", cfg.Ledger.PriceDB)
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("/non/existent/config.yaml")
	if err == nil {
		t.Error("expected an error for non-existent file, got nil")
	}
}

func TestLoadConfig_MalformedYAML(t *testing.T) {
	yaml := `
marketstack:
  key: "test"
  stocks: [AAPL
ledger:
  price_db: "/tmp/prices.db"
`
	path := filepath.Join(t.TempDir(), "malformed.yaml")
	if err := os.WriteFile(path, []byte(yaml), 0o600); err != nil {
		t.Fatalf("could not create temp config: %v", err)
	}

	_, err := LoadConfig(path)
	if err == nil {
		t.Error("expected YAML unmarshal error, got nil")
	}
}
