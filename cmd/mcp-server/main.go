// Command mcp-server provides an MCP (Model Context Protocol) server
// that exposes Confluent CLI commands as invocable skills for Claude Code.
//
// Usage:
//
//	mcp-server [skills.json]
//
// If skills.json path is not provided, defaults to pkg/mcp/skills.json
//
// Phase 4 will add actual MCP protocol handling (stdio/HTTP).
// For now, this demonstrates the runtime works end-to-end.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/confluentinc/cli/v4/pkg/mcp"
)

func main() {
	// TODO Phase 4: Implement MCP protocol server (stdio/HTTP transport)
	// For now, demonstrate basic usage

	// Load skills from default location or command-line arg
	skillsPath := filepath.Join("pkg", "mcp", "skills.json")
	if len(os.Args) > 1 {
		skillsPath = os.Args[1]
	}

	// Initialize MCP server
	server, err := mcp.NewMCPServer(skillsPath)
	if err != nil {
		log.Fatalf("Failed to initialize MCP server: %v", err)
	}

	// Display server info
	manifest := server.Manifest()
	fmt.Printf("MCP server initialized successfully\n")
	fmt.Printf("CLI Version: %s\n", manifest.Metadata.CLIVersion)
	fmt.Printf("Skills: %d\n", len(manifest.Tools))
	fmt.Printf("Generated: %s\n", manifest.Metadata.GeneratedAt)
	fmt.Println()

	// Demonstrate skill execution
	fmt.Println("Executing test skill: confluent_version")
	output, err := server.ExecuteSkill("confluent_version", nil)
	if err != nil {
		log.Fatalf("Skill execution failed: %v", err)
	}

	fmt.Printf("Output:\n%s\n", output)
	fmt.Println("---")
	fmt.Println("MCP server ready. MCP protocol handler will be added in Phase 4.")
}
