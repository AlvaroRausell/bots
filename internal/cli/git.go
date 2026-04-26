package cli

import (
	"fmt"
	"os"
	"strings"

	gitintegration "bots/internal/git"
	"bots/internal/workspace"
)

func GitCommitCheckpointCommand(args []string) {
	if len(args) > 0 && args[0] == "help" {
		printGitCommitCheckpointUsage()
		return
	}
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: commit message required")
		fmt.Fprintln(os.Stderr, "Usage: bots git_commit_checkpoint <message>")
		os.Exit(1)
	}

	ws, err := workspace.FromCurrent(false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding project workspace: %v\n", err)
		os.Exit(1)
	}

	result, err := gitintegration.NewWorkspaceIntegration(ws, nil).CommitCheckpoint(strings.Join(args, " "))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error committing checkpoint: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(gitintegration.FormatCommitResult(result))
}

func GitBranchInfoCommand(args []string) {
	if len(args) > 0 && args[0] == "help" {
		printGitBranchInfoUsage()
		return
	}

	info, err := gitintegration.DefaultIntegration().BranchInfo()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading git branch: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(gitintegration.FormatBranchInfo(info))
}

func printGitCommitCheckpointUsage() {
	fmt.Println(`Git Commit Checkpoint

Usage:
  bots git_commit_checkpoint <message>

Stages .bots project state and creates a git commit.`)
}

func printGitBranchInfoUsage() {
	fmt.Println(`Git Branch Info

Usage:
  bots git_branch_info

Prints the current git branch and short commit.`)
}
