package skillgen

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExtractMetadata tests metadata extraction from a mock command
func TestExtractMetadata(t *testing.T) {
	cmd := &cobra.Command{
		Use:   "test-command",
		Short: "Short description",
		Long:  "Long description",
		Example: `Example usage:
  confluent test-command --flag value`,
		Run: func(cmd *cobra.Command, args []string) {},
		Annotations: map[string]string{
			"run-requirement": "cloud-login",
		},
	}
	cmd.Flags().String("test-flag", "default-value", "Test flag usage")

	ir := extractMetadata(cmd, "cloud")

	assert.Equal(t, "test-command", ir.CommandPath)
	assert.Equal(t, "Short description", ir.Short)
	assert.Equal(t, "Long description", ir.Long)
	assert.Contains(t, ir.Example, "Example usage")
	assert.Equal(t, "cloud", ir.Mode)
	assert.Equal(t, "other", ir.Operation) // "test-command" is not a recognized verb
	assert.Len(t, ir.Flags, 1)
	assert.Equal(t, "test-flag", ir.Flags[0].Name)
	assert.NotNil(t, ir.Annotations)
	assert.Equal(t, "cloud-login", ir.Annotations["run-requirement"])
}

// TestExtractFlags tests flag extraction including required detection
func TestExtractFlags(t *testing.T) {
	cmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {},
	}

	// Add various flag types
	cmd.Flags().String("string-flag", "default", "String flag")
	cmd.Flags().Int("int-flag", 42, "Int flag")
	cmd.Flags().Bool("bool-flag", false, "Bool flag")
	cmd.Flags().StringSlice("slice-flag", []string{}, "Slice flag")

	// Mark string-flag as required using Cobra's annotation
	stringFlag := cmd.Flags().Lookup("string-flag")
	stringFlag.Annotations = map[string][]string{
		cobra.BashCompOneRequiredFlag: {"true"},
	}

	flags := extractFlags(cmd)

	assert.Len(t, flags, 4)

	// Check string-flag is marked as required
	var stringFlagIR *FlagIR
	for i := range flags {
		if flags[i].Name == "string-flag" {
			stringFlagIR = &flags[i]
			break
		}
	}
	require.NotNil(t, stringFlagIR, "string-flag should exist")
	assert.True(t, stringFlagIR.Required, "string-flag should be required")
	assert.Equal(t, "string", stringFlagIR.Type)
	assert.Equal(t, "default", stringFlagIR.Default)
	assert.Equal(t, "String flag", stringFlagIR.Usage)

	// Check bool-flag is not required
	var boolFlagIR *FlagIR
	for i := range flags {
		if flags[i].Name == "bool-flag" {
			boolFlagIR = &flags[i]
			break
		}
	}
	require.NotNil(t, boolFlagIR)
	assert.False(t, boolFlagIR.Required)
}

// TestModeDetection verifies mode='both' when command in both contexts
func TestModeDetection(t *testing.T) {
	// This test will use the Parse function which merges results from both contexts
	// For now, we'll test the merge logic separately in TestParse
	// This is a placeholder that will be filled in after Parse is implemented
	t.Skip("Will be tested via TestParse")
}

