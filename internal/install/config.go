package install

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// AgentConfigDocument is an AI agent configuration file that can contain a Bots MCP server entry.
type AgentConfigDocument struct {
	Path   string
	Format AgentFormat
}

func (d AgentConfigDocument) HasBotsMCP() (bool, error) {
	data, err := os.ReadFile(d.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("read config: %w", err)
	}

	if d.Format == FormatOpenCode {
		var config struct {
			MCP map[string]json.RawMessage `json:"mcp"`
		}
		if err := json.Unmarshal(data, &config); err != nil {
			return false, nil
		}
		_, exists := config.MCP["bots"]
		return exists, nil
	}

	var config struct {
		MCPServers map[string]json.RawMessage `json:"mcpServers"`
	}
	if err := json.Unmarshal(data, &config); err != nil {
		return false, nil
	}

	_, exists := config.MCPServers["bots"]
	return exists, nil
}

func (d AgentConfigDocument) InstallBotsMCP(execPath string) error {
	configDir := filepath.Dir(d.Path)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	if d.Format == FormatOpenCode {
		return d.installOpenCode(execPath)
	}
	return d.installStandard(execPath)
}

func (d AgentConfigDocument) backup() error {
	data, err := os.ReadFile(d.Path)
	if err != nil {
		return nil
	}

	ts := time.Now().UnixNano()
	backupPath := fmt.Sprintf("%s.bak.%d", d.Path, ts)
	return os.WriteFile(backupPath, data, 0644)
}

func (d AgentConfigDocument) installStandard(execPath string) error {
	existing := make(map[string]standardMCPEntry)

	data, err := os.ReadFile(d.Path)
	if err == nil {
		if err := d.backup(); err != nil {
			return fmt.Errorf("failed to backup config: %w", err)
		}
		var parsed struct {
			MCPServers map[string]standardMCPEntry `json:"mcpServers"`
		}
		if json.Unmarshal(data, &parsed) == nil && parsed.MCPServers != nil {
			existing = parsed.MCPServers
		}
	}

	existing["bots"] = standardMCPEntry{
		Command: execPath,
		Args:    []string{"mcp", "serve"},
	}

	mcpConfig := struct {
		MCPServers map[string]standardMCPEntry `json:"mcpServers"`
	}{
		MCPServers: existing,
	}

	jsonData, err := json.MarshalIndent(mcpConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(d.Path, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

func (d AgentConfigDocument) installOpenCode(execPath string) error {
	existing := make(map[string]json.RawMessage)

	data, err := os.ReadFile(d.Path)
	if err == nil {
		if backupErr := d.backup(); backupErr != nil {
			return fmt.Errorf("failed to backup config: %w", backupErr)
		}
		var parsed map[string]json.RawMessage
		if json.Unmarshal(data, &parsed) == nil {
			existing = parsed
		}
	}

	mcpServers := make(map[string]json.RawMessage)

	if rawMcp, ok := existing["mcp"]; ok {
		var parsed map[string]json.RawMessage
		if json.Unmarshal(rawMcp, &parsed) == nil {
			mcpServers = parsed
		}
	}

	botsEntry := openCodeMCPEntry{
		Type:    "local",
		Command: []string{execPath, "mcp", "serve"},
	}
	botsJSON, err := json.Marshal(botsEntry)
	if err != nil {
		return fmt.Errorf("failed to marshal bots entry: %w", err)
	}
	mcpServers["bots"] = json.RawMessage(botsJSON)

	mcpJSON, err := json.MarshalIndent(mcpServers, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal mcp config: %w", err)
	}

	config := make(map[string]json.RawMessage)
	for k, v := range existing {
		if k != "mcp" {
			config[k] = v
		}
	}
	config["mcp"] = json.RawMessage(mcpJSON)

	jsonData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(d.Path, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

type standardMCPEntry struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

type openCodeMCPEntry struct {
	Type    string   `json:"type"`
	Command []string `json:"command"`
}

func isBotsMCPConfigured(configFile string, format AgentFormat) bool {
	configured, err := AgentConfigDocument{Path: configFile, Format: format}.HasBotsMCP()
	return err == nil && configured
}

func installToAgent(configPath string, execPath string, format AgentFormat) error {
	return AgentConfigDocument{Path: configPath, Format: format}.InstallBotsMCP(execPath)
}
