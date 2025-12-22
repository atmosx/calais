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

func TestLoadConfig_WhenFieldInNotification(t *testing.T) {
	yamlData := `
marketstack:
  key: "test-key"
  stocks:
    - AAPL

pushover:
  config:
    - token: "test-token"
      recipient: "test-recipient"
  notify:
    - stock: "AAPL"
      price: 150.0
      when: "below"
`
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(yamlData), 0o600); err != nil {
		t.Fatalf("could not create temp config: %v", err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if len(cfg.Pushover.Notify) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(cfg.Pushover.Notify))
	}

	notification := cfg.Pushover.Notify[0]
	if notification.Stock != "AAPL" {
		t.Errorf("expected Stock 'AAPL', got %q", notification.Stock)
	}
	if notification.Price != 150.0 {
		t.Errorf("expected Price 150.0, got %f", notification.Price)
	}
	if notification.When != "below" {
		t.Errorf("expected When 'below', got %q", notification.When)
	}
}

func TestLoadConfig_WhenFieldOmittedInNotification(t *testing.T) {
	yamlData := `
marketstack:
  key: "test-key"
  stocks:
    - MSFT

pushover:
  config:
    - token: "test-token"
      recipient: "test-recipient"
  notify:
    - stock: "MSFT"
      price: 200.0
`
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(yamlData), 0o600); err != nil {
		t.Fatalf("could not create temp config: %v", err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if len(cfg.Pushover.Notify) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(cfg.Pushover.Notify))
	}

	notification := cfg.Pushover.Notify[0]
	if notification.Stock != "MSFT" {
		t.Errorf("expected Stock 'MSFT', got %q", notification.Stock)
	}
	if notification.Price != 200.0 {
		t.Errorf("expected Price 200.0, got %f", notification.Price)
	}
	// When field should be empty string when omitted
	if notification.When != "" {
		t.Errorf("expected When to be empty string, got %q", notification.When)
	}
}

func TestLoadConfig_WhenFieldInCurrencyNotification(t *testing.T) {
	yamlData := `
fixer:
  key: "test-fixer-key"
  pairs:
    - from: "EUR"
      to: "USD"

pushover:
  config:
    - token: "test-token"
      recipient: "test-recipient"
  notify_currency:
    - from: "EUR"
      to: "USD"
      price: 1.13
      when: "below"
`
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(yamlData), 0o600); err != nil {
		t.Fatalf("could not create temp config: %v", err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if len(cfg.Pushover.NotifyCurrency) != 1 {
		t.Fatalf("expected 1 currency notification, got %d", len(cfg.Pushover.NotifyCurrency))
	}

	notification := cfg.Pushover.NotifyCurrency[0]
	if notification.From != "EUR" {
		t.Errorf("expected From 'EUR', got %q", notification.From)
	}
	if notification.To != "USD" {
		t.Errorf("expected To 'USD', got %q", notification.To)
	}
	if notification.Price != 1.13 {
		t.Errorf("expected Price 1.13, got %f", notification.Price)
	}
	if notification.When != "below" {
		t.Errorf("expected When 'below', got %q", notification.When)
	}
}

func TestLoadConfig_WhenFieldOmittedInCurrencyNotification(t *testing.T) {
	yamlData := `
fixer:
  key: "test-fixer-key"
  pairs:
    - from: "GBP"
      to: "USD"

pushover:
  config:
    - token: "test-token"
      recipient: "test-recipient"
  notify_currency:
    - from: "GBP"
      to: "USD"
      price: 1.25
`
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(yamlData), 0o600); err != nil {
		t.Fatalf("could not create temp config: %v", err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if len(cfg.Pushover.NotifyCurrency) != 1 {
		t.Fatalf("expected 1 currency notification, got %d", len(cfg.Pushover.NotifyCurrency))
	}

	notification := cfg.Pushover.NotifyCurrency[0]
	if notification.From != "GBP" {
		t.Errorf("expected From 'GBP', got %q", notification.From)
	}
	if notification.To != "USD" {
		t.Errorf("expected To 'USD', got %q", notification.To)
	}
	if notification.Price != 1.25 {
		t.Errorf("expected Price 1.25, got %f", notification.Price)
	}
	// When field should be empty string when omitted
	if notification.When != "" {
		t.Errorf("expected When to be empty string, got %q", notification.When)
	}
}

func TestLoadConfig_WhenFieldInvalidValue(t *testing.T) {
	yamlData := `
marketstack:
  key: "test-key"
  stocks:
    - AAPL

pushover:
  config:
    - token: "test-token"
      recipient: "test-recipient"
  notify:
    - stock: "AAPL"
      price: 150.0
      when: "invalid"
  notify_currency:
    - from: "EUR"
      to: "USD"
      price: 1.13
      when: "above"
`
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(yamlData), 0o600); err != nil {
		t.Fatalf("could not create temp config: %v", err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// YAML parsing should succeed even with invalid values
	// The validation should happen at runtime
	if len(cfg.Pushover.Notify) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(cfg.Pushover.Notify))
	}
	if cfg.Pushover.Notify[0].When != "invalid" {
		t.Errorf("expected When 'invalid', got %q", cfg.Pushover.Notify[0].When)
	}

	if len(cfg.Pushover.NotifyCurrency) != 1 {
		t.Fatalf("expected 1 currency notification, got %d", len(cfg.Pushover.NotifyCurrency))
	}
	if cfg.Pushover.NotifyCurrency[0].When != "above" {
		t.Errorf("expected When 'above', got %q", cfg.Pushover.NotifyCurrency[0].When)
	}
}
