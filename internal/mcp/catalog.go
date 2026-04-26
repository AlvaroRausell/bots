package mcp

import (
	"errors"
	"fmt"
	"strings"

	"bots/internal/checkpoint"
	gitintegration "bots/internal/git"
	sessionlog "bots/internal/log"
	"bots/internal/task"
	"bots/internal/workspace"
)

var (
	ErrToolNotFound    = errors.New("unknown tool")
	ErrMissingArgument = errors.New("missing required argument")
)

type ToolHandler func(arguments map[string]interface{}) (string, error)

type Tool struct {
	Name        string
	Description string
	InputSchema map[string]interface{}
	Handler     ToolHandler
}

type Catalog struct {
	Tools []Tool
}

func NewCatalog(tools []Tool) Catalog {
	return Catalog{Tools: tools}
}

func DefaultCatalog() Catalog {
	return NewCatalog([]Tool{
		checkpointReadTool(),
		checkpointUpdateTool(),
		checkpointListTool(),
		logCreateTool(),
		logAppendTool(),
		logSearchTool(),
		logSummarizeTool(),
		logListTool(),
		taskCreateTool(),
		taskReadTool(),
		taskUpdateStatusTool(),
		gitCommitCheckpointTool(),
		gitBranchInfoTool(),
	})
}

func (c Catalog) ToolSchemas() []map[string]interface{} {
	tools := make([]map[string]interface{}, 0, len(c.Tools))
	for _, tool := range c.Tools {
		tools = append(tools, map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": tool.InputSchema,
		})
	}
	return tools
}

func (c Catalog) Call(name string, arguments map[string]interface{}) (string, error) {
	if arguments == nil {
		arguments = make(map[string]interface{})
	}
	for _, tool := range c.Tools {
		if tool.Name == name {
			return tool.Handler(arguments)
		}
	}
	return "", fmt.Errorf("%w: %s", ErrToolNotFound, name)
}

func checkpointReadTool() Tool {
	return Tool{
		Name:        "checkpoint_read",
		Description: "Read the current CHECKPOINTS.md content",
		InputSchema: objectSchema(nil),
		Handler: func(arguments map[string]interface{}) (string, error) {
			store, err := checkpoint.NewDefaultStore()
			if err != nil {
				return "", err
			}
			result, err := store.Read()
			if err != nil {
				return "", err
			}
			var b strings.Builder
			checkpoint.WriteReadResult(&b, result)
			return strings.TrimSpace(b.String()), nil
		},
	}
}

func checkpointUpdateTool() Tool {
	return Tool{
		Name:        "checkpoint_update",
		Description: "Update a specific section in CHECKPOINTS.md",
		InputSchema: objectSchema(map[string]interface{}{
			"section": stringProperty("Section name to update"),
			"content": stringProperty("New content for the section"),
		}, "section", "content"),
		Handler: func(arguments map[string]interface{}) (string, error) {
			section, err := requireString(arguments, "section")
			if err != nil {
				return "", err
			}
			content, err := requireStringAllowEmpty(arguments, "content")
			if err != nil {
				return "", err
			}
			store, err := checkpoint.NewDefaultStore()
			if err != nil {
				return "", err
			}
			result, err := store.Update(section, content)
			if err != nil {
				return "", err
			}
			var b strings.Builder
			checkpoint.WriteUpdateResult(&b, result)
			return strings.TrimSpace(b.String()), nil
		},
	}
}

func checkpointListTool() Tool {
	return Tool{
		Name:        "checkpoint_list",
		Description: "List all checkpoint sections",
		InputSchema: objectSchema(nil),
		Handler: func(arguments map[string]interface{}) (string, error) {
			store, err := checkpoint.NewDefaultStore()
			if err != nil {
				return "", err
			}
			result, err := store.List()
			if err != nil {
				return "", err
			}
			var b strings.Builder
			checkpoint.WriteListResult(&b, result)
			return strings.TrimSpace(b.String()), nil
		},
	}
}

