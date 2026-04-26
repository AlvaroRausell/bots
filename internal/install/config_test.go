package install

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestAgentConfigDocumentInstallsStandardBotsEntry(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mcp.json")
	doc := AgentConfigDocument{Path: path, Format: FormatStandard}

	configured, err := doc.HasBotsMCP()
	if err != nil {
		t.Fatalf("HasBotsMCP returned error: %v", err)
	}
	if configured {
		t.Fatal("expected missing config to be unconfigured")
	}

	if err := doc.InstallBotsMCP("/usr/local/bin/bots"); err != nil {
		t.Fatalf("InstallBotsMCP returned error: %v", err)
	}

	configured, err = doc.HasBotsMCP()
	if err != nil {
		t.Fatalf("HasBotsMCP after install returned error: %v", err)
	}
	if !configured {
		t.Fatal("expected bots MCP to be configured")
	}
}

func TestAgentConfigDocumentInstallsOpenCodeBotsEntryWithoutLosingExistingConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "opencode.json")
	existing := map[string]interface{}{
		"$schema": "https://opencode.ai/config.json",
		"mcp": map[string]interface{}{
			"remote": map[string]interface{}{
				"type":    "remote",
				"url":     "https://example.com/mcp",
				"enabled": true,
			},
		},
	}
	data, _ := json.MarshalIndent(existing, "", "  ")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	doc := AgentConfigDocument{Path: path, Format: FormatOpenCode}
	if err := doc.InstallBotsMCP("/usr/local/bin/bots"); err != nil {
		t.Fatalf("InstallBotsMCP returned error: %v", err)
	}

	var config map[string]json.RawMessage
	installed, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	if err := json.Unmarshal(installed, &config); err != nil {
		t.Fatalf("parse config: %v", err)
	}
	if _, ok := config["$schema"]; !ok {
		t.Fatal("expected top-level schema key to be preserved")
	}
	var mcp map[string]json.RawMessage
	if err := json.Unmarshal(config["mcp"], &mcp); err != nil {
		t.Fatalf("parse mcp: %v", err)
	}
	if _, ok := mcp["remote"]; !ok {
		t.Fatal("expected existing remote MCP entry to be preserved")
	}
	if _, ok := mcp["bots"]; !ok {
		t.Fatal("expected bots MCP entry")
	}
}
