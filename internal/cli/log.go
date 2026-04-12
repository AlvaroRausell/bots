package cli

import (
	"fmt"
	"os"

	"bots/internal/log"
)

func LogCommand(args []string) {
	if len(args) < 1 {
		printLogUsage()
		os.Exit(1)
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "start":
		if len(subArgs) < 1 {
			fmt.Fprintln(os.Stderr, "Error: topic required")
			fmt.Fprintln(os.Stderr, "Usage: bots log start <topic>")
			os.Exit(1)
		}
		log.StartLog(subArgs[0])
	case "append":
		if len(subArgs) < 2 {
			fmt.Fprintln(os.Stderr, "Error: file and message required")
			fmt.Fprintln(os.Stderr, "Usage: bots log append <file> <message>")
			os.Exit(1)
		}
		log.AppendEntry(subArgs[0], subArgs[1])
	case "search":
		if len(subArgs) < 1 {
			fmt.Fprintln(os.Stderr, "Error: query required")
			fmt.Fprintln(os.Stderr, "Usage: bots log search <query>")
			os.Exit(1)
		}
		log.SearchLogs(subArgs[0])
	case "summarize":
		if len(subArgs) < 1 {
			fmt.Fprintln(os.Stderr, "Error: file required")
			fmt.Fprintln(os.Stderr, "Usage: bots log summarize <file>")
			os.Exit(1)
		}
		log.SummarizeLog(subArgs[0])
	case "list":
		log.ListLogs()
	case "help":
		printLogUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown log subcommand: %s\n", subcommand)
		printLogUsage()
		os.Exit(1)
	}
}

func printLogUsage() {
	fmt.Println(`Log Management

Usage:
  bots log <subcommand> [options]

Subcommands:
  start <topic>              Start a new session log
  append <file> <message>    Append entry to session log
  search <query>             Search across all logs
  summarize <file>           Generate summary of decisions
  list                       List all session logs

Examples:
  bots log start "feature-x"
  bots log append feature-x "Decision: using MapLibre for maps"
  bots log search "api design"
  bots log summarize feature-x
  bots log list`)
}
