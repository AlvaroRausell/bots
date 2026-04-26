package checkpoint

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"bots/internal/workspace"
)

func TestStoreUpdateReadAndListCheckpointSections(t *testing.T) {
	root := t.TempDir()
	ws := workspace.Workspace{Root: root}
	if err := os.Mkdir(ws.BotsDir(), 0755); err != nil {
		t.Fatalf("create .bots: %v", err)
	}
	store := NewStore(ws)

	updated, err := store.Update("Current Checkpoint", "- Ready")
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated.Section != "Current Checkpoint" {
		t.Fatalf("unexpected updated section: %q", updated.Section)
	}

	read, err := store.Read()
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}
	if !read.Found {
		t.Fatal("expected checkpoint to be found")
	}
	wantContent := "## Current Checkpoint\n\n- Ready\n"
	if read.Content != wantContent {
		t.Fatalf("expected %q, got %q", wantContent, read.Content)
	}

	listed, err := store.List()
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if !reflect.DeepEqual(listed.Sections, []string{"Current Checkpoint"}) {
		t.Fatalf("unexpected sections: %v", listed.Sections)
	}
}

func TestStoreReadMissingCheckpointIsStructuredOutcome(t *testing.T) {
	root := t.TempDir()
	ws := workspace.Workspace{Root: root}
	if err := os.Mkdir(ws.BotsDir(), 0755); err != nil {
		t.Fatalf("create .bots: %v", err)
	}
	store := NewStore(ws)

	read, err := store.Read()
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}
	if read.Found {
		t.Fatal("expected missing checkpoint")
	}
	if read.Path != filepath.Join(root, ".bots", "CHECKPOINTS.md") {
		t.Fatalf("unexpected checkpoint path: %q", read.Path)
	}
}