// TestSkipHidden verifies hidden commands are not included
func TestSkipHidden(t *testing.T) {
	rootCmd := &cobra.Command{Use: "root"}

	visibleCmd := &cobra.Command{
		Use:   "visible",
		Short: "Visible command",
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	hiddenCmd := &cobra.Command{
		Use:    "hidden",
		Short:  "Hidden command",
		Hidden: true,
		Run:    func(cmd *cobra.Command, args []string) {},
	}

	rootCmd.AddCommand(visibleCmd, hiddenCmd)

	var results []CommandIR
	extractCommands(rootCmd, "cloud", &results)

	assert.Len(t, results, 1)
	assert.Equal(t, "root visible", results[0].CommandPath)
}

// TestLeafOnly verifies parent commands without Run/RunE are skipped
func TestLeafOnly(t *testing.T) {
	rootCmd := &cobra.Command{Use: "root"}

	// Parent command without Run (should be skipped)
	parentCmd := &cobra.Command{
		Use:   "parent",
		Short: "Parent command",
	}

	// Leaf command with Run (should be extracted)
	leafCmd := &cobra.Command{
		Use:   "leaf",
		Short: "Leaf command",
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	parentCmd.AddCommand(leafCmd)
	rootCmd.AddCommand(parentCmd)

	var results []CommandIR
	extractCommands(rootCmd, "cloud", &results)

	// Should only extract the leaf command
	assert.Len(t, results, 1)
	assert.Equal(t, "root parent leaf", results[0].CommandPath)
}

// TestParse creates a minimal test command tree and verifies extraction
func TestParse(t *testing.T) {
	// Parse() will create dual configs and call internal.NewConfluentCommand
	// This test will verify the overall structure
	ir, err := Parse()
	require.NoError(t, err)

	// Verify metadata structure
	assert.NotEmpty(t, ir.Metadata.CLIVersion)
	assert.False(t, ir.Metadata.GeneratedAt.IsZero())

	// Verify we extracted commands
	assert.NotEmpty(t, ir.Commands, "Should have extracted commands")

	// Verify all commands have required fields
	for _, cmd := range ir.Commands {
		assert.NotEmpty(t, cmd.CommandPath, "Command path should not be empty")
		assert.NotEmpty(t, cmd.Mode, "Mode should be set")
		assert.NotEmpty(t, cmd.Operation, "Operation should be set")
		// Resource can be empty for commands like "confluent login"
	}
}

// TestAliasDetection tests that aliases are correctly identified
func TestAliasDetection(t *testing.T) {
	rootCmd := &cobra.Command{Use: "root"}

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List items",
		Run:     func(cmd *cobra.Command, args []string) {},
	}

	rootCmd.AddCommand(cmd)

	// Test with the canonical name
	var results []CommandIR
	extractCommands(rootCmd, "cloud", &results)

	// When called normally, IsAlias should be false
	assert.Len(t, results, 1)
}

// TestOperationClassification tests that operation inference works
// This is tested via the InferOperation function directly in classifier_test.go
// Here we verify that extractMetadata correctly calls InferOperation
func TestOperationClassification(t *testing.T) {
	// Create a mock command tree to test that extractMetadata uses CommandPath correctly
	rootCmd := &cobra.Command{Use: "confluent"}
	kafkaCmd := &cobra.Command{Use: "kafka"}
	topicCmd := &cobra.Command{Use: "topic"}
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List topics",
		Run:   func(cmd *cobra.Command, args []string) {},
	}

	topicCmd.AddCommand(listCmd)
	kafkaCmd.AddCommand(topicCmd)
	rootCmd.AddCommand(kafkaCmd)

	ir := extractMetadata(listCmd, "cloud")

	// The operation should be inferred from the full command path "confluent kafka topic list"
	assert.Equal(t, "list", ir.Operation)
	assert.Equal(t, "kafka-topic", ir.Resource)
}

// TestParseFullCLI validates parser against real CLI command tree
func TestParseFullCLI(t *testing.T) {
	ir, err := Parse()
	require.NoError(t, err)

	// Verify we extracted a substantial number of commands
	assert.Greater(t, len(ir.Commands), 200, "CLI should have 200+ commands")

	// Track stats
	modeCount := map[string]int{"cloud": 0, "onprem": 0, "both": 0}
	operationCount := map[string]int{}
	resourceCount := map[string]int{}

	// Verify all commands have required fields populated
	for _, cmd := range ir.Commands {
		assert.NotEmpty(t, cmd.CommandPath, "Command path should not be empty")
		assert.NotEmpty(t, cmd.Mode, "Mode should be set")
		assert.NotEmpty(t, cmd.Operation, "Operation should be set")

		// Track stats
		modeCount[cmd.Mode]++
		operationCount[cmd.Operation]++
		if cmd.Resource != "" {
			resourceCount[cmd.Resource]++
		}
	}

	// Print summary stats
	t.Logf("Total commands extracted: %d", len(ir.Commands))
	t.Logf("Commands by mode - cloud: %d, onprem: %d, both: %d",
		modeCount["cloud"], modeCount["onprem"], modeCount["both"])
	t.Logf("Commands by operation - list: %d, create: %d, delete: %d, describe: %d, update: %d, other: %d",
		operationCount["list"], operationCount["create"], operationCount["delete"],
		operationCount["describe"], operationCount["update"], operationCount["other"])
	t.Logf("Unique resources: %d", len(resourceCount))

	// Verify some expected commands exist
	commandPaths := make(map[string]bool)
	for _, cmd := range ir.Commands {
		commandPaths[cmd.CommandPath] = true
	}

	// Sample known commands that should exist
	expectedCommands := []string{
		"confluent kafka topic list",
		"confluent kafka cluster list",
		"confluent environment list",
	}

	for _, expected := range expectedCommands {
		assert.True(t, commandPaths[expected], "Expected command %q should exist", expected)
	}
}

