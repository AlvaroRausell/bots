package cli

import (
	"fmt"
	"os"

	"bots/internal/mcp"
)

func MCPCommand(args []string) {
	if len(args) < 1 {
		printMCPUsage()
		os.Exit(1)
	}

	subcommand := args[0]

	switch subcommand {
	case "serve":
		mcp.Serve()
	case "help":
		printMCPUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown mcp subcommand: %s\n", subcommand)
		printMCPUsage()
		os.Exit(1)
	}
}

func printMCPUsage() {
	fmt.Println(`MCP Server

Usage:
  bots mcp <subcommand>

Subcommands:
  serve                      Start MCP server (stdio)

Examples:
  bots mcp serve             Start MCP server for AI agent integration

The MCP server provides programmatic access to:
  - Checkpoint management tools
  - Session log tools
  - Task handoff tools
  - Git integration tools

Configure your AI agent to use this MCP server for structured
project state management.`)
}
