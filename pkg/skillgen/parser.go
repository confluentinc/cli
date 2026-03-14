package skillgen

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/confluentinc/cli/v4/internal"
	"github.com/confluentinc/cli/v4/pkg/config"
	pversion "github.com/confluentinc/cli/v4/pkg/version"
)

// Parse builds Cloud and OnPrem command trees using dual-config pattern,
// extracts metadata from all available commands, and returns a merged IR
// with mode detection.
func Parse() (*IR, error) {
	// Create two configs following the dual-config pattern from cmd/docs/main.go
	configs := []*config.Config{
		{
			CurrentContext: "Cloud",
			Contexts: map[string]*config.Context{
				"Cloud": {PlatformName: "https://confluent.cloud"},
			},
		},
		{
			CurrentContext: "On-Prem",
			Contexts: map[string]*config.Context{
				"On-Prem": {PlatformName: "https://example.com"},
			},
		},
	}

	// Set required fields for both configs
	for _, cfg := range configs {
		cfg.IsTest = true
		// Initialize version with default test values
		cfg.Version = pversion.NewVersion("0.0.0-test", "test", "test")
	}

	// Build command trees
	cloudCmd := internal.NewConfluentCommand(configs[0])
	onpremCmd := internal.NewConfluentCommand(configs[1])

	// Extract commands from each tree
	var cloudCommands []CommandIR
	var onpremCommands []CommandIR

	extractCommands(cloudCmd, "cloud", &cloudCommands)
	extractCommands(onpremCmd, "onprem", &onpremCommands)

	// Merge results, marking commands that appear in both as mode='both'
	merged := mergeCommands(cloudCommands, onpremCommands)

	// Build IR structure
	ir := &IR{
		Metadata: Metadata{
			CLIVersion:  configs[0].Version.Version,
			GeneratedAt: time.Now().UTC(),
		},
		Commands: merged,
	}

	return ir, nil
}

// extractCommands recursively traverses the command tree and extracts metadata
// from leaf commands (commands with Run/RunE). Hidden and unavailable commands
// are filtered out.
func extractCommands(cmd *cobra.Command, mode string, results *[]CommandIR) {
	// Skip unavailable or hidden commands
	if !cmd.IsAvailableCommand() || cmd.Hidden {
		return
	}

	// Only extract metadata from leaf commands (runnable commands)
	if cmd.Runnable() {
		ir := extractMetadata(cmd, mode)
		*results = append(*results, ir)
	}

	// Recurse into subcommands
	for _, sub := range cmd.Commands() {
		extractCommands(sub, mode, results)
	}
}

// extractMetadata extracts all metadata from a single command and builds a CommandIR.
func extractMetadata(cmd *cobra.Command, mode string) CommandIR {
	ir := CommandIR{
		CommandPath: cmd.CommandPath(),
		Short:       cmd.Short,
		Long:        cmd.Long,
		Example:     cmd.Example,
		Flags:       extractFlags(cmd),
		Annotations: cmd.Annotations,
		Operation:   InferOperation(cmd.CommandPath()),
		Resource:    InferResource(cmd.CommandPath()),
		Mode:        mode,
	}

	return ir
}

// extractFlags extracts flag metadata from a command, including required status
// detected via Cobra's BashCompOneRequiredFlag annotation.
func extractFlags(cmd *cobra.Command) []FlagIR {
	var flags []FlagIR

	cmd.Flags().VisitAll(func(pf *pflag.Flag) {
		// Guard against malformed flags with nil Value
		if pf.Value == nil {
			return
		}

		// Check if flag is required via annotation
		required := false
		if pf.Annotations != nil {
			if _, ok := pf.Annotations[cobra.BashCompOneRequiredFlag]; ok {
				required = true
			}
		}

		flag := FlagIR{
			Name:     pf.Name,
			Type:     pf.Value.Type(),
			Required: required,
			Default:  pf.DefValue,
			Usage:    pf.Usage,
		}

		flags = append(flags, flag)
	})

	return flags
}

// mergeCommands merges Cloud and OnPrem command lists, detecting commands that
// appear in both contexts and marking them as mode='both'.
func mergeCommands(cloudCommands, onpremCommands []CommandIR) []CommandIR {
	// Build index map of command paths from Cloud extraction.
	// We store indices (not pointers) because append below may reallocate
	// the slice, which would invalidate any pointers into it.
	cloudMap := make(map[string]int)
	for i := range cloudCommands {
		cloudMap[cloudCommands[i].CommandPath] = i
	}

	// Process OnPrem results
	for _, onpremCmd := range onpremCommands {
		if idx, exists := cloudMap[onpremCmd.CommandPath]; exists {
			// Command exists in both contexts - update mode
			cloudCommands[idx].Mode = "both"
		} else {
			// Command only exists in OnPrem - add it
			cloudCommands = append(cloudCommands, onpremCmd)
		}
	}

	return cloudCommands
}
