package projectinit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"bots/internal/workspace"
)

func TestInitializerCreatesProjectStateAndRootPointers(t *testing.T) {
	root := t.TempDir()
	ws := workspace.Workspace{Root: root}
	initializer := NewInitializer(ws, func() time.Time {
		return time.Date(2026, 4, 12, 0, 0, 0, 0, time.UTC)
	})

	result, err := initializer.Initialize("my-app")
	if err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}
	if result.ProjectName != "my-app" {
		t.Fatalf("unexpected project name: %q", result.ProjectName)
	}

	for _, path := range []string{
		ws.CheckpointFile(),
		ws.AgentsFile(),
		filepath.Join(ws.LogsDir(), ".gitkeep"),
		filepath.Join(ws.TasksDir(), ".gitkeep"),
		ws.RootAgentsFile(),
		ws.RootClaudeFile(),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected %s to exist: %v", path, err)
		}
	}

	checkpoint, err := os.ReadFile(ws.CheckpointFile())
	if err != nil {
		t.Fatalf("read checkpoint: %v", err)
	}
	if !strings.Contains(string(checkpoint), "- **Project**: my-app") || !strings.Contains(string(checkpoint), "- **Started**: 2026-04-12") {
		t.Fatalf("checkpoint content missing project metadata:\n%s", checkpoint)
	}
}

func TestInitializerAppendsRootPointersWithoutDuplicating(t *testing.T) {
	root := t.TempDir()
	ws := workspace.Workspace{Root: root}
	if err := os.WriteFile(ws.RootAgentsFile(), []byte("# Existing\n"), 0644); err != nil {
		t.Fatalf("write existing AGENTS.md: %v", err)
	}

	initializer := NewInitializer(ws, nil)
	if _, err := initializer.Initialize("my-app"); err != nil {
		t.Fatalf("Initialize returned error: %v", err)
	}
	if _, err := initializer.Initialize("my-app"); err != nil {
		t.Fatalf("second Initialize returned error: %v", err)
	}

	content, err := os.ReadFile(ws.RootAgentsFile())
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	if strings.Count(string(content), "This project uses [bots]") != 1 {
		t.Fatalf("expected one bots pointer block, got:\n%s", content)
	}
}
