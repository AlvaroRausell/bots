package cli

import (
	"fmt"
	"os"

	"bots/internal/projectinit"
)

func InitCommand(args []string) {
	if len(args) > 0 && args[0] == "help" {
		printInitUsage()
		return
	}

	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: project name required")
		fmt.Fprintln(os.Stderr, "Usage: bots init <project-name>")
		os.Exit(1)
	}

	projectName := args[0]
	projectinit.Initialize(projectName)
}

func printInitUsage() {
	fmt.Println(`Project Initialization

Usage:
  bots init <project-name>

Initializes a new project with the .bots directory structure:
  - .bots/AGENTS.md         # AI agent instructions
  - .bots/CHECKPOINTS.md    # Living project state + startup checklist
  - .bots/logs/             # Session decision logs
  - .bots/tasks/            # Task handoff files
  - .bots/skills/           # AI agent skills

The generated checkpoint requires planning first:
  - Fill the Project Startup Checklist
  - Define Project Phases
  - Then start the first session log

Also creates root-level entry points:
  - AGENTS.md               # Points to .bots/AGENTS.md
  - CLAUDE.md               # Points to AGENTS.md

The project name is used to create an initial checkpoint entry.

Examples:
  bots init "my-app"
  bots init "travel-guide"
  bots init "api-redesign"`)
}
