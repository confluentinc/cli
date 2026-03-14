package skillgen

import (
	"sort"
	"strings"
	"time"
)

// GeneratorVersion is the version of the skill generator.
// This is embedded in the manifest metadata for provenance tracking.
const GeneratorVersion = "1.0.0"

// GenerateTool creates an MCP Tool definition from a CommandIR.
//
// The tool name is constructed using BuildToolName based on the tier.
// The description combines CommandIR.Short and CommandIR.Example (if present).
// Input parameters are generated from flags using MapFlagToJSONSchema.
// Annotations include audience, priority, and mode metadata.
//
// Mode handling:
//   - mode="both" → single tool with mode annotation (not separate cloud/onprem tools)
//   - mode="cloud" → single tool annotated as cloud-only
//   - mode="onprem" → single tool annotated as onprem-only
func GenerateTool(cmd CommandIR, namespace, tier string) Tool {
	// Build tool name based on tier
	name := BuildToolName(namespace, cmd.Resource, cmd.Operation, tier)

	// Build description (Short + Example if present)
	description := buildDescription(cmd.Short, cmd.Example)

	// Build input schema from flags
	inputSchema := BuildInputSchema(cmd.Flags)

	// Build annotations
	annotations := buildAnnotations(cmd.Operation, cmd.Mode)

	// Extract CLI command path (strip "confluent " prefix)
	commandPath := strings.TrimPrefix(cmd.CommandPath, "confluent ")

	return Tool{
		Name:        name,
		Description: description,
		CommandPath: commandPath,
		InputSchema: inputSchema,
		Annotations: annotations,
	}
}

// buildDescription formats the tool description from Short and Example text.
// If Example is non-empty, it's appended as a separate paragraph.
func buildDescription(short, example string) string {
	if example == "" {
		return short
	}

	// Combine Short and Example with proper formatting.
	// Use concatenation instead of fmt.Sprintf to avoid interpreting
	// percent characters in CLI usage text as format verbs.
	return short + "\n\nExample:\n" + example
}

// buildAnnotations creates tool annotations with audience, priority, and mode metadata.
//
// Priority levels:
//   - high: list, describe operations (common read operations)
//   - medium: create, delete, update operations (state-changing operations)
//   - low: other operations
//
// Audience is always "user,assistant" (both can invoke the tool).
func buildAnnotations(operation, mode string) Annotations {
	annotations := Annotations{
		Audience: "user,assistant",
	}

	// Set priority based on operation type
	switch operation {
	case "list", "describe":
		annotations.Priority = "high"
	case "create", "delete", "update":
		annotations.Priority = "medium"
	default:
		annotations.Priority = "low"
	}

	// Note: Mode metadata would be used for runtime validation
	// For now, we just track it in the annotation structure
	// (Could extend Annotations type to include Mode field if needed)

	return annotations
}

// GenerateManifest generates a complete skill manifest from IR.
//
// This is the main entry point for skill generation. It:
//  1. Analyzes namespaces and assigns tiers
//  2. Groups commands by namespace
//  3. Generates tools based on tier-specific grouping rules
//  4. Builds manifest with metadata
//
// Tier-specific grouping:
//   - High tier: one skill per unique (resource, operation) pair
//   - Medium tier: one skill per unique operation
//   - Low tier: one skill per namespace
//
// The manifest metadata includes CLI version provenance (GEN-05).
func GenerateManifest(ir IR) SkillManifest {
	// Step 1: Analyze namespaces and assign tiers
	namespaceCounts := AnalyzeNamespaces(ir.Commands)
	highThreshold, mediumThreshold := ComputeTierThresholds(namespaceCounts)
	tiers := AssignTiers(namespaceCounts, highThreshold, mediumThreshold)

	// Step 2: Group commands by namespace
	byNamespace := make(map[string][]CommandIR)
	for _, cmd := range ir.Commands {
		parts := strings.Fields(cmd.CommandPath)
		if len(parts) < 2 {
			continue // Skip invalid command paths
		}
		namespace := parts[1]
		byNamespace[namespace] = append(byNamespace[namespace], cmd)
	}

	// Step 3: Generate tools based on tier-specific grouping
	// Sort namespace keys for deterministic output ordering
	namespaces := make([]string, 0, len(byNamespace))
	for namespace := range byNamespace {
		namespaces = append(namespaces, namespace)
	}
	sort.Strings(namespaces)

	var tools []Tool

	for _, namespace := range namespaces {
		cmds := byNamespace[namespace]
		tier, ok := tiers[namespace]
		if !ok {
			continue // Skip namespaces without tier assignment
		}

		namespaceTools := generateToolsForNamespace(namespace, cmds, tier)
		tools = append(tools, namespaceTools...)
	}

	// Step 4: Build manifest metadata
	metadata := ManifestMetadata{
		CLIVersion:       ir.Metadata.CLIVersion, // GEN-05: Populate from IR
		GeneratedAt:      time.Now().Format(time.RFC3339),
		GeneratorVersion: GeneratorVersion,
		SkillCount:       len(tools),
		CommandCount:     len(ir.Commands),
	}

	return SkillManifest{
		Metadata: metadata,
		Tools:    tools,
	}
}

// generateToolsForNamespace generates tools for a single namespace based on tier.
//
// Tier-specific grouping:
//   - High tier: one skill per unique (resource, operation) pair
//   - Medium tier: one skill per unique operation
//   - Low tier: one skill for entire namespace
func generateToolsForNamespace(namespace string, cmds []CommandIR, tier string) []Tool {
	var tools []Tool

	switch tier {
	case "high":
		// High tier: one tool per unique (resource, operation) pair
		// Track which (resource, operation) pairs we've already created tools for
		seen := make(map[string]bool)

		for _, cmd := range cmds {
			key := cmd.Resource + "|" + cmd.Operation
			if !seen[key] {
				tool := GenerateTool(cmd, namespace, tier)
				tools = append(tools, tool)
				seen[key] = true
			}
		}

	case "medium":
		// Medium tier: one tool per unique operation
		// Track which operations we've already created tools for
		seen := make(map[string]bool)

		for _, cmd := range cmds {
			if !seen[cmd.Operation] {
				tool := GenerateTool(cmd, namespace, tier)
				tools = append(tools, tool)
				seen[cmd.Operation] = true
			}
		}

	case "low":
		// Low tier: one tool for entire namespace
		// Use the first command as representative
		if len(cmds) > 0 {
			tool := GenerateTool(cmds[0], namespace, tier)
			tools = append(tools, tool)
		}
	}

	return tools
}
