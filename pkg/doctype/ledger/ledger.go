package ledger

import (
	"fmt"
	"os"

	"git.sr.ht/~atmosx/calais/pkg/doctype"
)

type Writer struct{ filePath string }

func NewWriter(filePath string) *Writer { return &Writer{filePath: filePath} }

func (w *Writer) Append(r doctype.Record) error {
	f, err := os.OpenFile(w.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	var line string
	switch r.Kind {
	case "currency":
		line = fmt.Sprintf("P %s %s $%.6f\n",
			r.Time.Format("2006/01/02 15:04:05"), r.Symbol, r.Price)
	case "commodity":
		line = fmt.Sprintf("P %s %s €%.2f\n",
			r.Time.Format("2006/01/02 15:04:05"), r.Symbol, r.Price)
	default:
		return fmt.Errorf("unknown kind %q", r.Kind)
	}
	_, err = f.WriteString(line)
	return err
}
