package projectinit

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"bots/internal/workspace"
)

// Initializer creates Bots project state in a project workspace.
type Initializer struct {
	workspace workspace.Workspace
	now       func() time.Time
}

type InitResult struct {
	ProjectName string
	Workspace   workspace.Workspace
	Created     []string
	Notes       []string
}

func NewInitializer(ws workspace.Workspace, now func() time.Time) Initializer {
	if now == nil {
		now = time.Now
	}
	return Initializer{workspace: ws, now: now}
}

// Initialize creates the .bots directory structure for a new project from cwd.
func Initialize(projectName string) {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting working directory: %v\n", err)
		os.Exit(1)
	}

	ws := workspace.Workspace{Root: wd}
	result, err := NewInitializer(ws, time.Now).Initialize(projectName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing project: %v\n", err)
		os.Exit(1)
	}

	PrintInitResult(result)
}

func (i Initializer) Initialize(projectName string) (InitResult, error) {
	if projectName == "" {
		return InitResult{}, fmt.Errorf("project name is required")
	}

	result := InitResult{ProjectName: projectName, Workspace: i.workspace}

	if err := i.workspace.EnsureProjectStateDirs(); err != nil {
		return InitResult{}, err
	}
	result.Created = append(result.Created,
		".bots/AGENTS.md",
		".bots/CHECKPOINTS.md",
		".bots/logs/",
		".bots/tasks/",
		".bots/skills/",
	)

	checkpointContent := createCheckpointContent(projectName, i.now().Format("2006-01-02"))
	if err := os.WriteFile(i.workspace.CheckpointFile(), []byte(checkpointContent), 0644); err != nil {
		return InitResult{}, fmt.Errorf("create CHECKPOINTS.md: %w", err)
	}

	if err := os.WriteFile(i.workspace.AgentsFile(), []byte(createAgentsContent()), 0644); err != nil {
		return InitResult{}, fmt.Errorf("create .bots/AGENTS.md: %w", err)
	}

	for _, file := range []string{
		filepath.Join(i.workspace.LogsDir(), ".gitkeep"),
		filepath.Join(i.workspace.TasksDir(), ".gitkeep"),
	} {
		if err := os.WriteFile(file, []byte(""), 0644); err != nil {
			return InitResult{}, fmt.Errorf("create %s: %w", file, err)
		}
	}

	if note, err := createOrAppendFile(i.workspace.RootAgentsFile(), createRootAgentsContent(), "AGENTS.md"); err != nil {
		return InitResult{}, fmt.Errorf("create AGENTS.md: %w", err)
	} else if note != "" {
		result.Notes = append(result.Notes, note)
	}
	result.Created = append(result.Created, "AGENTS.md")

	if note, err := createOrAppendFile(i.workspace.RootClaudeFile(), createRootClaudeContent(), "CLAUDE.md"); err != nil {
		return InitResult{}, fmt.Errorf("create CLAUDE.md: %w", err)
	} else if note != "" {
		result.Notes = append(result.Notes, note)
	}
	result.Created = append(result.Created, "CLAUDE.md")

	return result, nil
}

func PrintInitResult(result InitResult) {
	fmt.Printf("✅ Initialized .bots directory for project: %s\n\n", result.ProjectName)
	for _, note := range result.Notes {
		fmt.Println(note)
	}
	if len(result.Notes) > 0 {
		fmt.Println()
	}
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
	fmt.Println("  1. Fill out the Project Startup Checklist in .bots/CHECKPOINTS.md")
	fmt.Println("  2. Define your Project Phases before starting implementation work")
	fmt.Println("  3. Edit .bots/AGENTS.md to add project-specific rules")
	fmt.Println("  4. Start your first session only after planning is current: bots log start \"<topic>\"")
	fmt.Println("  5. Run 'bots install' to configure MCP for your AI agent")
}

func createCheckpointContent(projectName, date string) string {
	return fmt.Sprintf(`# %s - Project Checkpoints

> **This is a living project document.** Update it as decisions are made and work progresses.

---

## Current Checkpoint

- **Project**: %s
- **Started**: %s
- **Last session log**: None yet
- **Status**: Planning required before first work session

---

## Project Startup Checklist

> Complete this checklist before starting the first implementation or logging session.

- [ ] Project goal and desired outcome are clearly defined
- [ ] Scope and non-goals are documented
- [ ] Project stages/phases are listed in the Project Phases section
- [ ] Initial acceptance criteria or definition of done are captured
- [ ] Risks, dependencies, and open questions are documented
- [ ] Initial tasks have been outlined in `+"`"+`.bots/tasks/`+"`"+` if work is ready to begin

---

## Project Phases

| Phase | Description | Status | Started | Completed |
|-------|-------------|--------|---------|-----------|
| 1     | Planning and stage definition | [ ] | - | - |
| 2     | Project setup | [ ] | - | - |

---

## Key Decisions

*No decisions recorded yet. Start logging with `+"`"+`bots log append`+"`"+` after the startup checklist is complete.*

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
2. Complete or update the ` + "`Project Startup Checklist`" + ` before beginning a new project or major initiative
3. Define or revise the ` + "`Project Phases`" + ` so the stages are clear before implementation
4. Review ` + "`.bots/logs/`" + ` and ` + "`.bots/tasks/`" + ` for recent context when they exist
5. Start a new session log only after planning is current: ` + "`bots log start <topic>`" + `

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

This project uses [bots](https://github.com/AlvaroRausell/bots) for AI agent coordination.

See [` + "`.bots/AGENTS.md`" + `](` + "`.bots/AGENTS.md`" + `) for instructions.
`
}

func createRootClaudeContent() string {
	return `# Claude Instructions

See [AGENTS.md](AGENTS.md) for AI agent instructions.
`
}

// createOrAppendFile creates a file or appends bots instructions if file exists.
func createOrAppendFile(path, content, filename string) (string, error) {
	existing, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", os.WriteFile(path, []byte(content), 0644)
		}
		return "", err
	}

	existingStr := string(existing)
	if filename == "AGENTS.md" && strings.Contains(existingStr, ".bots/AGENTS.md") {
		return fmt.Sprintf("  ℹ️  %s already exists with bots reference, skipping", filename), nil
	}
	if filename == "CLAUDE.md" && strings.Contains(existingStr, "AGENTS.md") {
		return fmt.Sprintf("  ℹ️  %s already exists with AGENTS.md reference, skipping", filename), nil
	}

	appendedContent := fmt.Sprintf("\n\n---\n\n%s", content)
	if err := os.WriteFile(path, append(existing, []byte(appendedContent)...), 0644); err != nil {
		return "", err
	}

	return fmt.Sprintf("  ℹ️  Appended bots instructions to existing %s", filename), nil
}

// GetProjectRoot returns the root path of the discovered project workspace.
func GetProjectRoot() string {
	ws, err := workspace.FromCurrent(false)
	if err != nil {
		return ""
	}
	return ws.Root
}
