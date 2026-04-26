package git

import (
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	"bots/internal/workspace"
)

type fakeRunner struct {
	outputs []CommandResult
	calls   [][]string
}

func (r *fakeRunner) Run(name string, args ...string) CommandResult {
	call := append([]string{name}, args...)
	r.calls = append(r.calls, call)
	if len(r.outputs) == 0 {
		return CommandResult{}
	}
	out := r.outputs[0]
	r.outputs = r.outputs[1:]
	return out
}

func TestCommitCheckpointStagesAndCommitsProjectState(t *testing.T) {
	runner := &fakeRunner{outputs: []CommandResult{
		{Stdout: ".git"},
		{},
		{Stdout: "[main abc123] checkpoint"},
	}}
	integration := NewIntegration(runner)

	result, err := integration.CommitCheckpoint("checkpoint")
	if err != nil {
		t.Fatalf("CommitCheckpoint returned error: %v", err)
	}
	if result.Message != "checkpoint" {
		t.Fatalf("unexpected commit message: %q", result.Message)
	}

	wantCalls := [][]string{
		{"git", "rev-parse", "--git-dir"},
		{"git", "add", ".bots"},
		{"git", "commit", "-m", "checkpoint"},
	}
	if !reflect.DeepEqual(runner.calls, wantCalls) {
		t.Fatalf("unexpected git calls: %#v", runner.calls)
	}
}

func TestWorkspaceIntegrationStagesDiscoveredProjectState(t *testing.T) {
	root := t.TempDir()
	runner := &fakeRunner{outputs: []CommandResult{
		{Stdout: ".git"},
		{},
		{Stdout: "[main abc123] checkpoint"},
	}}
	integration := NewWorkspaceIntegration(workspace.Workspace{Root: root}, runner)

	_, err := integration.CommitCheckpoint("checkpoint")
	if err != nil {
		t.Fatalf("CommitCheckpoint returned error: %v", err)
	}

	wantAdd := []string{"git", "add", filepath.Join(root, ".bots")}
	if !reflect.DeepEqual(runner.calls[1], wantAdd) {
		t.Fatalf("unexpected add call: %#v", runner.calls[1])
	}
}

func TestBranchInfoReturnsBranchAndCommit(t *testing.T) {
	runner := &fakeRunner{outputs: []CommandResult{
		{Stdout: "feature/demo\n"},
		{Stdout: "abc123\n"},
	}}
	integration := NewIntegration(runner)

	info, err := integration.BranchInfo()
	if err != nil {
		t.Fatalf("BranchInfo returned error: %v", err)
	}
	if info.Branch != "feature/demo" || info.Commit != "abc123" {
		t.Fatalf("unexpected branch info: %#v", info)
	}
}

func TestCommitCheckpointDetectsMissingGitRepository(t *testing.T) {
	runner := &fakeRunner{outputs: []CommandResult{
		{Err: errors.New("not git")},
	}}
	integration := NewIntegration(runner)

	_, err := integration.CommitCheckpoint("checkpoint")
	if !errors.Is(err, ErrNotRepository) {
		t.Fatalf("expected ErrNotRepository, got %v", err)
	}
}
