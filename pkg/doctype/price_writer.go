package doctype

import "time"

type PriceWriter interface {
	Append(record Record) error
}

type Record struct {
	Time   time.Time
	Symbol string
	Price  float64
	Kind   string
}
