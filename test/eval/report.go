//go:build eval

package eval

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

type Report struct {
	Build string                 `json:"build"`
	Cells map[string]CellMetrics `json:"cells"`
}

func (r Report) WriteJSON(path string) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (r Report) Summary() string {
	var b strings.Builder
	fmt.Fprintf(&b, "eval report (build %s)\n", r.Build)
	names := make([]string, 0, len(r.Cells))
	for name := range r.Cells {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		c := r.Cells[name]
		fmt.Fprintf(&b, "  %-9s trials=%d collision_rate=%.3f corruption_rate=%.3f pass^k=%.3f\n",
			name, c.Trials, c.CollisionRate, c.CorruptionRate, c.PassCaretK)
	}
	return b.String()
}
