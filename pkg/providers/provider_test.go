package providers

import (
	"testing"
	"time"
)

// TestStockData ensures the StockData struct can be created and its fields assigned correctly.
func TestStockData(t *testing.T) {
	date := time.Now()
	sd := StockData{
		Symbol: "TEST",
		Date:   date,
		Close:  123.45,
		Volume: 1000000,
	}

	if sd.Symbol != "TEST" {
		t.Errorf("expected Symbol to be 'TEST', got '%s'", sd.Symbol)
	}
	if !sd.Date.Equal(date) {
		t.Errorf("expected Date to be '%v', got '%v'", date, sd.Date)
	}
	if sd.Close != 123.45 {
		t.Errorf("expected Close to be 123.45, got '%f'", sd.Close)
	}
	if sd.Volume != 1000000 {
		t.Errorf("expected Volume to be 1000000, got '%f'", sd.Volume)
	}
}
