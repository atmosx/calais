package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type MarketstackConfig struct {
	Key    string   `yaml:"key"`
	Stocks []string `yaml:"stocks"`
}

type YahooConfig struct {
	Stocks []string `yaml:"stocks"`
}

type Pair struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

type FixerConfig struct {
	Key   string `yaml:"key"`
	Pairs []Pair `yaml:"pairs"`
}

type LedgerConfig struct {
	PriceDB string `yaml:"price_db"`
}

type PushoverConfig struct {
	Token     string `yaml:"token"`
	Recipient string `yaml:"recipient"`
}

type Notification struct {
	Stock string  `yaml:"stock"`
	Price float64 `yaml:"price"`
	When  string  `yaml:"when"`
}

type CurrencyNotification struct {
	From  string  `yaml:"from"`
	To    string  `yaml:"to"`
	Price float64 `yaml:"price"`
	When  string  `yaml:"when"`
}

type Pushover struct {
	Config         []PushoverConfig       `yaml:"config"`
	Notify         []Notification         `yaml:"notify"`
	NotifyCurrency []CurrencyNotification `yaml:"notify_currency"`
}

type Config struct {
	Marketstack MarketstackConfig `yaml:"marketstack"`
	Yahoo       YahooConfig       `yaml:"yahoo"`
	Fixer       FixerConfig       `yaml:"fixer"`
	Ledger      LedgerConfig      `yaml:"ledger"`
	Pushover    Pushover          `yaml:"pushover"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Validate 'when' field in notifications
	if err := validateConfig(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// validateConfig validates the configuration, particularly the 'when' field values
func validateConfig(cfg *Config) error {
	// Validate stock notifications
	for i, notif := range cfg.Pushover.Notify {
		if err := validateWhenField(notif.When); err != nil {
			return fmt.Errorf("invalid 'when' value in notify[%d] for stock %q: %w", i, notif.Stock, err)
		}
	}

	// Validate currency notifications
	for i, notif := range cfg.Pushover.NotifyCurrency {
		if err := validateWhenField(notif.When); err != nil {
			return fmt.Errorf("invalid 'when' value in notify_currency[%d] for %s/%s: %w", i, notif.From, notif.To, err)
		}
	}

	return nil
}

// validateWhenField validates that the 'when' field contains only supported values
func validateWhenField(when string) error {
	// Empty string and "below" are valid values
	if when != "" && when != "below" {
		return fmt.Errorf("must be either empty or \"below\", got %q", when)
	}
	return nil
}
