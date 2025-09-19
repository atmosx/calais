package ledger

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"git.sr.ht/~atmosx/calais/pkg/doctype"
)

func TestWriter_Append(t *testing.T) {
	now := time.Date(2025, 8, 19, 14, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		record   doctype.Record
		expected string
		wantErr  bool
	}{
		{
			name: "currency",
			record: doctype.Record{
				Time:   now,
				Symbol: "EUR",
				Price:  1.123456,
				Kind:   "currency",
			},
			expected: "P 2025/08/19 14:30:00 EUR $1.123456\n",
			wantErr:  false,
		},
		{
			name: "commodity",
			record: doctype.Record{
				Time:   now,
				Symbol: "AAPL",
				Price:  150.75,
				Kind:   "commodity",
			},
			expected: "P 2025/08/19 14:30:00 AAPL €150.75\n",
			wantErr:  false,
		},
		{
			name: "unknown kind",
			record: doctype.Record{
				Time: now, Symbol: "X", Price: 1, Kind: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := filepath.Join(t.TempDir(), t.Name()+".ledger")
			if err := os.MkdirAll(filepath.Dir(tmp), 0o755); err != nil {
				t.Fatalf("MkdirAll: %v", err)
			}
			w := NewWriter(tmp)

			err := w.Append(tt.record)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Append() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			got, err := os.ReadFile(tmp)
			if err != nil {
				t.Fatalf("reading output file: %v", err)
			}
			if string(got) != tt.expected {
				t.Errorf("unexpected file content:\ngot:  %q\nwant: %q", string(got), tt.expected)
			}
		})
	}
}

func TestWriter_FileCreation(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "nonexistent")
	w := NewWriter(tmp)

	r := doctype.Record{
		Time:   time.Now(),
		Symbol: "TEST",
		Price:  99.99,
		Kind:   "commodity",
	}
	if err := w.Append(r); err != nil {
		t.Fatalf("expected file creation, got error: %v", err)
	}
	if _, err := os.Stat(tmp); os.IsNotExist(err) {
		t.Error("expected file to be created")
	}
}
