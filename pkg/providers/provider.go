package providers

import "time"

// StockData represents a single stock end-of-day record.
type StockData struct {
	Symbol string
	Date   time.Time
	Close  float64
	Volume float64
}

// CurrencyData represents a single currency pair rate.
type CurrencyData struct {
	From string
	To   string
	Rate float64
	Date time.Time
}

// Provider interfaces
type StockProvider interface {
	FetchStock(symbol string) (*StockData, error)
}

type CurrencyProvider interface {
	FetchCurrency(from, to string) (*CurrencyData, error)
}