func logCreateTool() Tool {
	return Tool{
		Name:        "log_create",
		Description: "Create a new session log file",
		InputSchema: objectSchema(map[string]interface{}{
			"topic": stringProperty("Topic/title for the session log"),
		}, "topic"),
		Handler: func(arguments map[string]interface{}) (string, error) {
			topic, err := requireString(arguments, "topic")
			if err != nil {
				return "", err
			}
			store, err := sessionlog.NewDefaultStore()
			if err != nil {
				return "", err
			}
			result, err := store.Start(topic)
			if err != nil {
				return "", err
			}
			var b strings.Builder
			sessionlog.WriteStartResult(&b, result)
			return strings.TrimSpace(b.String()), nil
		},
	}
}

func logAppendTool() Tool {
	return Tool{
		Name:        "log_append",
		Description: "Append an entry to a session log",
		InputSchema: objectSchema(map[string]interface{}{
			"slug":    stringProperty("Log file slug or partial name"),
			"message": stringProperty("Entry content to append"),
		}, "slug", "message"),
		Handler: func(arguments map[string]interface{}) (string, error) {
			slug, err := requireString(arguments, "slug")
			if err != nil {
				return "", err
			}
			message, err := requireString(arguments, "message")
			if err != nil {
				return "", err
			}
			store, err := sessionlog.NewDefaultStore()
			if err != nil {
				return "", err
			}
			result, err := store.Append(slug, message)
			if err != nil {
				return "", err
			}
			var b strings.Builder
			sessionlog.WriteAppendResult(&b, result)
			return strings.TrimSpace(b.String()), nil
		},
	}
}

func logSearchTool() Tool {
	return Tool{
		Name:        "log_search",
		Description: "Search across all session logs",
		InputSchema: objectSchema(map[string]interface{}{
			"query": stringProperty("Search query"),
		}, "query"),
		Handler: func(arguments map[string]interface{}) (string, error) {
			query, err := requireString(arguments, "query")
			if err != nil {
				return "", err
			}
			store, err := sessionlog.NewDefaultStore()
			if err != nil {
				return "", err
			}
			result, err := store.Search(query)
			if err != nil {
				return "", err
			}
			var b strings.Builder
			sessionlog.WriteSearchResult(&b, result)
			return strings.TrimSpace(b.String()), nil
		},
	}
}

func logSummarizeTool() Tool {
	return Tool{
		Name:        "log_summarize",
		Description: "Generate a summary of decisions from a log",
		InputSchema: objectSchema(map[string]interface{}{
			"slug": stringProperty("Log file slug or partial name"),
		}, "slug"),
		Handler: func(arguments map[string]interface{}) (string, error) {
			slug, err := requireString(arguments, "slug")
			if err != nil {
				return "", err
			}
			store, err := sessionlog.NewDefaultStore()
			if err != nil {
				return "", err
			}
			result, err := store.Summarize(slug)
			if err != nil {
				return "", err
			}
			var b strings.Builder
			sessionlog.WriteSummaryResult(&b, result)
			return strings.TrimSpace(b.String()), nil
		},
	}
}

func logListTool() Tool {
	return Tool{
		Name:        "log_list",
		Description: "List all session log files",
		InputSchema: objectSchema(nil),
		Handler: func(arguments map[string]interface{}) (string, error) {
			store, err := sessionlog.NewDefaultStore()
			if err != nil {
				return "", err
			}
			result, err := store.List()
			if err != nil {
				return "", err
			}
			var b strings.Builder
			sessionlog.WriteListResult(&b, result)
			return strings.TrimSpace(b.String()), nil
		},
	}
}

func taskCreateTool() Tool {
	return Tool{
		Name:        "task_create",
		Description: "Create a new task file",
		InputSchema: objectSchema(map[string]interface{}{
			"slug": stringProperty("Task slug (e.g., 'phase-1.5')"),
		}, "slug"),
		Handler: func(arguments map[string]interface{}) (string, error) {
			slug, err := requireString(arguments, "slug")
			if err != nil {
				return "", err
			}
			store, err := task.NewDefaultStore()
			if err != nil {
				return "", err
			}
			result, err := store.Create(slug)
			if err != nil {
				return "", err
			}
			var b strings.Builder
			task.WriteCreateResult(&b, result)
			return strings.TrimSpace(b.String()), nil
		},
	}
}

