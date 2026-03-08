package archive

import (
	"fmt"
	"io"
)

// ProgressWriter reports download progress at 10% intervals.
type ProgressWriter struct {
	Out     io.Writer
	Label   string
	Total   int64
	written int64
	lastPct int
}

// NewProgressWriter constructs a ProgressWriter.
func NewProgressWriter(out io.Writer, label string, total int64) *ProgressWriter {
	return &ProgressWriter{
		Out:     out,
		Label:   label,
		Total:   total,
		lastPct: -1,
	}
}

func (w *ProgressWriter) Write(p []byte) (int, error) {
	n := len(p)
	w.written += int64(n)

	if w.Total > 0 {
		pct := min(int((w.written*100)/w.Total), 100)
		if pct != w.lastPct && pct%10 == 0 {
			w.lastPct = pct
			_, _ = fmt.Fprintf(w.Out, "Download progress: %d%%\n", pct)
		}
	}

	return n, nil
}
