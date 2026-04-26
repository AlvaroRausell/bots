package git

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"bots/internal/workspace"
)

var ErrNotRepository = errors.New("not in a git repository")

type CommandResult struct {
	Stdout string
	Stderr string
	Err    error
}

type CommandRunner interface {
	Run(name string, args ...string) CommandResult
}

type ExecRunner struct{}

func (ExecRunner) Run(name string, args ...string) CommandResult {
	cmd := exec.Command(name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return CommandResult{Stdout: stdout.String(), Stderr: stderr.String(), Err: err}
}

// Integration manages git operations for Bots project state.
type Integration struct {
	runner           CommandRunner
	projectStatePath string
}

type CommitResult struct {
	Message string
	Output  string
}

type BranchInfo struct {
	Branch string
	Commit string
}

func NewIntegration(runner CommandRunner) Integration {
	if runner == nil {
		runner = ExecRunner{}
	}
	return Integration{runner: runner, projectStatePath: ".bots"}
}

func NewWorkspaceIntegration(ws workspace.Workspace, runner CommandRunner) Integration {
	integration := NewIntegration(runner)
	integration.projectStatePath = ws.BotsDir()
	return integration
}

func DefaultIntegration() Integration {
	return NewIntegration(ExecRunner{})
}

func (i Integration) CommitCheckpoint(message string) (CommitResult, error) {
	if result := i.runner.Run("git", "rev-parse", "--git-dir"); result.Err != nil {
		return CommitResult{}, ErrNotRepository
	}

	projectStatePath := i.projectStatePath
	if projectStatePath == "" {
		projectStatePath = ".bots"
	}

	if result := i.runner.Run("git", "add", projectStatePath); result.Err != nil {
		return CommitResult{}, commandError("stage checkpoint", result)
	}

	result := i.runner.Run("git", "commit", "-m", message)
	if result.Err != nil {
		return CommitResult{}, commandError("commit checkpoint", result)
	}

	return CommitResult{Message: message, Output: strings.TrimSpace(result.Stdout)}, nil
}

func (i Integration) BranchInfo() (BranchInfo, error) {
	branchResult := i.runner.Run("git", "rev-parse", "--abbrev-ref", "HEAD")
	if branchResult.Err != nil {
		return BranchInfo{}, ErrNotRepository
	}

	commitResult := i.runner.Run("git", "rev-parse", "--short", "HEAD")
	if commitResult.Err != nil {
		return BranchInfo{Branch: strings.TrimSpace(branchResult.Stdout)}, commandError("read current commit", commitResult)
	}

	return BranchInfo{
		Branch: strings.TrimSpace(branchResult.Stdout),
		Commit: strings.TrimSpace(commitResult.Stdout),
	}, nil
}

func commandError(action string, result CommandResult) error {
	stderr := strings.TrimSpace(result.Stderr)
	if stderr == "" {
		return fmt.Errorf("%s: %w", action, result.Err)
	}
	return fmt.Errorf("%s: %w\n%s", action, result.Err, stderr)
}

func FormatCommitResult(result CommitResult) string {
	return "Committed checkpoint changes successfully"
}

func FormatBranchInfo(info BranchInfo) string {
	if info.Commit == "" {
		return info.Branch
	}
	return fmt.Sprintf("Branch: %s, Commit: %s", info.Branch, info.Commit)
}