func taskReadTool() Tool {
	return Tool{
		Name:        "task_read",
		Description: "Read a task file",
		InputSchema: objectSchema(map[string]interface{}{
			"slug": stringProperty("Task slug"),
		}, "slug"),
		Handler: func(arguments map[string]interface{}) (string, error) {
			slug, err := requireString(arguments, "slug")
			if err != nil {
				return "", err
			}
			store, err := task.NewDefaultStore()
			if err != nil {
				return "", err
			}
			result, err := store.Read(slug)
			if err != nil {
				return "", err
			}
			var b strings.Builder
			task.WriteReadResult(&b, result)
			return strings.TrimSpace(b.String()), nil
		},
	}
}

func taskUpdateStatusTool() Tool {
	return Tool{
		Name:        "task_update_status",
		Description: "Update task status",
		InputSchema: objectSchema(map[string]interface{}{
			"slug":   stringProperty("Task slug"),
			"status": enumStringProperty("New status", []string{"PENDING", "IN_PROGRESS", "READY_FOR_REVIEW", "CHANGES_REQUESTED", "DONE"}),
		}, "slug", "status"),
		Handler: func(arguments map[string]interface{}) (string, error) {
			slug, err := requireString(arguments, "slug")
			if err != nil {
				return "", err
			}
			status, err := requireString(arguments, "status")
			if err != nil {
				return "", err
			}
			store, err := task.NewDefaultStore()
			if err != nil {
				return "", err
			}
			result, err := store.UpdateStatus(slug, status)
			if err != nil {
				return "", err
			}
			var b strings.Builder
			task.WriteUpdateStatusResult(&b, result)
			return strings.TrimSpace(b.String()), nil
		},
	}
}

func gitCommitCheckpointTool() Tool {
	return Tool{
		Name:        "git_commit_checkpoint",
		Description: "Commit checkpoint changes to git",
		InputSchema: objectSchema(map[string]interface{}{
			"message": stringProperty("Commit message"),
		}, "message"),
		Handler: func(arguments map[string]interface{}) (string, error) {
			message, err := requireString(arguments, "message")
			if err != nil {
				return "", err
			}
			ws, err := workspace.FromCurrent(false)
			if err != nil {
				return "", err
			}
			result, err := gitintegration.NewWorkspaceIntegration(ws, nil).CommitCheckpoint(message)
			if err != nil {
				return "", err
			}
			return gitintegration.FormatCommitResult(result), nil
		},
	}
}

func gitBranchInfoTool() Tool {
	return Tool{
		Name:        "git_branch_info",
		Description: "Get current git branch information",
		InputSchema: objectSchema(nil),
		Handler: func(arguments map[string]interface{}) (string, error) {
			info, err := gitintegration.DefaultIntegration().BranchInfo()
			if err != nil {
				return "", err
			}
			return gitintegration.FormatBranchInfo(info), nil
		},
	}
}

func requireString(arguments map[string]interface{}, name string) (string, error) {
	stringValue, err := requireStringAllowEmpty(arguments, name)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(stringValue) == "" {
		return "", fmt.Errorf("%w: %s", ErrMissingArgument, name)
	}
	return stringValue, nil
}

func requireStringAllowEmpty(arguments map[string]interface{}, name string) (string, error) {
	value, ok := arguments[name]
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrMissingArgument, name)
	}
	stringValue, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrMissingArgument, name)
	}
	return stringValue, nil
}

func objectSchema(properties map[string]interface{}, required ...string) map[string]interface{} {
	if properties == nil {
		properties = map[string]interface{}{}
	}
	schema := map[string]interface{}{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

func stringProperty(description string) map[string]interface{} {
	return map[string]interface{}{
		"type":        "string",
		"description": description,
	}
}

func enumStringProperty(description string, values []string) map[string]interface{} {
	property := stringProperty(description)
	property["enum"] = values
	return property
}
