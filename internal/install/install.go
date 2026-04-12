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

type Agent struct {
	Name          string
	Description   string
	ConfigDir     string
	ConfigFile    string
	AgentPresent  bool
	MCPConfigured bool
	Selected      bool
}

type Model struct {
	agents     []Agent
	cursor     int
	installing bool
	done       bool
	message    string
	width      int
}

func getAgents() []Agent {
	homeDir, _ := os.UserHomeDir()

	agents := []Agent{
		{
			Name:         "opencode",
			Description:  "OpenCode AI agent",
			ConfigDir:    filepath.Join(homeDir, ".opencode"),
			ConfigFile:   filepath.Join(homeDir, ".opencode", "mcp.json"),
			AgentPresent: false,
			Selected:     true,
		},
		{
			Name:         "claude",
			Description:  "Claude Code (Anthropic)",
			ConfigDir:    filepath.Join(homeDir, ".claude"),
			ConfigFile:   filepath.Join(homeDir, ".claude", "mcp.json"),
			AgentPresent: false,
			Selected:     false,
		},
		{
			Name:         "codex",
			Description:  "OpenAI Codex",
			ConfigDir:    filepath.Join(homeDir, ".codex"),
			ConfigFile:   filepath.Join(homeDir, ".codex", "mcp.json"),
			AgentPresent: false,
			Selected:     false,
		},
	}

	for i := range agents {
		if _, err := os.Stat(agents[i].ConfigDir); err == nil {
			agents[i].AgentPresent = true
		}
		agents[i].AgentPresent = agents[i].AgentPresent || commandExists(agents[i].Name)

		agents[i].MCPConfigured = isBotsMCPConfigured(agents[i].ConfigFile)
	}

	return agents
}

func isBotsMCPConfigured(configFile string) bool {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return false
	}

	var config struct {
		MCPServers map[string]json.RawMessage `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return false
	}

	_, exists := config.MCPServers["bots"]
	return exists
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func InitialModel() Model {
	return Model{
		agents:     getAgents(),
		cursor:     0,
		installing: false,
		done:       false,
		message:    "",
		width:      60,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.done {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			}
		}
		return m, nil
	}

	if m.installing {
		if _, ok := msg.(installResultMsg); ok {
			m.done = true
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
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
			anySelected := false
			for _, a := range m.agents {
				if a.Selected {
					anySelected = true
					break
				}
			}
			if !anySelected {
				return m, nil
			}
			m.installing = true
			return m, installMCP(m.agents)
		}
	}

	return m, nil
}

func (m Model) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Bots MCP Installer"))
	b.WriteString("\n\n")
	b.WriteString(subtitleStyle.Render("Select agents to configure with the bots MCP server"))
	b.WriteString("\n\n")

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
		if agent.MCPConfigured {
			status = configuredStyle.Render(" ✓ configured")
		} else if agent.AgentPresent {
			status = agentPresentStyle.Render(" ● available")
		} else {
			status = notFoundStyle.Render(" ✗ not found")
		}

		line := fmt.Sprintf("%s%s %s%s\n", cursor, checkbox, agentStyle.Render(agent.Name), status)
		b.WriteString(line)

		pathLine := fmt.Sprintf("    %s\n", pathStyle.Render(agent.ConfigFile))
		b.WriteString(pathLine)
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("↑/k ↓/j Navigate • Space/Enter Toggle • i Install • q/Ctrl+C Quit"))
	b.WriteString("\n\n")

	if m.installing && !m.done {
		b.WriteString(installingStyle.Render("⏳ Installing MCP server..."))
		b.WriteString("\n\n")
	}

	if m.message != "" {
		b.WriteString(m.message)
		b.WriteString("\n")
	}

	if m.done {
		b.WriteString(successStyle.Render("Installation complete! Press q or Ctrl+C to exit."))
		b.WriteString("\n")
	}

	return b.String()
}

type installResultMsg string

func installMCP(agents []Agent) tea.Cmd {
	return func() tea.Msg {
		execPath, err := os.Executable()
		if err != nil {
			return installResultMsg(fmt.Sprintf("Error getting executable: %v", err))
		}

		var installed []string
		var failed []string
		var skipped []string

		for _, agent := range agents {
			if !agent.Selected {
				continue
			}

			if agent.MCPConfigured {
				skipped = append(skipped, agent.Name)
				continue
			}

			if err := installToAgent(agent.ConfigFile, execPath); err != nil {
				failed = append(failed, fmt.Sprintf("%s: %v", agent.Name, err))
			} else {
				installed = append(installed, agent.Name)
			}
		}

		var result string
		if len(installed) > 0 {
			result += successStyle.Render(fmt.Sprintf("Installed to: %s", strings.Join(installed, ", ")))
			result += "\n"
		}
		if len(skipped) > 0 {
			result += configuredStyle.Render(fmt.Sprintf("Already configured: %s", strings.Join(skipped, ", ")))
			result += "\n"
		}
		if len(failed) > 0 {
			result += failStyle.Render(fmt.Sprintf("Failed:\n  %s", strings.Join(failed, "\n  ")))
			result += "\n"
		}
		if len(installed) > 0 {
			result += "\nRestart your AI agent to use the MCP server."
		}

		return installResultMsg(result)
	}
}

func installToAgent(configPath string, execPath string) error {
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	existing := make(map[string]struct {
		Command string   `json:"command"`
		Args    []string `json:"args"`
	})

	data, err := os.ReadFile(configPath)
	if err == nil {
		var parsed struct {
			MCPServers map[string]struct {
				Command string   `json:"command"`
				Args    []string `json:"args"`
			} `json:"mcpServers"`
		}
		if json.Unmarshal(data, &parsed) == nil && parsed.MCPServers != nil {
			existing = parsed.MCPServers
		}
	}

	existing["bots"] = struct {
		Command string   `json:"command"`
		Args    []string `json:"args"`
	}{
		Command: execPath,
		Args:    []string{"mcp", "serve"},
	}

	mcpConfig := struct {
		MCPServers map[string]struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
		} `json:"mcpServers"`
	}{
		MCPServers: existing,
	}

	jsonData, err := json.MarshalIndent(mcpConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205"))

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	cursorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205"))

	checkboxStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82"))

	agentStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("252"))

	pathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("243"))

	configuredStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("82"))

	agentPresentStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214"))

	notFoundStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("203"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	installingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))

	successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("82"))

	failStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("203"))
)

func Run() {
	p := tea.NewProgram(InitialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running installer: %v\n", err)
		os.Exit(1)
	}
}
