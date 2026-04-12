package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Serve starts the MCP server on stdio
func Serve() {
	reader := os.Stdin
	decoder := json.NewDecoder(reader)
	encoder := json.NewEncoder(os.Stdout)

	// Send initialization response
	initResponse := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"result": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{
					"listChanged": true,
				},
			},
			"serverInfo": map[string]interface{}{
				"name":    "bots",
				"version": "0.1.0",
			},
		},
	}
	encoder.Encode(initResponse)

	// Process messages
	for {
		var message map[string]interface{}
		if err := decoder.Decode(&message); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			continue
		}

		method, ok := message["method"].(string)
		if !ok {
			sendError(encoder, -32600, "Invalid Request", message["id"])
			continue
		}

		id := message["id"]

		switch method {
		case "tools/list":
			handleToolsList(encoder, id)
		case "tools/call":
			params, ok := message["params"].(map[string]interface{})
			if !ok {
				sendError(encoder, -32602, "Invalid params", id)
				continue
			}
			handleToolsCall(encoder, id, params)
		case "initialize":
			// Already sent init response, ignore
			continue
		case "notifications/initialized":
			// Ignore notification
			continue
		default:
			sendError(encoder, -32601, "Method not found", id)
		}
	}
}

func handleToolsList(encoder *json.Encoder, id interface{}) {
	tools := []map[string]interface{}{
		// Checkpoint tools
		{
			"name":        "checkpoint_read",
			"description": "Read the current CHECKPOINTS.md content",
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			"name":        "checkpoint_update",
			"description": "Update a specific section in CHECKPOINTS.md",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"section": map[string]interface{}{
						"type":        "string",
						"description": "Section name to update",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "New content for the section",
					},
				},
				"required": []string{"section", "content"},
			},
		},
		{
			"name":        "checkpoint_list",
			"description": "List all checkpoint sections",
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		// Log tools
		{
			"name":        "log_create",
			"description": "Create a new session log file",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"topic": map[string]interface{}{
						"type":        "string",
						"description": "Topic/title for the session log",
					},
				},
				"required": []string{"topic"},
			},
		},
		{
			"name":        "log_append",
			"description": "Append an entry to a session log",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"slug": map[string]interface{}{
						"type":        "string",
						"description": "Log file slug or partial name",
					},
					"message": map[string]interface{}{
						"type":        "string",
						"description": "Entry content to append",
					},
				},
				"required": []string{"slug", "message"},
			},
		},
		{
			"name":        "log_search",
			"description": "Search across all session logs",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			"name":        "log_summarize",
			"description": "Generate a summary of decisions from a log",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"slug": map[string]interface{}{
						"type":        "string",
						"description": "Log file slug or partial name",
					},
				},
				"required": []string{"slug"},
			},
		},
		{
			"name":        "log_list",
			"description": "List all session log files",
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		// Task tools
		{
			"name":        "task_create",
			"description": "Create a new task file",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"slug": map[string]interface{}{
						"type":        "string",
						"description": "Task slug (e.g., 'phase-1.5')",
					},
				},
				"required": []string{"slug"},
			},
		},
		{
			"name":        "task_read",
			"description": "Read a task file",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"slug": map[string]interface{}{
						"type":        "string",
						"description": "Task slug",
					},
				},
				"required": []string{"slug"},
			},
		},
		{
			"name":        "task_update_status",
			"description": "Update task status",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"slug": map[string]interface{}{
						"type":        "string",
						"description": "Task slug",
					},
					"status": map[string]interface{}{
						"type":        "string",
						"description": "New status (PENDING, IN_PROGRESS, READY_FOR_REVIEW, CHANGES_REQUESTED, DONE)",
						"enum":        []string{"PENDING", "IN_PROGRESS", "READY_FOR_REVIEW", "CHANGES_REQUESTED", "DONE"},
					},
				},
				"required": []string{"slug", "status"},
			},
		},
		// Git tools
		{
			"name":        "git_commit_checkpoint",
			"description": "Commit checkpoint changes to git",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"message": map[string]interface{}{
						"type":        "string",
						"description": "Commit message",
					},
				},
				"required": []string{"message"},
			},
		},
		{
			"name":        "git_branch_info",
			"description": "Get current git branch information",
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}

	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result": map[string]interface{}{
			"tools": tools,
		},
	}
	encoder.Encode(response)
}

