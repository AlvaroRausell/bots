package projectinit

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Initialize creates the .bots directory structure for a new project
func Initialize(projectName string) {
	// Create directory structure
	dirs := []string{
		".bots",
		".bots/logs",
		".bots/tasks",
		".bots/skills/session-persistence",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory %s: %v\n", dir, err)
			os.Exit(1)
		}
	}

	// Create CHECKPOINTS.md with initial content
	checkpointContent := fmt.Sprintf(`# %s - Project Checkpoints

> **This is a living project document.** Update it as decisions are made and work progresses.

---

## Current Checkpoint

- **Project**: %s
- **Started**: %s
- **Last session log**: None yet
- **Status**: Project initialized

---

## Agent Instructions

**Before starting work:**
1. Read this checkpoint file for current project state
2. Check .bots/logs/ for recent session context
3. Start a new session log: `+"`"+"bots log start <topic>`"+`
4. Log decisions as they are made: `+"`"+"bots log append <slug> \"Decision: ...\"`"+`

**When completing work:**
1. Update this checkpoint file with new state
2. Link to the session log with decisions
3. Commit changes: `+"`"+"bots git_commit_checkpoint \"message\"`"+`

---

## Project Phases

| Phase | Description | Status | Started | Completed |
|-------|-------------|--------|---------|-----------|
| 1     | Project setup | [ ] | - | - |

---

## Key Decisions

*No decisions recorded yet. Start logging with `+"`"+"bots log append`"+`*

---

## Open Questions

*No open questions yet.*

---
`, projectName, projectName, time.Now().Format("2006-01-02"))

	if err := os.WriteFile(".bots/CHECKPOINTS.md", []byte(checkpointContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating CHECKPOINTS.md: %v\n", err)
		os.Exit(1)
	}

	// Create RULES.md with default content
	rulesContent := `# Project Rules

## Code Style

- Follow existing code conventions
- Keep functions small and focused
- Add comments for non-obvious decisions

## Git Workflow

- Commit checkpoints when project state changes
- Link session logs in commit messages when relevant
- Use descriptive commit messages

## AI Agent Guidelines

- Always read CHECKPOINTS.md before starting work
- Log decisions in real-time during sessions
- Update checkpoint when work is complete
- Use task handoff protocol for complex work

---

*Add project-specific rules below*

`

	if err := os.WriteFile(".bots/RULES.md", []byte(rulesContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating RULES.md: %v\n", err)
		os.Exit(1)
	}

	// Create .gitkeep files in empty directories
	gitkeepFiles := []string{
		".bots/logs/.gitkeep",
		".bots/tasks/.gitkeep",
	}

	for _, file := range gitkeepFiles {
		if err := os.WriteFile(file, []byte(""), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating %s: %v\n", file, err)
			os.Exit(1)
		}
	}

	fmt.Printf("✅ Initialized .bots directory for project: %s\n\n", projectName)
	fmt.Println("Created:")
	fmt.Println("  .bots/CHECKPOINTS.md    - Living project state")
	fmt.Println("  .bots/RULES.md          - Project rules")
	fmt.Println("  .bots/logs/             - Session logs")
	fmt.Println("  .bots/tasks/            - Task files")
	fmt.Println("  .bots/skills/           - AI agent skills")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Edit .bots/CHECKPOINTS.md to add project details")
	fmt.Println("  2. Edit .bots/RULES.md to add project-specific rules")
	fmt.Println("  3. Start your first session: bots log start \"<topic>\"")
	fmt.Println("  4. Run 'bots install' to configure MCP for your AI agent")
}

// GetProjectRoot returns the path to the .bots directory
func GetProjectRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}

	dir := wd
	for {
		botsDir := filepath.Join(dir, ".bots")
		if _, err := os.Stat(botsDir); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
