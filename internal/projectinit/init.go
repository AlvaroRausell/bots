package projectinit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Initialize creates the .bots directory structure for a new project
func Initialize(projectName string) {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting working directory: %v\n", err)
		os.Exit(1)
	}

	dirs := []string{
		filepath.Join(wd, ".bots"),
		filepath.Join(wd, ".bots", "logs"),
		filepath.Join(wd, ".bots", "tasks"),
		filepath.Join(wd, ".bots", "skills", "session-persistence"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory %s: %v\n", dir, err)
			os.Exit(1)
		}
	}

	// Create CHECKPOINTS.md with initial content (pure project state, no instructions)
	checkpointContent := createCheckpointContent(projectName, time.Now().Format("2006-01-02"))

	checkpointPath := filepath.Join(wd, ".bots", "CHECKPOINTS.md")
	if err := os.WriteFile(checkpointPath, []byte(checkpointContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating CHECKPOINTS.md: %v\n", err)
		os.Exit(1)
	}

	// Create AGENTS.md in .bots/ with AI agent instructions
	agentsContent := createAgentsContent()

	agentsPath := filepath.Join(wd, ".bots", "AGENTS.md")
	if err := os.WriteFile(agentsPath, []byte(agentsContent), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating .bots/AGENTS.md: %v\n", err)
		os.Exit(1)
	}

	// Create .gitkeep files in empty directories
	gitkeepFiles := []string{
		filepath.Join(wd, ".bots", "logs", ".gitkeep"),
		filepath.Join(wd, ".bots", "tasks", ".gitkeep"),
	}

	for _, file := range gitkeepFiles {
		if err := os.WriteFile(file, []byte(""), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating %s: %v\n", file, err)
			os.Exit(1)
		}
	}

	// Create root AGENTS.md (pointer to .bots/AGENTS.md)
	rootAgentsContent := createRootAgentsContent()

	rootAgentsPath := filepath.Join(wd, "AGENTS.md")
	if err := createOrAppendFile(rootAgentsPath, rootAgentsContent, "AGENTS.md"); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating AGENTS.md: %v\n", err)
		os.Exit(1)
	}

	// Create root CLAUDE.md (pointer to AGENTS.md)
	rootClaudeContent := createRootClaudeContent()

	rootClaudePath := filepath.Join(wd, "CLAUDE.md")
	if err := createOrAppendFile(rootClaudePath, rootClaudeContent, "CLAUDE.md"); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating CLAUDE.md: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Initialized .bots directory for project: %s\n\n", projectName)
	fmt.Println("Created:")
	fmt.Println("  .bots/AGENTS.md        - AI agent instructions")
	fmt.Println("  .bots/CHECKPOINTS.md   - Living project state")
	fmt.Println("  .bots/logs/            - Session logs")
	fmt.Println("  .bots/tasks/           - Task files")
	fmt.Println("  .bots/skills/          - AI agent skills")
	fmt.Println("  AGENTS.md              - Root pointer to .bots/AGENTS.md")
	fmt.Println("  CLAUDE.md              - Root pointer to AGENTS.md")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Edit .bots/CHECKPOINTS.md to add project details")
	fmt.Println("  2. Edit .bots/AGENTS.md to add project-specific rules")
	fmt.Println("  3. Start your first session: bots log start \"<topic>\"")
	fmt.Println("  4. Run 'bots install' to configure MCP for your AI agent")
}

func createCheckpointContent(projectName, date string) string {
	return fmt.Sprintf(`# %s - Project Checkpoints

> **This is a living project document.** Update it as decisions are made and work progresses.

---

## Current Checkpoint

- **Project**: %s
- **Started**: %s
- **Last session log**: None yet
- **Status**: Project initialized

---

## Project Phases

| Phase | Description | Status | Started | Completed |
|-------|-------------|--------|---------|-----------|
| 1     | Project setup | [ ] | - | - |

---

## Key Decisions

*No decisions recorded yet. Start logging with `+"`"+`bots log append`+"`"+`*

---

## Open Questions

*No open questions yet.*

---
`, projectName, projectName, date)
}

func createAgentsContent() string {
	return `# AI Agent Instructions

This project uses the **bots** framework for AI agent coordination.

## Quick Start

Before doing anything, read:
1. This file (` + "`.bots/AGENTS.md`" + `) for instructions
2. ` + "`.bots/CHECKPOINTS.md`" + ` for current project state

## Directory Structure

- ` + "`.bots/AGENTS.md`" + ` - AI agent instructions (this file)
- ` + "`.bots/CHECKPOINTS.md`" + ` - Living project state document
- ` + "`.bots/logs/`" + ` - Session decision logs
- ` + "`.bots/tasks/`" + ` - Task handoff files
- ` + "`.bots/skills/`" + ` - AI agent skills

## Workflow

### Before Starting Work

1. Read ` + "`.bots/CHECKPOINTS.md`" + ` for current project state
2. Check ` + "`.bots/logs/`" + ` for recent session context
3. Start a new session log: ` + "`bots log start <topic>`" + `

### During Work

1. Log decisions as they are made: ` + "`bots log append <slug> \"Decision: ...\"`" + `
2. Use skills from ` + "`.bots/skills/`" + ` as needed

### When Completing Work

1. Update ` + "`.bots/CHECKPOINTS.md`" + ` with new state
2. Link to the session log with decisions
3. Commit changes: ` + "`bots git_commit_checkpoint \"message\"`" + `

## Project Rules

### Code Style

- Follow existing code conventions
- Keep functions small and focused
- Add comments for non-obvious decisions

### Git Workflow

- Commit checkpoints when project state changes
- Link session logs in commit messages when relevant
- Use descriptive commit messages

---

*Add project-specific rules in this section as needed.*
`
}

func createRootAgentsContent() string {
	return `# AI Agent Instructions

This project uses [bots](https://github.com/example/bots) for AI agent coordination.

See [` + "`.bots/AGENTS.md`" + `](` + "`.bots/AGENTS.md`" + `) for instructions.
`
}

func createRootClaudeContent() string {
	return `# Claude Instructions

See [AGENTS.md](AGENTS.md) for AI agent instructions.
`
}

// createOrAppendFile creates a file or appends bots instructions if file exists
// For root AGENTS.md, checks for .bots/AGENTS.md reference
// For root CLAUDE.md, checks for AGENTS.md reference
func createOrAppendFile(path, content, filename string) error {
	existing, err := os.ReadFile(path)
	if err != nil {
		// File doesn't exist, create it
		return os.WriteFile(path, []byte(content), 0644)
	}

	// File exists - check if it already has the appropriate marker
	existingStr := string(existing)
	if filename == "AGENTS.md" && strings.Contains(existingStr, ".bots/AGENTS.md") {
		fmt.Printf("  ℹ️  %s already exists with bots reference, skipping\n", filename)
		return nil
	}
	if filename == "CLAUDE.md" && strings.Contains(existingStr, "AGENTS.md") {
		fmt.Printf("  ℹ️  %s already exists with AGENTS.md reference, skipping\n", filename)
		return nil
	}

	// Append bots instructions with a separator
	appendedContent := fmt.Sprintf("\n\n---\n\n%s", content)
	if err := os.WriteFile(path, append(existing, []byte(appendedContent)...), 0644); err != nil {
		return err
	}

	fmt.Printf("  ℹ️  Appended bots instructions to existing %s\n", filename)
	return nil
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