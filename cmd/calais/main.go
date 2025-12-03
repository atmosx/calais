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
	"git.sr.ht/~atmosx/calais/pkg/providers/pushover"
	"git.sr.ht/~atmosx/calais/pkg/providers/yahoo"
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

	// Initialize Notifiers (Pushover)
	// We use the Notifier interface so this list can hold email providers in the future.
	var notifiers []pushover.Notifier

	// Check if Pushover config exists and initialize providers
	// Note: This assumes you have added `Pushover Pushover yaml:"pushover"` to your main Config struct.
	for _, pConf := range cfg.Pushover.Config {
		// Create a new Pushover client using the configuration
		n := pushover.New(pConf.Token, pConf.Recipient, http.DefaultClient)
		notifiers = append(notifiers, n)
	}

	// Helper function to check rules and send notifications
	checkAndNotify := func(symbol string, currentPrice float64) {
		for _, rule := range cfg.Pushover.Notify {
			// Notify if the symbol matches and price meets the target (>=)
			if rule.Stock == symbol && currentPrice >= rule.Price {
				msg := fmt.Sprintf("Price Alert: %s has reached %.2f (Target: %.2f)", symbol, currentPrice, rule.Price)

				for _, n := range notifiers {
					if err := n.Send("Stock Alert", msg); err != nil {
						logger.Error("failed to send notification", "type", "pushover", "error", err)
					} else {
						logger.Info("notification sent", "symbol", symbol, "price", currentPrice)
					}
				}
			}
		}
	}

	checkAndNotifyCurrency := func(from, to string, currentRate float64) {
		for _, rule := range cfg.Pushover.NotifyCurrency {
			if rule.From == from && rule.To == to && currentRate >= rule.Price {
				msg := fmt.Sprintf("Currency Alert: %s/%s has reached %.4f (Target: %.4f)", from, to, currentRate, rule.Price)
				for _, n := range notifiers {
					if err := n.Send("Currency Alert", msg); err != nil {
						logger.Error("failed to send currency notification", "type", "pushover", "error", err)
					} else {
						logger.Info("currency notification sent", "pair", from+"/"+to, "rate", currentRate)
					}
				}
			}
		}
	}

	msProvider := marketstack.New(cfg.Marketstack.Key, http.DefaultClient, logger)
	yahooProvider := yahoo.New(http.DefaultClient, logger)

	var currencyProvider providers.CurrencyProvider
	if cfg.Fixer.Key != "" {
		currencyProvider = fixer.New(cfg.Fixer.Key, http.DefaultClient, logger)
	}

	writer := ledger.NewWriter(cfg.Ledger.PriceDB)

	// Process Marketstack Stocks
	for _, symbol := range cfg.Marketstack.Stocks {
		sd, err := msProvider.FetchStock(symbol)
		if err != nil {
			logger.Error("failed to fetch marketstack stock", "symbol", symbol, "error", err)
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
		logger.Info("wrote stock price", "source", "marketstack", "symbol", sd.Symbol, "price", sd.Close)

		// Check for notifications
		checkAndNotify(sd.Symbol, sd.Close)
	}

	// Process Yahoo Stocks
	for _, symbol := range cfg.Yahoo.Stocks {
		sd, err := yahooProvider.FetchStock(symbol)
		if err != nil {
			logger.Error("failed to fetch yahoo stock", "symbol", symbol, "error", err)
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
		logger.Info("wrote stock price", "source", "yahoo", "symbol", sd.Symbol, "price", sd.Close)

		// Check for notifications
		checkAndNotify(sd.Symbol, sd.Close)
	}

	// Process Currencies
	if currencyProvider != nil {
		for _, p := range cfg.Fixer.Pairs {
			cd, err := currencyProvider.FetchCurrency(p.From, p.To)
			if err != nil {
				logger.Error("failed to fetch currency", "pair", p.From+"/"+p.To, "error", err)
				continue
			}
			if err := writer.Append(doctype.Record{
				Time:   cd.Date,
				Symbol: cd.From, // Ideally this might record the pair, e.g., "EUR/USD" depending on your doctype logic
				Price:  cd.Rate,
				Kind:   "currency",
			}); err != nil {
				logger.Error("failed to write currency price", "pair", p.From+"/"+p.To, "error", err)
				continue
			}
			logger.Info("wrote currency price", "pair", p.From+"/"+p.To, "rate", cd.Rate)

			// Check currency notifications
			checkAndNotifyCurrency(cd.From, cd.To, cd.Rate)
		}
	}
}
