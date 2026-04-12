package cli

import (
	"fmt"
	"os"

	"bots/internal/task"
)

func TaskCommand(args []string) {
	if len(args) < 1 {
		printTaskUsage()
		os.Exit(1)
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "create":
		if len(subArgs) < 1 {
			fmt.Fprintln(os.Stderr, "Error: task slug required")
			fmt.Fprintln(os.Stderr, "Usage: bots task create <slug>")
			os.Exit(1)
		}
		task.Create(subArgs[0])
	case "read":
		if len(subArgs) < 1 {
			fmt.Fprintln(os.Stderr, "Error: task slug required")
			fmt.Fprintln(os.Stderr, "Usage: bots task read <slug>")
			os.Exit(1)
		}
		task.Read(subArgs[0])
	case "status":
		if len(subArgs) < 1 {
			fmt.Fprintln(os.Stderr, "Error: task slug required")
			fmt.Fprintln(os.Stderr, "Usage: bots task status <slug>")
			os.Exit(1)
		}
		task.Status(subArgs[0])
	case "update":
		if len(subArgs) < 2 {
			fmt.Fprintln(os.Stderr, "Error: task slug and status required")
			fmt.Fprintln(os.Stderr, "Usage: bots task update <slug> <status>")
			os.Exit(1)
		}
		task.UpdateStatus(subArgs[0], subArgs[1])
	case "list":
		task.List()
	case "help":
		printTaskUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown task subcommand: %s\n", subcommand)
		printTaskUsage()
		os.Exit(1)
	}
}

func printTaskUsage() {
	fmt.Println(`Task Handoff Protocol

Usage:
  bots task <subcommand> [options]

Subcommands:
  create <slug>              Create a new task file
  read <slug>                Read task file content
  status <slug>              Check task status
  update <slug> <status>     Update task status
  list                       List all tasks

Valid Status Values:
  PENDING             Task created, not yet started
  IN_PROGRESS         Actively working
  READY_FOR_REVIEW    Complete, awaiting review
  CHANGES_REQUESTED   Review feedback received
  DONE                Task completed and approved

Examples:
  bots task create "phase-1.5"
  bots task status "phase-1.5"
  bots task update "phase-1.5" "READY_FOR_REVIEW"
  bots task list`)
}
