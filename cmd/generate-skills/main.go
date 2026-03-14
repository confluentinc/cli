package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/confluentinc/cli/v4/pkg/skillgen"
)

func main() {
	// Parse command-line flags
	outputPath := flag.String("output", "cmd/generate-skills/ir.json", "output path for IR JSON file")
	skillsOutput := flag.String("skills-output", "pkg/mcp/skills.json", "output path for skills JSON file")
	validateOnly := flag.Bool("validate", false, "Validate existing manifest without regenerating")
	flag.Parse()

	// If validate-only mode, validate and exit
	if *validateOnly {
		if err := validateManifest(*skillsOutput); err != nil {
			fmt.Fprintf(os.Stderr, "Validation failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✓ Skills manifest validation passed")
		return
	}

	// Generate IR and write to file
	if err := generateIR(*outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Generate skill manifest from IR
	if err := generateSkills(*outputPath, *skillsOutput); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// generateIR extracts command metadata, builds IR, marshals to JSON, and writes to file.
// This is extracted as a separate function to enable testing.
func generateIR(outputPath string) error {
	// Call parser to extract commands
	ir, err := skillgen.Parse()
	if err != nil {
		return fmt.Errorf("extracting commands: %w", err)
	}

	// Update generation timestamp (CLI version already set by parser)
	ir.Metadata.GeneratedAt = time.Now().UTC()

	// Marshal to pretty-printed JSON
	data, err := json.MarshalIndent(ir, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling IR: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("writing IR to %s: %w", outputPath, err)
	}

	// Print summary diagnostics
	printSummary(ir, outputPath)

	return nil
}

// printSummary prints diagnostic information about the extraction results.
func printSummary(ir *skillgen.IR, outputPath string) {
	cloudCount := 0
	onpremCount := 0

	for _, cmd := range ir.Commands {
		if cmd.Mode == "cloud" || cmd.Mode == "both" {
			cloudCount++
		}
		if cmd.Mode == "onprem" || cmd.Mode == "both" {
			onpremCount++
		}
	}

	fmt.Printf("✓ Extracted %d commands (%d cloud, %d onprem) → %s\n",
		len(ir.Commands), cloudCount, onpremCount, outputPath)
}

// generateSkills reads IR from irPath, generates skill manifest, and writes to skillsPath.
// This is extracted as a separate function to enable testing.
func generateSkills(irPath, skillsPath string) error {
	// Read IR from file
	data, err := os.ReadFile(irPath)
	if err != nil {
		return fmt.Errorf("reading IR from %s: %w", irPath, err)
	}

	// Parse IR JSON
	var ir skillgen.IR
	if err := json.Unmarshal(data, &ir); err != nil {
		return fmt.Errorf("parsing IR JSON: %w", err)
	}

	// Generate skill manifest
	manifest := skillgen.GenerateManifest(ir)

	// Validate manifest before writing
	if err := validateManifestData(&manifest); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Marshal to pretty-printed JSON (2-space indent, consistent with ir.json)
	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling manifest: %w", err)
	}

	// Write to file
	if err := os.WriteFile(skillsPath, manifestData, 0644); err != nil {
		return fmt.Errorf("writing manifest to %s: %w", skillsPath, err)
	}

	// Print summary
	printSkillsSummary(manifest, skillsPath)

	return nil
}

// printSkillsSummary prints diagnostic information about the generated skill manifest.
func printSkillsSummary(manifest skillgen.SkillManifest, outputPath string) {
	// Count skills by namespace (if manifest is organized that way)
	// For now, just print total skill count
	fmt.Printf("✓ Generated %d skills from %d commands → %s\n",
		manifest.Metadata.SkillCount, manifest.Metadata.CommandCount, outputPath)

	// Warn if skill count is high (approaching limit of 500)
	if manifest.Metadata.SkillCount > 450 {
		fmt.Printf("⚠ Warning: Skill count (%d) is approaching limit of 500\n", manifest.Metadata.SkillCount)
	}
}

// validateManifest reads a skills manifest from a file and validates it.
func validateManifest(path string) error {
	// Read manifest file
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading manifest from %s: %w", path, err)
	}

	// Parse manifest JSON
	var manifest skillgen.SkillManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("parsing manifest JSON: %w", err)
	}

	// Validate manifest data
	return validateManifestData(&manifest)
}

// validateManifestData validates the contents of a skills manifest.
func validateManifestData(manifest *skillgen.SkillManifest) error {
	// Check required fields: Version, CLIVersion, Tools not empty
	if manifest.Metadata.CLIVersion == "" {
		return fmt.Errorf("manifest missing required field: cli_version")
	}

	if len(manifest.Tools) == 0 {
		return fmt.Errorf("manifest must contain at least one tool")
	}

	// Verify tool_count == len(manifest.Tools)
	if manifest.Metadata.SkillCount != len(manifest.Tools) {
		return fmt.Errorf("skill_count (%d) != len(tools) (%d)", manifest.Metadata.SkillCount, len(manifest.Tools))
	}

	// Check tool count < 500 (MCP tools limit with safety margin)
	// Note: Current CLI has ~420 tools, so limit set to 500 to accommodate growth
	// while still preventing unbounded expansion
	if len(manifest.Tools) >= 500 {
		return fmt.Errorf("tool count (%d) must be less than 500", len(manifest.Tools))
	}

	// Build map to detect duplicate tool names
	seen := make(map[string]bool)
	for i, tool := range manifest.Tools {
		// Verify each tool has required fields
		if tool.Name == "" {
			return fmt.Errorf("tool at index %d missing required field: name", i)
		}
		if tool.Description == "" {
			return fmt.Errorf("tool %q missing required field: description", tool.Name)
		}
		if tool.InputSchema.Type == "" {
			return fmt.Errorf("tool %q missing required field: inputSchema.type", tool.Name)
		}

		// Check for duplicate names
		if seen[tool.Name] {
			return fmt.Errorf("duplicate tool name: %q", tool.Name)
		}
		seen[tool.Name] = true
	}

	return nil
}
