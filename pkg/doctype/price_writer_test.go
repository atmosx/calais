package doctype

import (
	"testing"
	"time"
)

type mockWriter struct {
	calls []Record
}

func (m *mockWriter) Append(r Record) error {
	m.calls = append(m.calls, r)
	return nil
}

func TestPriceWriterInterface(t *testing.T) {
	now := time.Date(2025, 8, 19, 12, 0, 0, 0, time.UTC)
	w := &mockWriter{}

	records := []Record{
		{Time: now, Symbol: "EUR", Price: 1.123456, Kind: "currency"},
		{Time: now, Symbol: "AAPL", Price: 150.75, Kind: "commodity"},
	}

	for _, r := range records {
		if err := w.Append(r); err != nil {
			t.Fatalf("unexpected Append error: %v", err)
		}
	}

	if len(w.calls) != 2 {
		t.Fatalf("expected 2 Append calls, got %d", len(w.calls))
	}

	if w.calls[0] != records[0] || w.calls[1] != records[1] {
		t.Errorf("unexpected record flow: %+v", w.calls)
	}
}
