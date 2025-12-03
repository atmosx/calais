package providers

import "time"

type StockData struct {
	Symbol string
	Date   time.Time
	Close  float64
	Volume float64
}

type CurrencyData struct {
	From string
	To   string
	Rate float64
	Date time.Time
}

type StockProvider interface {
	FetchStock(symbol string) (*StockData, error)
}

type CurrencyProvider interface {
	FetchCurrency(from, to string) (*CurrencyData, error)
}
