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

func TestLoadConfig_ValidWhenField(t *testing.T) {
	tests := []struct {
		name      string
		yamlData  string
		wantError bool
	}{
		{
			name: "when field is empty",
			yamlData: `
marketstack:
  key: "test-key"
  stocks:
    - AAPL
ledger:
  price_db: "/tmp/prices.db"
pushover:
  config:
    - token: "test-token"
      recipient: "test-recipient"
  notify:
    - stock: "AAPL"
      price: 150.0
      when: ""
`,
			wantError: false,
		},
		{
			name: "when field is below",
			yamlData: `
marketstack:
  key: "test-key"
  stocks:
    - AAPL
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
`,
			wantError: false,
		},
		{
			name: "when field is omitted",
			yamlData: `
marketstack:
  key: "test-key"
  stocks:
    - AAPL
ledger:
  price_db: "/tmp/prices.db"
pushover:
  config:
    - token: "test-token"
      recipient: "test-recipient"
  notify:
    - stock: "AAPL"
      price: 150.0
`,
			wantError: false,
		},
		{
			name: "currency notification with below",
			yamlData: `
fixer:
  key: "test-key"
  pairs:
    - { from: "EUR", to: "USD" }
ledger:
  price_db: "/tmp/prices.db"
pushover:
  config:
    - token: "test-token"
      recipient: "test-recipient"
  notify_currency:
    - from: "EUR"
      to: "USD"
      price: 1.10
      when: "below"
`,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.yaml")
			if err := os.WriteFile(path, []byte(tt.yamlData), 0o600); err != nil {
				t.Fatalf("could not create temp config: %v", err)
			}

			_, err := LoadConfig(path)
			if (err != nil) != tt.wantError {
				t.Errorf("LoadConfig() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestLoadConfig_InvalidWhenField(t *testing.T) {
	tests := []struct {
		name         string
		yamlData     string
		wantErrorMsg string
	}{
		{
			name: "invalid when field - above",
			yamlData: `
marketstack:
  key: "test-key"
  stocks:
    - AAPL
ledger:
  price_db: "/tmp/prices.db"
pushover:
  config:
    - token: "test-token"
      recipient: "test-recipient"
  notify:
    - stock: "AAPL"
      price: 150.0
      when: "above"
`,
			wantErrorMsg: "invalid 'when' value in notify[0] for stock \"AAPL\"",
		},
		{
			name: "invalid when field - greater",
			yamlData: `
marketstack:
  key: "test-key"
  stocks:
    - AAPL
ledger:
  price_db: "/tmp/prices.db"
pushover:
  config:
    - token: "test-token"
      recipient: "test-recipient"
  notify:
    - stock: "AAPL"
      price: 150.0
      when: "greater"
`,
			wantErrorMsg: "invalid 'when' value in notify[0] for stock \"AAPL\"",
		},
		{
			name: "invalid when field - typo bellow",
			yamlData: `
marketstack:
  key: "test-key"
  stocks:
    - AAPL
ledger:
  price_db: "/tmp/prices.db"
pushover:
  config:
    - token: "test-token"
      recipient: "test-recipient"
  notify:
    - stock: "AAPL"
      price: 150.0
      when: "bellow"
`,
			wantErrorMsg: "invalid 'when' value in notify[0] for stock \"AAPL\"",
		},
		{
			name: "invalid when field in currency notification",
			yamlData: `
fixer:
  key: "test-key"
  pairs:
    - { from: "EUR", to: "USD" }
ledger:
  price_db: "/tmp/prices.db"
pushover:
  config:
    - token: "test-token"
      recipient: "test-recipient"
  notify_currency:
    - from: "EUR"
      to: "USD"
      price: 1.10
      when: "above"
`,
			wantErrorMsg: "invalid 'when' value in notify_currency[0] for EUR/USD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.yaml")
			if err := os.WriteFile(path, []byte(tt.yamlData), 0o600); err != nil {
				t.Fatalf("could not create temp config: %v", err)
			}

			_, err := LoadConfig(path)
			if err == nil {
				t.Error("expected error for invalid 'when' field, got nil")
			} else if tt.wantErrorMsg != "" && !contains(err.Error(), tt.wantErrorMsg) {
				t.Errorf("expected error message to contain %q, got %q", tt.wantErrorMsg, err.Error())
			}
		})
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
