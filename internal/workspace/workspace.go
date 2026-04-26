package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var ErrNotFound = errors.New("project workspace not found")

// Workspace is the project workspace that owns Bots project state.
type Workspace struct {
	Root string
}

// Discover finds the nearest ancestor project workspace from start.
// If create is true and no project workspace exists, it creates one at start.
func Discover(start string, create bool) (Workspace, error) {
	if start == "" {
		return Workspace{}, fmt.Errorf("workspace start directory is required")
	}

	absStart, err := filepath.Abs(start)
	if err != nil {
		return Workspace{}, fmt.Errorf("resolve workspace start: %w", err)
	}

	info, err := os.Stat(absStart)
	if err != nil {
		return Workspace{}, fmt.Errorf("stat workspace start: %w", err)
	}
	if !info.IsDir() {
		absStart = filepath.Dir(absStart)
	}

	for dir := absStart; ; dir = filepath.Dir(dir) {
		botsDir := filepath.Join(dir, ".bots")
		if info, err := os.Stat(botsDir); err == nil && info.IsDir() {
			return Workspace{Root: dir}, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}

	if !create {
		return Workspace{}, ErrNotFound
	}

	ws := Workspace{Root: absStart}
	if err := os.MkdirAll(ws.BotsDir(), 0755); err != nil {
		return Workspace{}, fmt.Errorf("create project state directory: %w", err)
	}
	return ws, nil
}

// FromCurrent discovers the project workspace from the current working directory.
func FromCurrent(create bool) (Workspace, error) {
	wd, err := os.Getwd()
	if err != nil {
		return Workspace{}, fmt.Errorf("get current directory: %w", err)
	}
	return Discover(wd, create)
}

func (w Workspace) BotsDir() string {
	return filepath.Join(w.Root, ".bots")
}

func (w Workspace) CheckpointFile() string {
	return filepath.Join(w.BotsDir(), "CHECKPOINTS.md")
}

func (w Workspace) LogsDir() string {
	return filepath.Join(w.BotsDir(), "logs")
}

func (w Workspace) TasksDir() string {
	return filepath.Join(w.BotsDir(), "tasks")
}

func (w Workspace) SkillsDir() string {
	return filepath.Join(w.BotsDir(), "skills")
}

func (w Workspace) SessionPersistenceSkillDir() string {
	return filepath.Join(w.SkillsDir(), "session-persistence")
}

func (w Workspace) AgentsFile() string {
	return filepath.Join(w.BotsDir(), "AGENTS.md")
}

func (w Workspace) RootAgentsFile() string {
	return filepath.Join(w.Root, "AGENTS.md")
}

func (w Workspace) RootClaudeFile() string {
	return filepath.Join(w.Root, "CLAUDE.md")
}

// EnsureProjectStateDirs creates the Bots project state directory layout.
func (w Workspace) EnsureProjectStateDirs() error {
	for _, dir := range []string{w.BotsDir(), w.LogsDir(), w.TasksDir(), w.SessionPersistenceSkillDir()} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create %s: %w", dir, err)
		}
	}
	return nil
}

func (w Workspace) EnsureLogsDir() error {
	if err := os.MkdirAll(w.LogsDir(), 0755); err != nil {
		return fmt.Errorf("create logs directory: %w", err)
	}
	return nil
}

func (w Workspace) EnsureTasksDir() error {
	if err := os.MkdirAll(w.TasksDir(), 0755); err != nil {
		return fmt.Errorf("create tasks directory: %w", err)
	}
	return nil
}
