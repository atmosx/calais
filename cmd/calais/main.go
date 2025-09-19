package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"git.sr.ht/~atmosx/calais/internal/config"
	"git.sr.ht/~atmosx/calais/pkg/doctype"
	"git.sr.ht/~atmosx/calais/pkg/doctype/ledger"
	"git.sr.ht/~atmosx/calais/pkg/log"
	"git.sr.ht/~atmosx/calais/pkg/providers"
	"git.sr.ht/~atmosx/calais/pkg/providers/fixer"
	"git.sr.ht/~atmosx/calais/pkg/providers/marketstack"
)

var (
	version = "v0.0.1"
	commit  = "abcd1235"
	date    = "someDay"
)

func main() {
	configPath := flag.String("c", "/etc/calais/config.yaml", "path to configuration file")
	logLevel := flag.String("l", "Info", "log level (Info, debug)")
	showVersion := flag.Bool("version", false, "show version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("calais version %s, commit %s, built at %s\n", version, commit, date)
		os.Exit(0)
	}

	logger := log.New(os.Stdout, *logLevel)

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	stockProvider := marketstack.New(cfg.Marketstack.Key, http.DefaultClient, logger)

	var currencyProvider providers.CurrencyProvider
	if cfg.Fixer.Key != "" {
		currencyProvider = fixer.New(cfg.Fixer.Key, http.DefaultClient, logger)
	}

	writer := ledger.NewWriter(cfg.Ledger.PriceDB)

	for _, symbol := range cfg.Marketstack.Stocks {
		sd, err := stockProvider.FetchStock(symbol)
		if err != nil {
			logger.Error("failed to fetch stock", "symbol", symbol, "error", err)
			continue
		}
		if err := writer.Append(doctype.Record{
			Time:   sd.Date,
			Symbol: sd.Symbol,
			Price:  sd.Close,
			Kind:   "commodity",
		}); err != nil {
			logger.Error("failed to write stock price", "symbol", symbol, "error", err)
			continue
		}
		logger.Info("wrote stock price", "symbol", sd.Symbol, "price", sd.Close, "date", sd.Date)
	}

	if currencyProvider != nil {
		for _, p := range cfg.Fixer.Pairs {
			cd, err := currencyProvider.FetchCurrency(p.From, p.To)
			if err != nil {
				logger.Error("failed to fetch currency", "pair", p.From+"/"+p.To, "error", err)
				continue
			}
			if err := writer.Append(doctype.Record{
				Time:   cd.Date,
				Symbol: cd.From,
				Price:  cd.Rate,
				Kind:   "currency",
			}); err != nil {
				logger.Error("failed to write currency price", "pair", p.From+"/"+p.To, "error", err)
				continue
			}
			logger.Info("wrote currency price", "pair", p.From+"/"+p.To, "rate", cd.Rate, "date", cd.Date)
		}
	}
}
