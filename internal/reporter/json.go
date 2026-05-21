package reporter

import (
	"encoding/json"
	"fmt"
	"io"
)

// WriteJSON emits the run as indented JSON. This is the default format
// for agent consumption — stable schema, deterministic field order.
func WriteJSON(w io.Writer, r Run) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(r); err != nil {
		return fmt.Errorf("encode json report: %w", err)
	}
	return nil
}
