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

// validateWhen checks if the When field contains a valid value.
// Valid values are "" (empty) or "below".
func validateWhen(when string) error {
	if when != "" && when != "below" {
		return fmt.Errorf("invalid 'when' value: %q (must be empty or \"below\")", when)
	}
	return nil
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
	
	// Validate 'when' field in stock notifications
	for i, notif := range cfg.Pushover.Notify {
		if err := validateWhen(notif.When); err != nil {
			return nil, fmt.Errorf("invalid notification config at index %d for stock %q: %w", i, notif.Stock, err)
		}
	}
	
	// Validate 'when' field in currency notifications
	for i, notif := range cfg.Pushover.NotifyCurrency {
		if err := validateWhen(notif.When); err != nil {
			return nil, fmt.Errorf("invalid currency notification config at index %d for %s/%s: %w", i, notif.From, notif.To, err)
		}
	}
	
	return &cfg, nil
}
