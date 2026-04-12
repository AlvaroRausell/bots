package install

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func keyPress(s string) tea.KeyMsg {
	switch s {
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	case " ":
		return tea.KeyMsg{Type: tea.KeySpace}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
}

func TestIsBotsMCPConfigured_NoFile(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "mcp.json")

	if isBotsMCPConfigured(configFile, FormatStandard) {
		t.Error("expected false when config file does not exist")
	}
}

func TestIsBotsMCPConfigured_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "mcp.json")
	os.WriteFile(configFile, []byte(""), 0644)

	if isBotsMCPConfigured(configFile, FormatStandard) {
		t.Error("expected false for empty config file")
	}
}

func TestIsBotsMCPConfigured_NoBotsEntry(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "mcp.json")

	config := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"other-server": map[string]interface{}{
				"command": "other",
				"args":    []string{"serve"},
			},
		},
	}
	data, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(configFile, data, 0644)

	if isBotsMCPConfigured(configFile, FormatStandard) {
		t.Error("expected false when bots entry does not exist")
	}
}

func TestIsBotsMCPConfigured_BotsEntryExists(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "mcp.json")

	config := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"bots": map[string]interface{}{
				"command": "/usr/local/bin/bots",
				"args":    []string{"mcp", "serve"},
			},
		},
	}
	data, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(configFile, data, 0644)

	if !isBotsMCPConfigured(configFile, FormatStandard) {
		t.Error("expected true when bots entry exists")
	}
}

func TestIsBotsMCPConfigured_OpenCodeNoBotsEntry(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "opencode.json")

	config := map[string]interface{}{
		"mcp": map[string]interface{}{
			"other-server": map[string]interface{}{
				"type":    "local",
				"command": []string{"other", "serve"},
			},
		},
	}
	data, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(configFile, data, 0644)

	if isBotsMCPConfigured(configFile, FormatOpenCode) {
		t.Error("expected false when bots entry does not exist in opencode format")
	}
}

func TestIsBotsMCPConfigured_OpenCodeBotsEntryExists(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "opencode.json")

	config := map[string]interface{}{
		"mcp": map[string]interface{}{
			"bots": map[string]interface{}{
				"type":    "local",
				"command": []string{"/usr/local/bin/bots", "mcp", "serve"},
			},
		},
	}
	data, _ := json.MarshalIndent(config, "", "  ")
	os.WriteFile(configFile, data, 0644)

	if !isBotsMCPConfigured(configFile, FormatOpenCode) {
		t.Error("expected true when bots entry exists in opencode format")
	}
}

func TestInstallToAgent_CreatesNewConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "mcp.json")

	err := installToAgent(configFile, "/usr/local/bin/bots", FormatStandard)
	if err != nil {
		t.Fatalf("installToAgent failed: %v", err)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	var config struct {
		MCPServers map[string]struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
		} `json:"mcpServers"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("failed to parse config JSON: %v", err)
	}

	botsEntry, ok := config.MCPServers["bots"]
	if !ok {
		t.Fatal("bots entry not found in config")
	}

	if botsEntry.Command != "/usr/local/bin/bots" {
		t.Errorf("expected command '/usr/local/bin/bots', got '%s'", botsEntry.Command)
	}

	if len(botsEntry.Args) != 2 || botsEntry.Args[0] != "mcp" || botsEntry.Args[1] != "serve" {
		t.Errorf("expected args ['mcp', 'serve'], got %v", botsEntry.Args)
	}
}

func TestInstallToAgent_MergesWithExistingConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "mcp.json")

	existingConfig := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"other-server": map[string]interface{}{
				"command": "other-bin",
				"args":    []string{"run"},
			},
		},
	}
	existingData, _ := json.MarshalIndent(existingConfig, "", "  ")
	os.WriteFile(configFile, existingData, 0644)

	err := installToAgent(configFile, "/usr/local/bin/bots", FormatStandard)
	if err != nil {
		t.Fatalf("installToAgent failed: %v", err)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	var config struct {
		MCPServers map[string]struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
		} `json:"mcpServers"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("failed to parse config JSON: %v", err)
	}

	if _, ok := config.MCPServers["other-server"]; !ok {
		t.Error("existing 'other-server' entry was lost after merge")
	}

	if _, ok := config.MCPServers["bots"]; !ok {
		t.Error("bots entry not found after merge")
	}
}