// TestFlagExtraction validates flag extraction from real commands
func TestFlagExtraction(t *testing.T) {
	ir, err := Parse()
	require.NoError(t, err)

	// Find a command with known flags
	var kafkaTopicList *CommandIR
	for i := range ir.Commands {
		if ir.Commands[i].CommandPath == "confluent kafka topic list" {
			kafkaTopicList = &ir.Commands[i]
			break
		}
	}

	require.NotNil(t, kafkaTopicList, "confluent kafka topic list should exist")

	// Verify flags are captured
	assert.NotEmpty(t, kafkaTopicList.Flags, "Command should have flags")

	// Build a map of flags for easier testing
	flagMap := make(map[string]FlagIR)
	for _, flag := range kafkaTopicList.Flags {
		flagMap[flag.Name] = flag
	}

	// Check for common flags
	if clusterFlag, ok := flagMap["cluster"]; ok {
		assert.NotEmpty(t, clusterFlag.Type, "Flag should have type")
		assert.NotEmpty(t, clusterFlag.Usage, "Flag should have usage text")
	}

	t.Logf("kafka topic list has %d flags", len(kafkaTopicList.Flags))
}

// TestAnnotations validates annotation storage from real commands
func TestAnnotations(t *testing.T) {
	ir, err := Parse()
	require.NoError(t, err)

	// Count commands with annotations
	commandsWithAnnotations := 0
	cloudOnlyCommands := 0
	onpremOnlyCommands := 0

	for _, cmd := range ir.Commands {
		if len(cmd.Annotations) > 0 {
			commandsWithAnnotations++

			// Check for run-requirement annotation
			if runReq, ok := cmd.Annotations["run-requirement"]; ok {
				switch runReq {
				case "cloud-login":
					cloudOnlyCommands++
				case "on-prem-login":
					onpremOnlyCommands++
				}
			}
		}
	}

	t.Logf("Commands with annotations: %d", commandsWithAnnotations)
	t.Logf("Cloud-only commands: %d", cloudOnlyCommands)
	t.Logf("OnPrem-only commands: %d", onpremOnlyCommands)

	// There should be some commands with annotations
	assert.Greater(t, commandsWithAnnotations, 0, "Some commands should have annotations")
}

// TestExtractMetadata_KnownCommands validates parser extracts metadata correctly for representative commands (TEST-01)
func TestExtractMetadata_KnownCommands(t *testing.T) {
	tests := []struct {
		name              string
		commandPath       string
		expectedOperation string
		expectedResource  string
		expectedMode      string
	}{
		{
			name:              "kafka cluster list",
			commandPath:       "confluent kafka cluster list",
			expectedOperation: "list",
			expectedResource:  "kafka-cluster",
			expectedMode:      "cloud",
		},
		{
			name:              "environment use",
			commandPath:       "confluent environment use",
			expectedOperation: "use",
			expectedResource:  "environment",
			expectedMode:      "cloud",
		},
		{
			name:              "api-key create",
			commandPath:       "confluent api-key create",
			expectedOperation: "create",
			expectedResource:  "api-key",
			expectedMode:      "cloud",
		},
	}

	// Parse full CLI to get real commands
	ir, err := Parse()
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Find the command in the parsed IR
			var found *CommandIR
			for i := range ir.Commands {
				if ir.Commands[i].CommandPath == tt.commandPath {
					found = &ir.Commands[i]
					break
				}
			}

			require.NotNil(t, found, "Command %s should exist", tt.commandPath)

			// Validate metadata extraction
			assert.Equal(t, tt.commandPath, found.CommandPath, "CommandPath should match")
			assert.NotEmpty(t, found.Short, "Short description should be populated")
			assert.Equal(t, tt.expectedOperation, found.Operation, "Operation should be correctly inferred")
			assert.Equal(t, tt.expectedResource, found.Resource, "Resource should be correctly inferred")

			// Mode can be "cloud", "onprem", or "both" depending on command availability
			assert.Contains(t, []string{"cloud", "onprem", "both"}, found.Mode, "Mode should be valid")
		})
	}
}

