package install

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Agent represents an AI agent that can use MCP servers
type Agent struct {
	Name        string
	Description string
	ConfigDir   string
	Installed   bool
	Selected    bool
}

// Model represents the TUI state
type Model struct {
	agents     []Agent
	cursor     int
	installing bool
	done       bool
	message    string
}

// Available agents
func getAgents() []Agent {
	homeDir, _ := os.UserHomeDir()

	agents := []Agent{
		{
			Name:        "opencode",
			Description: "OpenCode AI agent",
			ConfigDir:   filepath.Join(homeDir, ".opencode"),
			Installed:   false,
			Selected:    true,
		},
		{
			Name:        "claude",
			Description: "Claude Code (Anthropic)",
			ConfigDir:   filepath.Join(homeDir, ".claude"),
			Installed:   false,
			Selected:    false,
		},
		{
			Name:        "codex",
			Description: "OpenAI Codex",
			ConfigDir:   filepath.Join(homeDir, ".codex"),
			Installed:   false,
			Selected:    false,
		},
	}

	// Check which agents are installed
	for i := range agents {
		// Check if config dir exists
		if _, err := os.Stat(agents[i].ConfigDir); err == nil {
			agents[i].Installed = true
		}

		// Also check if the agent CLI exists
		agents[i].Installed = agents[i].Installed || commandExists(agents[i].Name)
	}

	return agents
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// Initial model creation
func InitialModel() Model {
	return Model{
		agents:     getAgents(),
		cursor:     0,
		installing: false,
		done:       false,
		message:    "",
	}
}

// Init returns the initial command
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.done {
		if msg, ok := msg.(tea.KeyMsg); ok && msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		return m, nil
	}

	if m.installing {
		// Check for completion message
		if _, ok := msg.(installResultMsg); ok {
			m.done = true
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.agents)-1 {
				m.cursor++
			}

		case "enter", " ":
			if m.cursor >= 0 && m.cursor < len(m.agents) {
				m.agents[m.cursor].Selected = !m.agents[m.cursor].Selected
			}

		case "i", "I":
			m.installing = true
			return m, installMCP(m.agents)
		}
	}

	return m, nil
}

// View renders the UI
func (m Model) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("🤖 Bots MCP Installer\n\n"))
	b.WriteString(subtitleStyle.Render("Select agents to configure with MCP server\n\n"))

	for i, agent := range m.agents {
		cursor := "  "
		if i == m.cursor {
			cursor = cursorStyle.Render("❯ ")
		}

		checkbox := "☐"
		if agent.Selected {
			checkbox = checkboxStyle.Render("☑")
		}

		status := ""
		if agent.Installed {
			status = installedStyle.Render(" (installed)")
		}

		line := fmt.Sprintf("%s%s %s%s%s\n",
			cursor,
			checkbox,
			agentStyle.Render(agent.Name),
			status,
			descStyle.Render(" - "+agent.Description))
		b.WriteString(line)
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("\n↑/j ↓/k: Navigate  •  Space/Enter: Toggle  •  i: Install  •  Ctrl+C: Quit\n\n"))

	if m.installing && !m.done {
		b.WriteString(installingStyle.Render("⏳ Installing MCP server...\n\n"))
	}

	if m.message != "" {
		b.WriteString(m.message + "\n")
	}

	if m.done {
		b.WriteString(successStyle.Render("✅ Installation complete! Press Ctrl+C to exit.\n"))
	}

	return b.String()
}

// Message types
type installResultMsg string

// Install command
func installMCP(agents []Agent) tea.Cmd {
	return func() tea.Msg {
		// Get the current executable path
		execPath, err := os.Executable()
		if err != nil {
			return installResultMsg(fmt.Sprintf("Error getting executable: %v", err))
		}

		// Install to selected agents
		var installed []string
		var failed []string

		for _, agent := range agents {
			if !agent.Selected {
				continue
			}

			configPath := filepath.Join(agent.ConfigDir, "mcp.json")
			if err := installToAgent(configPath, execPath); err != nil {
				failed = append(failed, fmt.Sprintf("%s: %v", agent.Name, err))
			} else {
				installed = append(installed, agent.Name)
			}
		}

		var result string
		if len(installed) > 0 {
			result += fmt.Sprintf("✅ Successfully installed to: %s\n\n", strings.Join(installed, ", "))
		}
		if len(failed) > 0 {
			result += fmt.Sprintf("❌ Failed:\n  %s", strings.Join(failed, "\n  "))
		} else {
			result += "\nRestart your AI agent to use the MCP server."
		}

		return installResultMsg(result)
	}
}

func installToAgent(configPath string, execPath string) error {
	// Create config directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	// Create MCP config struct
	mcpConfig := struct {
		MCPServers map[string]struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
		} `json:"mcpServers"`
	}{
		MCPServers: map[string]struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
		}{
			"bots": {
				Command: execPath,
				Args:    []string{"mcp", "serve"},
			},
		},
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(mcpConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			Align(lipgloss.Center)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205"))

	checkboxStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82"))

	agentStyle = lipgloss.NewStyle().
			Bold(true)

	descStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	installedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true)

	installingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))

	successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("82"))
)

// Run starts the installer TUI
func Run() {
	p := tea.NewProgram(InitialModel())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running installer: %v\n", err)
		os.Exit(1)
	}
}
