package main

import (
	"fmt"
	"os"

	"bots/internal/cli"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "log":
		cli.LogCommand(args)
	case "checkpoint":
		cli.CheckpointCommand(args)
	case "task":
		cli.TaskCommand(args)
	case "mcp":
		cli.MCPCommand(args)
	case "install":
		cli.InstallCommand(args)
	case "init":
		cli.InitCommand(args)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Bots - Session Persistence & Decision Tracking

Usage:
  bots <command> [subcommand] [options]

Commands:
  init        Initialize new project
  log         Session log management
  checkpoint  Checkpoint management
  task        Task handoff protocol
  mcp         MCP server
  install     Install MCP to AI agents

Examples:
  bots log start "feature-x"           Start a new session log
  bots log append feature-x "message"  Append to session log
  bots checkpoint read                 Read current checkpoint
  bots task create "phase-1.5"         Create task file
  bots mcp serve                       Start MCP server

Run 'bots <command> help' for more information on a command.`)
}
