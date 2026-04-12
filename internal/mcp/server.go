package mcp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"bots/internal/checkpoint"
	"bots/internal/log"
	"bots/internal/task"
)

// Serve starts the MCP server on stdio
func Serve() {
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for {
		var message map[string]interface{}
		if err := decoder.Decode(&message); err != nil {
			if err == io.EOF {
				return
			}
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			continue
		}

		method, ok := message["method"].(string)
		if !ok {
			sendError(encoder, -32600, "Invalid Request", nil)
			continue
		}

		id := message["id"]

		switch method {
		case "initialize":
			sendInitializeResponse(encoder, id)
		case "notifications/initialized":
			continue
		case "tools/list":
			handleToolsList(encoder, id)
		case "tools/call":
			params, ok := message["params"].(map[string]interface{})
			if !ok {
				sendError(encoder, -32602, "Invalid params", id)
				continue
			}
			handleToolsCall(encoder, id, params)
		default:
			sendError(encoder, -32601, "Method not found", id)
		}
	}
}

func sendInitializeResponse(encoder *json.Encoder, id interface{}) {
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result": map[string]interface{}{
			"protocolVersion": "2025-11-25",
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
	encoder.Encode(response)
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

	arguments, ok := params["arguments"].(map[string]interface{})
	if !ok {
		arguments = make(map[string]interface{})
	}

	missingArg := func(argName string) {
		sendToolError(encoder, id, fmt.Sprintf("Missing required argument: %s", argName))
	}

	switch name {
	case "checkpoint_read":
		result := executeCheckpointRead()
		sendSuccess(encoder, id, result)
	case "checkpoint_update":
		section, _ := arguments["section"].(string)
		content, _ := arguments["content"].(string)
		if section == "" {
			missingArg("section")
			return
		}
		result := executeCheckpointUpdate(section, content)
		sendSuccess(encoder, id, result)
	case "checkpoint_list":
		result := executeCheckpointList()
		sendSuccess(encoder, id, result)
	case "log_create":
		topic, _ := arguments["topic"].(string)
		if topic == "" {
			missingArg("topic")
			return
		}
		result := executeLogCreate(topic)
		sendSuccess(encoder, id, result)
	case "log_append":
		slug, _ := arguments["slug"].(string)
		message, _ := arguments["message"].(string)
		if slug == "" {
			missingArg("slug")
			return
		}
		if message == "" {
			missingArg("message")
			return
		}
		result := executeLogAppend(slug, message)
		sendSuccess(encoder, id, result)
	case "log_search":
		query, _ := arguments["query"].(string)
		if query == "" {
			missingArg("query")
			return
		}
		result := executeLogSearch(query)
		sendSuccess(encoder, id, result)
	case "log_summarize":
		slug, _ := arguments["slug"].(string)
		if slug == "" {
			missingArg("slug")
			return
		}
		result := executeLogSummarize(slug)
		sendSuccess(encoder, id, result)
	case "log_list":
		result := executeLogList()
		sendSuccess(encoder, id, result)
	case "task_create":
		slug, _ := arguments["slug"].(string)
		if slug == "" {
			missingArg("slug")
			return
		}
		result := executeTaskCreate(slug)
		sendSuccess(encoder, id, result)
	case "task_read":
		slug, _ := arguments["slug"].(string)
		if slug == "" {
			missingArg("slug")
			return
		}
		result := executeTaskRead(slug)
		sendSuccess(encoder, id, result)
	case "task_update_status":
		slug, _ := arguments["slug"].(string)
		status, _ := arguments["status"].(string)
		if slug == "" {
			missingArg("slug")
			return
		}
		if status == "" {
			missingArg("status")
			return
		}
		result := executeTaskUpdateStatus(slug, status)
		sendSuccess(encoder, id, result)
	case "git_commit_checkpoint":
		message, _ := arguments["message"].(string)
		if message == "" {
			missingArg("message")
			return
		}
		result := executeGitCommitCheckpoint(message)
		sendSuccess(encoder, id, result)
	case "git_branch_info":
		result := executeGitBranchInfo()
		sendSuccess(encoder, id, result)
	default:
		sendToolError(encoder, id, "Unknown tool: "+name)
	}
}

// Tool execution wrappers that call internal packages directly.
// This avoids argument-splitting issues that would occur when passing
// multi-word content through CLI subprocess args.

func executeCheckpointRead() string {
	var buf strings.Builder
	checkpoint.ReadTo(&buf)
	return strings.TrimSpace(buf.String())
}

func executeCheckpointUpdate(section, content string) string {
	var buf strings.Builder
	checkpoint.UpdateTo(section, content, &buf)
	return strings.TrimSpace(buf.String())
}

func executeCheckpointList() string {
	var buf strings.Builder
	checkpoint.ListTo(&buf)
	return strings.TrimSpace(buf.String())
}

func executeLogCreate(topic string) string {
	var buf strings.Builder
	log.StartLogTo(topic, &buf)
	return strings.TrimSpace(buf.String())
}

func executeLogAppend(slug, message string) string {
	var buf strings.Builder
	log.AppendEntryTo(slug, message, &buf)
	return strings.TrimSpace(buf.String())
}

func executeLogSearch(query string) string {
	var buf strings.Builder
	log.SearchLogsTo(query, &buf)
	return strings.TrimSpace(buf.String())
}

func executeLogSummarize(slug string) string {
	var buf strings.Builder
	log.SummarizeLogTo(slug, &buf)
	return strings.TrimSpace(buf.String())
}

func executeLogList() string {
	var buf strings.Builder
	log.ListLogsTo(&buf)
	return strings.TrimSpace(buf.String())
}

func executeTaskCreate(slug string) string {
	var buf strings.Builder
	task.CreateTo(slug, &buf)
	return strings.TrimSpace(buf.String())
}

func executeTaskRead(slug string) string {
	var buf strings.Builder
	task.ReadTo(slug, &buf)
	return strings.TrimSpace(buf.String())
}

func executeTaskUpdateStatus(slug, status string) string {
	var buf strings.Builder
	task.UpdateStatusTo(slug, status, &buf)
	return strings.TrimSpace(buf.String())
}

func executeGitCommitCheckpoint(message string) string {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "Not in a git repository"
	}

	cmd = exec.Command("git", "add", ".bots")
	stderr.Reset()
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Sprintf("Error staging: %v\n%s", err, stderr.String())
	}

	cmd = exec.Command("git", "commit", "-m", message)
	var commitStdout, commitStderr bytes.Buffer
	cmd.Stdout = &commitStdout
	cmd.Stderr = &commitStderr
	if err := cmd.Run(); err != nil {
		return fmt.Sprintf("Error committing: %v\n%s", err, commitStderr.String())
	}

	return "Committed checkpoint changes successfully"
}

func executeGitBranchInfo() string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	var branchStderr bytes.Buffer
	cmd.Stderr = &branchStderr
	output, err := cmd.Output()
	if err != nil {
		return "Not in a git repository"
	}
	branch := strings.TrimSpace(string(output))

	cmd = exec.Command("git", "rev-parse", "--short", "HEAD")
	var commitStderr bytes.Buffer
	cmd.Stderr = &commitStderr
	output, err = cmd.Output()
	if err != nil {
		return branch
	}
	commit := strings.TrimSpace(string(output))

	_ = branchStderr
	_ = commitStderr

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
			"isError": false,
		},
	}
	encoder.Encode(response)
}

func sendToolError(encoder *json.Encoder, id interface{}, message string) {
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result": map[string]interface{}{
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": message,
				},
			},
			"isError": true,
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
