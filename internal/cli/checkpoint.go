package cli

import (
	"fmt"
	"os"

	"bots/internal/checkpoint"
)

func CheckpointCommand(args []string) {
	if len(args) < 1 {
		printCheckpointUsage()
		os.Exit(1)
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "read":
		checkpoint.Read()
	case "update":
		if len(subArgs) < 2 {
			fmt.Fprintln(os.Stderr, "Error: section and content required")
			fmt.Fprintln(os.Stderr, "Usage: bots checkpoint update <section> <content>")
			os.Exit(1)
		}
		checkpoint.Update(subArgs[0], subArgs[1])
	case "list":
		checkpoint.List()
	case "help":
		printCheckpointUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown checkpoint subcommand: %s\n", subcommand)
		printCheckpointUsage()
		os.Exit(1)
	}
}

func printCheckpointUsage() {
	fmt.Println(`Checkpoint Management

Usage:
  bots checkpoint <subcommand> [options]

Subcommands:
  read                       Read current checkpoint state
  update <section> <content> Update a checkpoint section
  list                       List all checkpoints

Examples:
  bots checkpoint read
  bots checkpoint update "Current Checkpoint" "New content"
  bots checkpoint list`)
}
