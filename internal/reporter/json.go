package reporter

import (
	"encoding/json"
	"fmt"
	"io"
)

// WriteJSON writes the run result as pretty-printed JSON to w.
func WriteJSON(w io.Writer, result RunResult) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(result); err != nil {
		return fmt.Errorf("reporter: %w", err)
	}
	return nil
}