func TestInstallToAgent_OverwritesExistingBotsEntry(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "mcp.json")

	existingConfig := map[string]interface{}{
		"mcpServers": map[string]interface{}{
			"bots": map[string]interface{}{
				"command": "/old/path/bots",
				"args":    []string{"old", "args"},
			},
		},
	}
	existingData, _ := json.MarshalIndent(existingConfig, "", "  ")
	os.WriteFile(configFile, existingData, 0644)

	err := installToAgent(configFile, "/new/path/bots", FormatStandard)
	if err != nil {
		t.Fatalf("installToAgent failed: %v", err)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	var config struct {
		MCPServers map[string]struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
		} `json:"mcpServers"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("failed to parse config JSON: %v", err)
	}

	botsEntry := config.MCPServers["bots"]
	if botsEntry.Command != "/new/path/bots" {
		t.Errorf("expected command '/new/path/bots', got '%s'", botsEntry.Command)
	}

	if len(botsEntry.Args) != 2 || botsEntry.Args[0] != "mcp" || botsEntry.Args[1] != "serve" {
		t.Errorf("expected args ['mcp', 'serve'], got %v", botsEntry.Args)
	}
}

func TestInstallToAgent_CreatesConfigDir(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "subdir", "nested")
	configFile := filepath.Join(configDir, "mcp.json")

	err := installToAgent(configFile, "/usr/local/bin/bots", FormatStandard)
	if err != nil {
		t.Fatalf("installToAgent failed: %v", err)
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Error("config file was not created in nested directory")
	}
}

func TestInstallToAgent_InvalidJSONExisting(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "mcp.json")

	os.WriteFile(configFile, []byte("not valid json{{{"), 0644)

	err := installToAgent(configFile, "/usr/local/bin/bots", FormatStandard)
	if err != nil {
		t.Fatalf("installToAgent should handle invalid JSON gracefully, got: %v", err)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	var config struct {
		MCPServers map[string]struct {
			Command string   `json:"command"`
			Args    []string `json:"args"`
		} `json:"mcpServers"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("result should be valid JSON, got parse error: %v", err)
	}

	if _, ok := config.MCPServers["bots"]; !ok {
		t.Error("bots entry not found after recovering from invalid JSON")
	}
}

func TestInitialModel_HasAllAgents(t *testing.T) {
	m := InitialModel()

	if len(m.agents) != 3 {
		t.Fatalf("expected 3 agents, got %d", len(m.agents))
	}

	names := map[string]bool{}
	for _, a := range m.agents {
		names[a.Name] = true
		if a.ConfigFile == "" {
			t.Errorf("agent %s has empty ConfigFile", a.Name)
		}
		if a.ConfigDir == "" {
			t.Errorf("agent %s has empty ConfigDir", a.Name)
		}
	}

	for _, name := range []string{"opencode", "claude", "codex"} {
		if !names[name] {
			t.Errorf("missing agent: %s", name)
		}
	}
}

func TestInitialModel_DefaultSelection(t *testing.T) {
	m := InitialModel()

	if !m.agents[0].Selected {
		t.Error("opencode should be selected by default")
	}
	if m.agents[1].Selected {
		t.Error("claude should not be selected by default")
	}
	if m.agents[2].Selected {
		t.Error("codex should not be selected by default")
	}
}

func TestUpdate_NavigateUp(t *testing.T) {
	m := InitialModel()
	m.cursor = 1

	model, _ := m.Update(keyPress("up"))
	updated := model.(Model)
	if updated.cursor != 0 {
		t.Errorf("expected cursor 0, got %d", updated.cursor)
	}
}

func TestUpdate_NavigateUpAtTop(t *testing.T) {
	m := InitialModel()
	m.cursor = 0

	model, _ := m.Update(keyPress("up"))
	updated := model.(Model)
	if updated.cursor != 0 {
		t.Errorf("cursor should not go below 0, got %d", updated.cursor)
	}
}

func TestUpdate_NavigateDown(t *testing.T) {
	m := InitialModel()

	model, _ := m.Update(keyPress("down"))
	updated := model.(Model)
	if updated.cursor != 1 {
		t.Errorf("expected cursor 1, got %d", updated.cursor)
	}
}

func TestUpdate_NavigateDownAtBottom(t *testing.T) {
	m := InitialModel()
	m.cursor = len(m.agents) - 1

	model, _ := m.Update(keyPress("down"))
	updated := model.(Model)
	if updated.cursor != len(m.agents)-1 {
		t.Errorf("cursor should not go past last item, got %d", updated.cursor)
	}
}

