package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type MarketstackConfig struct {
	Key    string   `yaml:"key"`
	Stocks []string `yaml:"stocks"`
}

// New: Yahoo configuration struct
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

type Config struct {
	Marketstack MarketstackConfig `yaml:"marketstack"`
	Yahoo       YahooConfig       `yaml:"yahoo"`
	Fixer       FixerConfig       `yaml:"fixer"`
	Ledger      LedgerConfig      `yaml:"ledger"`
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
	return &cfg, nil
}
