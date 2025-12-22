package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_Success(t *testing.T) {
	yamlData := `
marketstack:
  key: "test-ms-key"
  stocks:
    - AAPL
    - MSFT

yahoo:
  stocks:
    - MTLN
    - GOOG

fixer:
  key: "test-fixer-key"
  pairs:
    - { from: "EUR", to: "USD" }
    - { from: "GBP", to: "USD" }

ledger:
  price_db: "/tmp/prices.db"
`
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(yamlData), 0o600); err != nil {
		t.Fatalf("could not create temp config: %v", err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Marketstack.Key != "test-ms-key" {
		t.Errorf("expected Marketstack.Key 'test-ms-key', got %q", cfg.Marketstack.Key)
	}
	if len(cfg.Marketstack.Stocks) != 2 || cfg.Marketstack.Stocks[0] != "AAPL" || cfg.Marketstack.Stocks[1] != "MSFT" {
		t.Errorf("unexpected Marketstack.Stocks: %v", cfg.Marketstack.Stocks)
	}

	if len(cfg.Yahoo.Stocks) != 2 {
		t.Errorf("expected 2 yahoo stocks, got %d", len(cfg.Yahoo.Stocks))
	}
	if cfg.Yahoo.Stocks[0] != "MTLN" || cfg.Yahoo.Stocks[1] != "GOOG" {
		t.Errorf("unexpected Yahoo.Stocks: %v", cfg.Yahoo.Stocks)
	}

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
	yamlData := `
marketstack:
  key: "test"
  stocks: [AAPL
ledger:
  price_db: "/tmp/prices.db"
`
	path := filepath.Join(t.TempDir(), "malformed.yaml")
	if err := os.WriteFile(path, []byte(yamlData), 0o600); err != nil {
		t.Fatalf("could not create temp config: %v", err)
	}

	_, err := LoadConfig(path)
	if err == nil {
		t.Error("expected YAML unmarshal error, got nil")
	}
}

func TestLoadConfig_WhenField_Below(t *testing.T) {
	yamlData := `
marketstack:
  key: "test-key"
  stocks: []

yahoo:
  stocks: []

fixer:
  key: "test-fixer-key"
  pairs: []

ledger:
  price_db: "/tmp/prices.db"

pushover:
  config:
    - token: "test-token"
      recipient: "test-recipient"
  notify:
    - stock: "AAPL"
      price: 150.0
      when: "below"
  notify_currency:
    - from: "EUR"
      to: "USD"
      price: 1.10
      when: "below"
`
	path := filepath.Join(t.TempDir(), "config_when_below.yaml")
	if err := os.WriteFile(path, []byte(yamlData), 0o600); err != nil {
		t.Fatalf("could not create temp config: %v", err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify stock notification with "when" field
	if len(cfg.Pushover.Notify) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(cfg.Pushover.Notify))
	}
	notif := cfg.Pushover.Notify[0]
	if notif.Stock != "AAPL" {
		t.Errorf("expected stock 'AAPL', got %q", notif.Stock)
	}
	if notif.Price != 150.0 {
		t.Errorf("expected price 150.0, got %f", notif.Price)
	}
	if notif.When != "below" {
		t.Errorf("expected when 'below', got %q", notif.When)
	}

	// Verify currency notification with "when" field
	if len(cfg.Pushover.NotifyCurrency) != 1 {
		t.Fatalf("expected 1 currency notification, got %d", len(cfg.Pushover.NotifyCurrency))
	}
	currNotif := cfg.Pushover.NotifyCurrency[0]
	if currNotif.From != "EUR" {
		t.Errorf("expected from 'EUR', got %q", currNotif.From)
	}
	if currNotif.To != "USD" {
		t.Errorf("expected to 'USD', got %q", currNotif.To)
	}
	if currNotif.Price != 1.10 {
		t.Errorf("expected price 1.10, got %f", currNotif.Price)
	}
	if currNotif.When != "below" {
		t.Errorf("expected when 'below', got %q", currNotif.When)
	}
}

func TestLoadConfig_WhenField_Omitted(t *testing.T) {
	yamlData := `
marketstack:
  key: "test-key"
  stocks: []

yahoo:
  stocks: []

fixer:
  key: "test-fixer-key"
  pairs: []

ledger:
  price_db: "/tmp/prices.db"

pushover:
  config:
    - token: "test-token"
      recipient: "test-recipient"
  notify:
    - stock: "MSFT"
      price: 300.0
  notify_currency:
    - from: "GBP"
      to: "USD"
      price: 1.25
`
	path := filepath.Join(t.TempDir(), "config_when_omitted.yaml")
	if err := os.WriteFile(path, []byte(yamlData), 0o600); err != nil {
		t.Fatalf("could not create temp config: %v", err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify stock notification without "when" field (should be empty string)
	if len(cfg.Pushover.Notify) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(cfg.Pushover.Notify))
	}
	notif := cfg.Pushover.Notify[0]
	if notif.Stock != "MSFT" {
		t.Errorf("expected stock 'MSFT', got %q", notif.Stock)
	}
	if notif.Price != 300.0 {
		t.Errorf("expected price 300.0, got %f", notif.Price)
	}
	if notif.When != "" {
		t.Errorf("expected when to be empty string, got %q", notif.When)
	}

	// Verify currency notification without "when" field (should be empty string)
	if len(cfg.Pushover.NotifyCurrency) != 1 {
		t.Fatalf("expected 1 currency notification, got %d", len(cfg.Pushover.NotifyCurrency))
	}
	currNotif := cfg.Pushover.NotifyCurrency[0]
	if currNotif.From != "GBP" {
		t.Errorf("expected from 'GBP', got %q", currNotif.From)
	}
	if currNotif.To != "USD" {
		t.Errorf("expected to 'USD', got %q", currNotif.To)
	}
	if currNotif.Price != 1.25 {
		t.Errorf("expected price 1.25, got %f", currNotif.Price)
	}
	if currNotif.When != "" {
		t.Errorf("expected when to be empty string, got %q", currNotif.When)
	}
}