func TestUpdate_ToggleSelection(t *testing.T) {
	m := InitialModel()

	if !m.agents[0].Selected {
		t.Error("opencode should start selected")
	}

	model, _ := m.Update(keyPress(" "))
	updated := model.(Model)
	if updated.agents[0].Selected {
		t.Error("opencode should be unselected after toggle")
	}

	model, _ = updated.Update(keyPress(" "))
	updated = model.(Model)
	if !updated.agents[0].Selected {
		t.Error("opencode should be selected after second toggle")
	}
}

func TestUpdate_ToggleWithEnter(t *testing.T) {
	m := InitialModel()

	model, _ := m.Update(keyPress("enter"))
	updated := model.(Model)
	if updated.agents[0].Selected {
		t.Error("opencode should be unselected after enter toggle")
	}
}

func TestUpdate_JKeyNavigateDown(t *testing.T) {
	m := InitialModel()

	model, _ := m.Update(keyPress("j"))
	updated := model.(Model)
	if updated.cursor != 1 {
		t.Errorf("j should move cursor down, got %d", updated.cursor)
	}
}

func TestUpdate_KKeyNavigateUp(t *testing.T) {
	m := InitialModel()
	m.cursor = 1

	model, _ := m.Update(keyPress("k"))
	updated := model.(Model)
	if updated.cursor != 0 {
		t.Errorf("k should move cursor up, got %d", updated.cursor)
	}
}

func TestUpdate_QuitWithCtrlC(t *testing.T) {
	m := InitialModel()

	_, cmd := m.Update(keyPress("ctrl+c"))
	if cmd == nil {
		t.Error("ctrl+c should return quit command")
	}
}

func TestUpdate_QuitWithQ(t *testing.T) {
	m := InitialModel()

	_, cmd := m.Update(keyPress("q"))
	if cmd == nil {
		t.Error("q should return quit command")
	}
}

func TestUpdate_InstallWithNoSelection(t *testing.T) {
	m := InitialModel()
	for i := range m.agents {
		m.agents[i].Selected = false
	}

	model, _ := m.Update(keyPress("i"))
	updated := model.(Model)
	if updated.installing {
		t.Error("should not start installing when no agents selected")
	}
}

func TestUpdate_DoneStateQuitWithQ(t *testing.T) {
	m := InitialModel()
	m.done = true

	_, cmd := m.Update(keyPress("q"))
	if cmd == nil {
		t.Error("q should quit in done state")
	}
}

func TestUpdate_DoneStateCtrlC(t *testing.T) {
	m := InitialModel()
	m.done = true

	_, cmd := m.Update(keyPress("ctrl+c"))
	if cmd == nil {
		t.Error("ctrl+c should quit in done state")
	}
}

func TestUpdate_InstallStartsWithSelection(t *testing.T) {
	m := InitialModel()
	m.agents[0].Selected = true

	model, cmd := m.Update(keyPress("i"))
	updated := model.(Model)
	if !updated.installing {
		t.Error("should start installing when agents selected")
	}
	if cmd == nil {
		t.Error("should return install command")
	}
}

func TestView_ShowsConfigPaths(t *testing.T) {
	m := InitialModel()
	view := m.View()

	for _, agent := range m.agents {
		if !strings.Contains(view, agent.ConfigFile) {
			t.Errorf("view should show config path %s", agent.ConfigFile)
		}
	}
}

func TestView_ShowsInstalling(t *testing.T) {
	m := InitialModel()
	m.installing = true
	view := m.View()

	if !strings.Contains(view, "Installing") {
		t.Error("view should show installing message")
	}
}

func TestView_ShowsDone(t *testing.T) {
	m := InitialModel()
	m.done = true
	m.message = "test result"
	view := m.View()

	if !strings.Contains(view, "Installation complete") {
		t.Error("view should show completion message")
	}
}

func TestView_StatusLabels(t *testing.T) {
	m := InitialModel()
	view := m.View()

	hasStatus := false
	for _, label := range []string{"configured", "available", "not found"} {
		if strings.Contains(view, label) {
			hasStatus = true
			break
		}
	}
	if !hasStatus {
		t.Error("view should show one of the status labels")
	}
}

func TestView_ShowsHelp(t *testing.T) {
	m := InitialModel()
	view := m.View()

	if !strings.Contains(view, "Navigate") {
		t.Error("view should show navigation help")
	}
	if !strings.Contains(view, "Toggle") {
		t.Error("view should show toggle help")
	}
	if !strings.Contains(view, "Install") {
		t.Error("view should show install help")
	}
}

