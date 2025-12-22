package config

import (
	"fmt"
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

func TestLoadConfig_ValidWhenField(t *testing.T) {
	tests := []struct {
		name      string
		whenValue string
		wantError bool
	}{
		{
			name:      "empty when field (default behavior)",
			whenValue: "",
			wantError: false,
		},
		{
			name:      "when field set to below",
			whenValue: "below",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			whenLine := ""
			if tt.whenValue != "" {
				whenLine = fmt.Sprintf("\n      when: %q", tt.whenValue)
			}

			yamlData := fmt.Sprintf(`
marketstack:
  key: "test-key"
  stocks:
    - AAPL

pushover:
  notify:
    - stock: "AAPL"
      price: 150.0%s
`, whenLine)

			path := filepath.Join(t.TempDir(), "config.yaml")
			if err := os.WriteFile(path, []byte(yamlData), 0o600); err != nil {
				t.Fatalf("could not create temp config: %v", err)
			}

			cfg, err := LoadConfig(path)
			if tt.wantError {
				if err == nil {
					t.Errorf("expected error for when=%q, got nil", tt.whenValue)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for when=%q: %v", tt.whenValue, err)
				}
				if cfg == nil {
					t.Error("expected config to be non-nil")
				}
			}
		})
	}
}

func TestLoadConfig_InvalidWhenField(t *testing.T) {
	tests := []struct {
		name      string
		whenValue string
	}{
		{
			name:      "invalid value: above",
			whenValue: "above",
		},
		{
			name:      "invalid value: greater",
			whenValue: "greater",
		},
		{
			name:      "typo: bellow",
			whenValue: "bellow",
		},
		{
			name:      "invalid value: less",
			whenValue: "less",
		},
		{
			name:      "invalid value: under",
			whenValue: "under",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yamlData := fmt.Sprintf(`
marketstack:
  key: "test-key"
  stocks:
    - AAPL

pushover:
  notify:
    - stock: "AAPL"
      price: 150.0
      when: %q
`, tt.whenValue)

			path := filepath.Join(t.TempDir(), "config.yaml")
			if err := os.WriteFile(path, []byte(yamlData), 0o600); err != nil {
				t.Fatalf("could not create temp config: %v", err)
			}

			_, err := LoadConfig(path)
			if err == nil {
				t.Errorf("expected error for when=%q, got nil", tt.whenValue)
			}
		})
	}
}

func TestLoadConfig_InvalidWhenFieldCurrency(t *testing.T) {
	yamlData := `
fixer:
  key: "test-key"
  pairs:
    - { from: "EUR", to: "USD" }

pushover:
  notify_currency:
    - from: "EUR"
      to: "USD"
      price: 1.2
      when: "above"
`

	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(yamlData), 0o600); err != nil {
		t.Fatalf("could not create temp config: %v", err)
	}

	_, err := LoadConfig(path)
	if err == nil {
		t.Error("expected error for invalid currency notification when field, got nil")
	}
}