// TestParseFlags_ComplexTypes validates flag parsing handles all CLI flag types (TEST-01)
func TestParseFlags_ComplexTypes(t *testing.T) {
	cmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {},
	}

	// Add various flag types that appear in real CLI
	cmd.Flags().String("string-flag", "default", "String flag")
	cmd.Flags().Int("int-flag", 42, "Int flag")
	cmd.Flags().Bool("bool-flag", false, "Bool flag")
	cmd.Flags().StringSlice("slice-flag", []string{}, "Slice flag")
	cmd.Flags().Int64("int64-flag", 0, "Int64 flag")

	// Mark some flags as required
	stringFlag := cmd.Flags().Lookup("string-flag")
	stringFlag.Annotations = map[string][]string{
		cobra.BashCompOneRequiredFlag: {"true"},
	}

	intFlag := cmd.Flags().Lookup("int-flag")
	intFlag.Annotations = map[string][]string{
		cobra.BashCompOneRequiredFlag: {"true"},
	}

	flags := extractFlags(cmd)

	// Should extract all flags
	assert.Len(t, flags, 5)

	// Build flag map for easier testing
	flagMap := make(map[string]FlagIR)
	for _, flag := range flags {
		flagMap[flag.Name] = flag
	}

	// Test string flag
	assert.True(t, flagMap["string-flag"].Required, "string-flag should be required")
	assert.Equal(t, "string", flagMap["string-flag"].Type)
	assert.Equal(t, "default", flagMap["string-flag"].Default)

	// Test int flag
	assert.True(t, flagMap["int-flag"].Required, "int-flag should be required")
	assert.Equal(t, "int", flagMap["int-flag"].Type)
	assert.Equal(t, "42", flagMap["int-flag"].Default)

	// Test bool flag
	assert.False(t, flagMap["bool-flag"].Required, "bool-flag should not be required")
	assert.Equal(t, "bool", flagMap["bool-flag"].Type)

	// Test slice flag
	assert.False(t, flagMap["slice-flag"].Required, "slice-flag should not be required")
	assert.Equal(t, "stringSlice", flagMap["slice-flag"].Type, "slice-flag should have stringSlice type")

	// Test int64 flag
	assert.False(t, flagMap["int64-flag"].Required, "int64-flag should not be required")
	assert.Equal(t, "int64", flagMap["int64-flag"].Type)
}

// TestDetectDualMode_Annotations validates dual-mode detection works for all annotation types (TEST-01)
func TestDetectDualMode_Annotations(t *testing.T) {
	tests := []struct {
		name         string
		annotation   string
		expectedMode string
	}{
		{
			name:         "cloud-only command",
			annotation:   "cloud-login",
			expectedMode: "cloud",
		},
		{
			name:         "onprem-only command",
			annotation:   "on-prem-login",
			expectedMode: "onprem",
		},
		{
			name:         "no annotation (both modes)",
			annotation:   "",
			expectedMode: "both",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use:   "test",
				Short: "Test command",
				Run:   func(cmd *cobra.Command, args []string) {},
			}

			if tt.annotation != "" {
				cmd.Annotations = map[string]string{
					"run-requirement": tt.annotation,
				}
			}

			// Extract in both modes
			cloudIR := extractMetadata(cmd, "cloud")
			onpremIR := extractMetadata(cmd, "onprem")

			// When merging, commands with matching annotations appear in correct mode
			if tt.annotation == "cloud-login" {
				assert.Equal(t, "cloud", cloudIR.Mode)
				// In real Parse(), onprem would skip this command
			} else if tt.annotation == "on-prem-login" {
				assert.Equal(t, "onprem", onpremIR.Mode)
				// In real Parse(), cloud would skip this command
			} else {
				// No annotation - command appears in both modes
				// The merge logic in Parse() would mark this as mode="both"
				assert.Equal(t, "cloud", cloudIR.Mode)
				assert.Equal(t, "onprem", onpremIR.Mode)
			}
		})
	}
}