func TestCommandExists_InvalidCommand(t *testing.T) {
	if commandExists("definitely-not-a-real-command-xyz-123") {
		t.Error("expected false for non-existent command")
	}
}

func TestCommandExists_ValidCommand(t *testing.T) {
	if !commandExists("ls") {
		t.Error("expected true for 'ls' command")
	}
}

func TestInstallToAgent_OpenCodeCreatesNewConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "opencode.json")

	err := installToAgent(configFile, "/usr/local/bin/bots", FormatOpenCode)
	if err != nil {
		t.Fatalf("installToAgent failed: %v", err)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	var config struct {
		MCP map[string]struct {
			Type    string   `json:"type"`
			Command []string `json:"command"`
		} `json:"mcp"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("failed to parse config JSON: %v", err)
	}

	botsEntry, ok := config.MCP["bots"]
	if !ok {
		t.Fatal("bots entry not found in opencode config")
	}

	if botsEntry.Type != "local" {
		t.Errorf("expected type 'local', got '%s'", botsEntry.Type)
	}

	if len(botsEntry.Command) != 3 || botsEntry.Command[0] != "/usr/local/bin/bots" || botsEntry.Command[1] != "mcp" || botsEntry.Command[2] != "serve" {
		t.Errorf("expected command ['/usr/local/bin/bots', 'mcp', 'serve'], got %v", botsEntry.Command)
	}
}

func TestInstallToAgent_OpenCodeMergesWithExistingConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "opencode.json")

	existingConfig := map[string]interface{}{
		"$schema": "https://opencode.ai/config.json",
		"mcp": map[string]interface{}{
			"existing-server": map[string]interface{}{
				"type":    "remote",
				"url":     "https://example.com/mcp",
				"enabled": true,
			},
		},
	}
	existingData, _ := json.MarshalIndent(existingConfig, "", "  ")
	os.WriteFile(configFile, existingData, 0644)

	err := installToAgent(configFile, "/usr/local/bin/bots", FormatOpenCode)
	if err != nil {
		t.Fatalf("installToAgent failed: %v", err)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	var config map[string]json.RawMessage
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("failed to parse config JSON: %v", err)
	}

	if _, ok := config["$schema"]; !ok {
		t.Error("existing top-level keys were lost after merge")
	}

	var mcp map[string]interface{}
	if err := json.Unmarshal(config["mcp"], &mcp); err != nil {
		t.Fatalf("failed to parse mcp JSON: %v", err)
	}

	if _, ok := mcp["existing-server"]; !ok {
		t.Error("existing 'existing-server' entry was lost after merge")
	}
	if _, ok := mcp["bots"]; !ok {
		t.Error("bots entry not found after merge")
	}
}

func TestInstallToAgent_OpenCodePreservesRemoteEntries(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "opencode.json")

	existingConfig := map[string]interface{}{
		"$schema": "https://opencode.ai/config.json",
		"mcp": map[string]interface{}{
			"context7": map[string]interface{}{
				"type":    "remote",
				"url":     "https://mcp.context7.com/mcp",
				"enabled": true,
				"headers": map[string]interface{}{
					"CONTEXT7_API_KEY": "test-key",
				},
			},
		},
	}
	existingData, _ := json.MarshalIndent(existingConfig, "", "  ")
	os.WriteFile(configFile, existingData, 0644)

	err := installToAgent(configFile, "/usr/local/bin/bots", FormatOpenCode)
	if err != nil {
		t.Fatalf("installToAgent failed: %v", err)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	var config struct {
		MCP map[string]json.RawMessage `json:"mcp"`
	}
	if err := json.Unmarshal(data, &config); err != nil {
		t.Fatalf("failed to parse config JSON: %v", err)
	}

	var context7 map[string]interface{}
	if err := json.Unmarshal(config.MCP["context7"], &context7); err != nil {
		t.Fatalf("failed to parse context7 entry: %v", err)
	}

	if context7["type"] != "remote" {
		t.Errorf("expected context7 type 'remote', got '%v'", context7["type"])
	}
	if context7["url"] != "https://mcp.context7.com/mcp" {
		t.Errorf("expected context7 url preserved, got '%v'", context7["url"])
	}
	if _, ok := context7["headers"]; !ok {
		t.Error("context7 headers were lost after merge")
	}
}
