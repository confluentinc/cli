package flink

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

// capTestCmd builds a bare command carrying just the two flags applyHumanRowCap
// reads, so the defaulting logic can be tested without the full command tree.
func capTestCmd(t *testing.T, output string, maxRows string) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{Use: "query", RunE: func(*cobra.Command, []string) error { return nil }}
	cmd.Flags().String("output", "human", "")
	cmd.Flags().Int("max-rows", 0, "")

	args := []string{"--output", output}
	if maxRows != "" {
		args = append(args, "--max-rows", maxRows)
	}
	cmd.SetArgs(args)
	require.NoError(t, cmd.Execute())
	return cmd
}

func TestApplyHumanRowCap(t *testing.T) {
	tests := []struct {
		name       string
		output     string
		maxRows    string // "" means the flag is not set
		wantMax    int
		wantcapped bool
	}{
		{name: "human, no --max-rows, gets the default cap", output: "human", maxRows: "", wantMax: humanRowCap, wantcapped: true},
		{name: "human, explicit --max-rows wins", output: "human", maxRows: "50", wantMax: 50, wantcapped: false},
		{name: "human, explicit --max-rows 0 means unlimited", output: "human", maxRows: "0", wantMax: 0, wantcapped: false},
		{name: "json is never capped", output: "json", maxRows: "", wantMax: 0, wantcapped: false},
		{name: "yaml is never capped", output: "yaml", maxRows: "", wantMax: 0, wantcapped: false},
		{name: "json respects explicit --max-rows", output: "json", maxRows: "5", wantMax: 5, wantcapped: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cmd := capTestCmd(t, test.output, test.maxRows)
			// The passed-in maxRows mirrors what resolveQueryFlags would have read.
			in := 0
			if test.maxRows != "" {
				in, _ = cmd.Flags().GetInt("max-rows")
			}
			gotMax, gotCapped := applyHumanRowCap(cmd, in)
			require.Equal(t, test.wantMax, gotMax)
			require.Equal(t, test.wantcapped, gotCapped)
		})
	}
}
