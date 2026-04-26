package task

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"bots/internal/workspace"
)

func TestStoreCreatesReadsUpdatesAndListsTaskHandoff(t *testing.T) {
	root := t.TempDir()
	ws := workspace.Workspace{Root: root}
	store := NewStore(ws)

	created, err := store.Create("Phase 1.5")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if created.Slug != "phase-1-5" {
		t.Fatalf("expected sanitized slug phase-1-5, got %q", created.Slug)
	}
	if created.Path != filepath.Join(root, ".bots", "tasks", "phase-1-5.md") {
		t.Fatalf("unexpected task path: %q", created.Path)
	}

	read, err := store.Read("phase-1-5")
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}
	if read.Slug != "phase-1-5" || read.Content == "" {
		t.Fatalf("unexpected read result: %#v", read)
	}

	status, err := store.Status("phase-1-5")
	if err != nil {
		t.Fatalf("Status returned error: %v", err)
	}
	if status.Status != StatusPending {
		t.Fatalf("expected status %s, got %s", StatusPending, status.Status)
	}

	updated, err := store.UpdateStatus("phase-1-5", "done")
	if err != nil {
		t.Fatalf("UpdateStatus returned error: %v", err)
	}
	if updated.Status != StatusDone {
		t.Fatalf("expected updated status %s, got %s", StatusDone, updated.Status)
	}

	status, err = store.Status("phase")
	if err != nil {
		t.Fatalf("partial Status returned error: %v", err)
	}
	if status.Status != StatusDone {
		t.Fatalf("expected status %s, got %s", StatusDone, status.Status)
	}

	listed, err := store.List()
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(listed.Tasks) != 1 || listed.Tasks[0].Slug != "phase-1-5" || listed.Tasks[0].Status != StatusDone {
		t.Fatalf("unexpected list result: %#v", listed.Tasks)
	}
}

func TestStoreUpdateStatusRejectsUnknownStatus(t *testing.T) {
	root := t.TempDir()
	store := NewStore(workspace.Workspace{Root: root})
	if _, err := store.Create("review"); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	_, err := store.UpdateStatus("review", "BLOCKED")
	if !errors.Is(err, ErrInvalidStatus) {
		t.Fatalf("expected ErrInvalidStatus, got %v", err)
	}
}

func TestStoreReadMissingTaskReturnsStructuredError(t *testing.T) {
	root := t.TempDir()
	ws := workspace.Workspace{Root: root}
	if err := os.MkdirAll(ws.TasksDir(), 0755); err != nil {
		t.Fatalf("create tasks dir: %v", err)
	}
	store := NewStore(ws)

	_, err := store.Read("missing")
	if !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}
}

func TestStoreReadBlankTaskSlugReturnsStructuredError(t *testing.T) {
	root := t.TempDir()
	store := NewStore(workspace.Workspace{Root: root})

	if _, err := store.Create("Existing Task"); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	_, err := store.Read("   ")
	if !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("expected ErrTaskNotFound, got %v", err)
	}
}