func handleToolsCall(encoder *json.Encoder, id interface{}, params map[string]interface{}) {
	name, ok := params["name"].(string)
	if !ok {
		sendError(encoder, -32602, "Invalid tool name", id)
		return
	}

	arguments, _ := params["arguments"].(map[string]interface{})

	switch name {
	case "checkpoint_read":
		result := executeCheckpointRead()
		sendSuccess(encoder, id, result)
	case "checkpoint_update":
		section, _ := arguments["section"].(string)
		content, _ := arguments["content"].(string)
		result := executeCheckpointUpdate(section, content)
		sendSuccess(encoder, id, result)
	case "checkpoint_list":
		result := executeCheckpointList()
		sendSuccess(encoder, id, result)
	case "log_create":
		topic, _ := arguments["topic"].(string)
		result := executeLogCreate(topic)
		sendSuccess(encoder, id, result)
	case "log_append":
		slug, _ := arguments["slug"].(string)
		message, _ := arguments["message"].(string)
		result := executeLogAppend(slug, message)
		sendSuccess(encoder, id, result)
	case "log_search":
		query, _ := arguments["query"].(string)
		result := executeLogSearch(query)
		sendSuccess(encoder, id, result)
	case "log_summarize":
		slug, _ := arguments["slug"].(string)
		result := executeLogSummarize(slug)
		sendSuccess(encoder, id, result)
	case "log_list":
		result := executeLogList()
		sendSuccess(encoder, id, result)
	case "task_create":
		slug, _ := arguments["slug"].(string)
		result := executeTaskCreate(slug)
		sendSuccess(encoder, id, result)
	case "task_read":
		slug, _ := arguments["slug"].(string)
		result := executeTaskRead(slug)
		sendSuccess(encoder, id, result)
	case "task_update_status":
		slug, _ := arguments["slug"].(string)
		status, _ := arguments["status"].(string)
		result := executeTaskUpdateStatus(slug, status)
		sendSuccess(encoder, id, result)
	case "git_commit_checkpoint":
		message, _ := arguments["message"].(string)
		result := executeGitCommitCheckpoint(message)
		sendSuccess(encoder, id, result)
	case "git_branch_info":
		result := executeGitBranchInfo()
		sendSuccess(encoder, id, result)
	default:
		sendError(encoder, -32602, "Unknown tool: "+name, id)
	}
}

// Tool execution wrappers that call the CLI

func runCLICommand(args ...string) (string, error) {
	// Get the path to the bots binary
	execPath, err := os.Executable()
	if err != nil {
		// Fallback to "bots" in PATH
		execPath = "bots"
	}

	cmd := exec.Command(execPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return stderr.String(), err
	}

	return stdout.String(), nil
}

func executeCheckpointRead() string {
	output, err := runCLICommand("checkpoint", "read")
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return strings.TrimSpace(output)
}

func executeCheckpointUpdate(section, content string) string {
	output, err := runCLICommand("checkpoint", "update", section, content)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return strings.TrimSpace(output)
}

func executeCheckpointList() string {
	output, err := runCLICommand("checkpoint", "list")
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return strings.TrimSpace(output)
}

func executeLogCreate(topic string) string {
	output, err := runCLICommand("log", "start", topic)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return strings.TrimSpace(output)
}

func executeLogAppend(slug, message string) string {
	output, err := runCLICommand("log", "append", slug, message)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return strings.TrimSpace(output)
}

func executeLogSearch(query string) string {
	output, err := runCLICommand("log", "search", query)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return strings.TrimSpace(output)
}

func executeLogSummarize(slug string) string {
	output, err := runCLICommand("log", "summarize", slug)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return strings.TrimSpace(output)
}

func executeLogList() string {
	output, err := runCLICommand("log", "list")
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return strings.TrimSpace(output)
}

func executeTaskCreate(slug string) string {
	output, err := runCLICommand("task", "create", slug)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return strings.TrimSpace(output)
}

func executeTaskRead(slug string) string {
	output, err := runCLICommand("task", "read", slug)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return strings.TrimSpace(output)
}

func executeTaskUpdateStatus(slug, status string) string {
	output, err := runCLICommand("task", "update", slug, status)
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}
	return strings.TrimSpace(output)
}

func executeGitCommitCheckpoint(message string) string {
	// Check if we're in a git repo
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		return "Not in a git repository"
	}

	// Stage .bots directory
	if err := exec.Command("git", "add", ".bots").Run(); err != nil {
		return fmt.Sprintf("Error staging: %v", err)
	}

	// Commit
	if err := exec.Command("git", "commit", "-m", message).Run(); err != nil {
		return fmt.Sprintf("Error committing: %v", err)
	}

	return "Committed checkpoint changes successfully"
}

func executeGitBranchInfo() string {
	// Get current branch
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "Not in a git repository"
	}
	branch := strings.TrimSpace(string(output))

	// Get current commit
	cmd = exec.Command("git", "rev-parse", "--short", "HEAD")
	output, err = cmd.Output()
	if err != nil {
		return branch
	}
	commit := strings.TrimSpace(string(output))

	return fmt.Sprintf("Branch: %s, Commit: %s", branch, commit)
}

func sendSuccess(encoder *json.Encoder, id interface{}, content string) {
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result": map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": content,
				},
			},
		},
	}
	encoder.Encode(response)
}

func sendError(encoder *json.Encoder, code int, message string, id interface{}) {
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}
	if id != nil {
		response["id"] = id
	}
	encoder.Encode(response)
}
