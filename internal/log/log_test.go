package log

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"bots/internal/workspace"
)

func fixedClock() time.Time {
	return time.Date(2026, 4, 12, 10, 30, 0, 0, time.UTC)
}

func TestStoreStartsAppendsSearchesSummarizesAndListsSessionLog(t *testing.T) {
	root := t.TempDir()
	ws := workspace.Workspace{Root: root}
	store := NewStoreWithClock(ws, fixedClock)

	started, err := store.Start("API Redesign!")
	if err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	if started.Slug != "api-redesign" {
		t.Fatalf("expected slug api-redesign, got %q", started.Slug)
	}
	if started.Path != filepath.Join(root, ".bots", "logs", "2026-04-12-api-redesign.md") {
		t.Fatalf("unexpected log path: %q", started.Path)
	}

	appended, err := store.Append("api", "Decision: keep MCP over stdio")
	if err != nil {
		t.Fatalf("Append returned error: %v", err)
	}
	if appended.Filename != started.Filename {
		t.Fatalf("expected append to target %q, got %q", started.Filename, appended.Filename)
	}

	search, err := store.Search("mcp")
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(search.Matches) != 1 || search.Matches[0].File != started.Filename || search.Matches[0].Content != "- Decision: keep MCP over stdio" {
		t.Fatalf("unexpected search matches: %#v", search.Matches)
	}

	summary, err := store.Summarize("redesign")
	if err != nil {
		t.Fatalf("Summarize returned error: %v", err)
	}
	if len(summary.Decisions) != 1 || summary.Decisions[0].Date != "2026-04-12" || summary.Decisions[0].Content != "- Decision: keep MCP over stdio" {
		t.Fatalf("unexpected summary: %#v", summary.Decisions)
	}

	listed, err := store.List()
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(listed.Logs) != 1 || listed.Logs[0].Filename != started.Filename {
		t.Fatalf("unexpected logs: %#v", listed.Logs)
	}
}

func TestStoreAppendClustersMessagesUnderExistingDateSection(t *testing.T) {
	root := t.TempDir()
	ws := workspace.Workspace{Root: root}
	currentTime := fixedClock()
	store := NewStoreWithClock(ws, func() time.Time { return currentTime })

	started, err := store.Start("Daily Notes")
	if err != nil {
		t.Fatalf("Start returned error: %v", err)
	}

	if _, err := store.Append("daily", "First update"); err != nil {
		t.Fatalf("first Append returned error: %v", err)
	}
	if _, err := store.Append("daily", "Second update"); err != nil {
		t.Fatalf("second Append returned error: %v", err)
	}

	contentBytes, err := os.ReadFile(started.Path)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	content := string(contentBytes)
	if count := strings.Count(content, "## 2026-04-12"); count != 1 {
		t.Fatalf("expected one date section for 2026-04-12, got %d in:\n%s", count, content)
	}
	if !strings.Contains(content, "## 2026-04-12\n\n- Started session on topic: Daily Notes\n- First update\n- Second update\n\n") {
		t.Fatalf("expected same-day updates under one date section, got:\n%s", content)
	}

	currentTime = currentTime.AddDate(0, 0, 1)
	if _, err := store.Append("daily", "Next day update"); err != nil {
		t.Fatalf("next-day Append returned error: %v", err)
	}

	contentBytes, err = os.ReadFile(started.Path)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	content = string(contentBytes)
	if !strings.Contains(content, "\n\n## 2026-04-13\n\n- Next day update\n\n") {
		t.Fatalf("expected new section for a new date, got:\n%s", content)
	}
}

func TestStoreAppendIgnoresHeadingsInsideFencedCodeBlocks(t *testing.T) {
	root := t.TempDir()
	ws := workspace.Workspace{Root: root}
	store := NewStoreWithClock(ws, fixedClock)

	started, err := store.Start("Fence Notes")
	if err != nil {
		t.Fatalf("Start returned error: %v", err)
	}

	message := "Captured output:\n```md\n## not a real section\n```"
	if _, err := store.Append("fence", message); err != nil {
		t.Fatalf("first Append returned error: %v", err)
	}
	if _, err := store.Append("fence", "Second update"); err != nil {
		t.Fatalf("second Append returned error: %v", err)
	}

	contentBytes, err := os.ReadFile(started.Path)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	content := string(contentBytes)
	want := "- Captured output:\n```md\n## not a real section\n```\n- Second update"
	if !strings.Contains(content, want) {
		t.Fatalf("expected append after fenced code block, got:\n%s", content)
	}
}

func TestStoreAppendMissingSessionLogReturnsStructuredError(t *testing.T) {
	root := t.TempDir()
	store := NewStore(workspace.Workspace{Root: root})

	_, err := store.Append("missing", "message")
	if !errors.Is(err, ErrLogNotFound) {
		t.Fatalf("expected ErrLogNotFound, got %v", err)
	}
}

func TestStoreAppendBlankSlugReturnsStructuredError(t *testing.T) {
	root := t.TempDir()
	store := NewStore(workspace.Workspace{Root: root})

	if _, err := store.Start("Existing Log"); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}

	_, err := store.Append("   ", "message")
	if !errors.Is(err, ErrLogNotFound) {
		t.Fatalf("expected ErrLogNotFound, got %v", err)
	}
}
