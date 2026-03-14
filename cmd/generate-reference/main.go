package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Manifest represents the skills.json file structure
type Manifest struct {
	Metadata struct {
		CLIVersion       string `json:"cli_version"`
		GeneratedAt      string `json:"generated_at"`
		GeneratorVersion string `json:"generator_version"`
		SkillCount       int    `json:"skill_count"`
		CommandCount     int    `json:"command_count"`
	} `json:"metadata"`
	Tools []Tool `json:"tools"`
}

// Tool represents a single skill in the manifest
type Tool struct {
	Name        string      `json:"name"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
	Annotations Annotations `json:"annotations"`
}

// InputSchema represents the JSON schema for skill parameters
type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
}

// Property represents a single parameter definition
type Property struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Default     interface{} `json:"default,omitempty"`
}

// Annotations contains metadata about the skill
type Annotations struct {
	Audience string `json:"audience"`
	Priority string `json:"priority"`
}

// NamespaceGroup organizes skills by namespace
type NamespaceGroup struct {
	Name   string
	Title  string
	Skills []Tool
}

func main() {
	// Parse command-line flags
	inputPath := flag.String("input", "pkg/mcp/skills.json", "Input skills.json path")
	outputPath := flag.String("output", "docs/skills/REFERENCE.md", "Output REFERENCE.md path")
	flag.Parse()

	// Load manifest
	manifest, err := loadManifest(*inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading manifest: %v\n", err)
		os.Exit(1)
	}

	// Group by namespace
	groups := groupByNamespace(manifest)

	// Generate markdown
	markdown := generateMarkdown(manifest, groups)

	// Write to output
	if err := os.WriteFile(*outputPath, []byte(markdown), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Generated skill reference documentation → %s\n", *outputPath)
	fmt.Printf("  %d skills organized across %d namespaces\n", manifest.Metadata.SkillCount, len(groups))
}

// loadManifest reads and parses the skills.json file
func loadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	return &manifest, nil
}

// extractNamespace extracts the namespace from a skill name
// Example: "confluent_kafka_topic_list" → "kafka"
func extractNamespace(skillName string) string {
	// Remove "confluent_" prefix
	withoutPrefix := strings.TrimPrefix(skillName, "confluent_")

	// Split by underscore and take first part
	parts := strings.SplitN(withoutPrefix, "_", 2)
	if len(parts) > 0 {
		return parts[0]
	}

	return "other"
}

// namespaceTitle converts namespace slug to display title
// Example: "kafka" → "Kafka", "api-key" → "API Key", "schema-registry" → "Schema Registry"
func namespaceTitle(namespace string) string {
	// Handle special cases
	switch namespace {
	case "api-key":
		return "API Key"
	case "service-account":
		return "Service Account"
	case "schema-registry":
		return "Schema Registry"
	case "iam":
		return "IAM"
	default:
		// Title case for single words - capitalize first letter
		if len(namespace) == 0 {
			return namespace
		}
		return strings.ToUpper(namespace[:1]) + namespace[1:]
	}
}

// groupByNamespace organizes skills into namespace groups
func groupByNamespace(manifest *Manifest) []NamespaceGroup {
	// Map to collect skills by namespace
	namespaceMap := make(map[string][]Tool)

	for _, tool := range manifest.Tools {
		namespace := extractNamespace(tool.Name)
		namespaceMap[namespace] = append(namespaceMap[namespace], tool)
	}

	// Convert map to sorted slice
	var groups []NamespaceGroup
	for namespace, skills := range namespaceMap {
		// Sort skills alphabetically within namespace
		sort.Slice(skills, func(i, j int) bool {
			return skills[i].Name < skills[j].Name
		})

		groups = append(groups, NamespaceGroup{
			Name:   namespace,
			Title:  namespaceTitle(namespace),
			Skills: skills,
		})
	}

	// Sort groups alphabetically by namespace
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Name < groups[j].Name
	})

	return groups
}

// generateMarkdown produces the complete REFERENCE.md content
func generateMarkdown(manifest *Manifest, groups []NamespaceGroup) string {
	var b strings.Builder

	// Header
	b.WriteString("# Skill Reference\n\n")
	b.WriteString("**⚠️ This file is auto-generated from skills.json. Do not edit manually.**\n\n")
	b.WriteString(fmt.Sprintf("**CLI Version:** %s\n", manifest.Metadata.CLIVersion))
	b.WriteString(fmt.Sprintf("**Generated:** %s\n", manifest.Metadata.GeneratedAt))
	b.WriteString(fmt.Sprintf("**Total Skills:** %d\n\n", manifest.Metadata.SkillCount))

	b.WriteString("This reference documents all available Claude Code skills for the Confluent CLI. ")
	b.WriteString("Skills are organized by namespace, corresponding to the CLI command structure.\n\n")

	// Table of contents
	b.WriteString("## Table of Contents\n\n")
	for _, group := range groups {
		anchor := strings.ToLower(strings.ReplaceAll(group.Title, " ", "-"))
		b.WriteString(fmt.Sprintf("- [%s Skills](#%s-skills) (%d skills)\n", group.Title, anchor, len(group.Skills)))
	}
	b.WriteString("\n")

	// Namespace sections
	for _, group := range groups {
		b.WriteString(fmt.Sprintf("## %s Skills\n\n", group.Title))

		for _, skill := range group.Skills {
			formatSkill(&b, skill)
		}
	}

	return b.String()
}

// formatSkill writes a single skill entry to the markdown builder
func formatSkill(b *strings.Builder, skill Tool) {
	// Skill name as heading
	b.WriteString(fmt.Sprintf("### `%s`\n\n", skill.Name))

	// Description
	if skill.Description != "" {
		// Clean up description - trim whitespace and ensure single spacing
		description := strings.TrimSpace(skill.Description)
		// Replace multiple newlines with double newline for proper markdown paragraphs
		description = strings.ReplaceAll(description, "\n\n\n", "\n\n")
		b.WriteString(fmt.Sprintf("**Description:** %s\n\n", description))
	}

	// Required parameters
	if len(skill.InputSchema.Required) > 0 {
		b.WriteString("**Required Parameters:**\n\n")
		for _, paramName := range skill.InputSchema.Required {
			if prop, exists := skill.InputSchema.Properties[paramName]; exists {
				b.WriteString(fmt.Sprintf("- `%s`: %s\n", paramName, prop.Description))
			}
		}
		b.WriteString("\n")
	}

	// Optional parameters
	optionalParams := getOptionalParams(skill.InputSchema)
	if len(optionalParams) > 0 {
		b.WriteString("**Optional Parameters:**\n\n")
		for _, paramName := range optionalParams {
			if prop, exists := skill.InputSchema.Properties[paramName]; exists {
				defaultValue := formatDefault(prop.Default)
				if defaultValue != "" {
					b.WriteString(fmt.Sprintf("- `%s`: %s (default: `%s`)\n", paramName, prop.Description, defaultValue))
				} else {
					b.WriteString(fmt.Sprintf("- `%s`: %s\n", paramName, prop.Description))
				}
			}
		}
		b.WriteString("\n")
	}

	// Priority
	if skill.Annotations.Priority != "" {
		b.WriteString(fmt.Sprintf("**Priority:** %s\n\n", skill.Annotations.Priority))
	}

	b.WriteString("---\n\n")
}

// getOptionalParams returns parameter names that are not required, in sorted order
func getOptionalParams(schema InputSchema) []string {
	requiredSet := make(map[string]bool)
	for _, r := range schema.Required {
		requiredSet[r] = true
	}

	var optional []string
	for name := range schema.Properties {
		if !requiredSet[name] {
			optional = append(optional, name)
		}
	}

	sort.Strings(optional)
	return optional
}

// formatDefault converts a default value to a string representation
func formatDefault(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case bool:
		return fmt.Sprintf("%t", v)
	case float64:
		return fmt.Sprintf("%g", v)
	case int:
		return fmt.Sprintf("%d", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
