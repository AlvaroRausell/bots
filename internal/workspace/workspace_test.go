package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverFindsProjectWorkspaceFromNestedDirectory(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".bots"), 0755); err != nil {
		t.Fatalf("create .bots: %v", err)
	}
	nested := filepath.Join(root, "a", "b")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatalf("create nested dir: %v", err)
	}

	ws, err := Discover(nested, false)
	if err != nil {
		t.Fatalf("Discover returned error: %v", err)
	}

	if ws.Root != root {
		t.Fatalf("expected root %q, got %q", root, ws.Root)
	}
	if ws.CheckpointFile() != filepath.Join(root, ".bots", "CHECKPOINTS.md") {
		t.Fatalf("unexpected checkpoint path: %q", ws.CheckpointFile())
	}
}

func TestDiscoverCreatesProjectWorkspaceAtStartWhenMissing(t *testing.T) {
	root := t.TempDir()

	ws, err := Discover(root, true)
	if err != nil {
		t.Fatalf("Discover returned error: %v", err)
	}

	if ws.Root != root {
		t.Fatalf("expected root %q, got %q", root, ws.Root)
	}
	if _, err := os.Stat(filepath.Join(root, ".bots")); err != nil {
		t.Fatalf("expected .bots directory to be created: %v", err)
	}
}

func TestWorkspaceEnsuresProjectStateDirectories(t *testing.T) {
	root := t.TempDir()
	ws := Workspace{Root: root}

	if err := ws.EnsureProjectStateDirs(); err != nil {
		t.Fatalf("EnsureProjectStateDirs returned error: %v", err)
	}

	for _, dir := range []string{ws.BotsDir(), ws.LogsDir(), ws.TasksDir(), ws.SessionPersistenceSkillDir()} {
		info, err := os.Stat(dir)
		if err != nil {
			t.Fatalf("expected %s to exist: %v", dir, err)
		}
		if !info.IsDir() {
			t.Fatalf("expected %s to be a directory", dir)
		}
	}
}
