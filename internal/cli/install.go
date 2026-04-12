package cli

import (
	"fmt"

	"bots/internal/install"
)

func InstallCommand(args []string) {
	if len(args) > 0 && args[0] == "help" {
		printInstallUsage()
		return
	}

	install.Run()
}

func printInstallUsage() {
	fmt.Println(`MCP Installation

Usage:
  bots install

Opens an interactive TUI to install the MCP server to AI agents:
  - opencode
  - Claude Code
  - Codex

The installer will:
  1. Detect which agents are available on your system
  2. Let you select which agents to configure
  3. Add the MCP server configuration to each agent
  4. Verify the installation

Run without arguments to start the interactive installer.`)
}
