package mcp

import (
	"errors"
	"reflect"
	"testing"
)

func TestCatalogCoLocatesToolSchemaAndHandler(t *testing.T) {
	catalog := NewCatalog([]Tool{
		{
			Name:        "echo",
			Description: "Echo a message",
			InputSchema: objectSchema(map[string]interface{}{
				"message": stringProperty("Message to echo"),
			}, "message"),
			Handler: func(arguments map[string]interface{}) (string, error) {
				message, err := requireString(arguments, "message")
				if err != nil {
					return "", err
				}
				return message, nil
			},
		},
	})

	schemas := catalog.ToolSchemas()
	if len(schemas) != 1 || schemas[0]["name"] != "echo" || schemas[0]["description"] != "Echo a message" {
		t.Fatalf("unexpected schemas: %#v", schemas)
	}

	got, err := catalog.Call("echo", map[string]interface{}{"message": "hello"})
	if err != nil {
		t.Fatalf("Call returned error: %v", err)
	}
	if got != "hello" {
		t.Fatalf("expected hello, got %q", got)
	}

	_, err = catalog.Call("echo", map[string]interface{}{})
	if !errors.Is(err, ErrMissingArgument) {
		t.Fatalf("expected ErrMissingArgument, got %v", err)
	}

	_, err = catalog.Call("echo", map[string]interface{}{"message": "   "})
	if !errors.Is(err, ErrMissingArgument) {
		t.Fatalf("expected ErrMissingArgument for blank message, got %v", err)
	}
}

func TestRequireStringAllowEmptyPreservesIntentionalEmptyContent(t *testing.T) {
	got, err := requireStringAllowEmpty(map[string]interface{}{"content": ""}, "content")
	if err != nil {
		t.Fatalf("requireStringAllowEmpty returned error: %v", err)
	}
	if got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

func TestDefaultCatalogIncludesDocumentTaskLogAndGitTools(t *testing.T) {
	catalog := DefaultCatalog()
	var names []string
	for _, tool := range catalog.Tools {
		names = append(names, tool.Name)
	}

	want := []string{
		"checkpoint_read",
		"checkpoint_update",
		"checkpoint_list",
		"log_create",
		"log_append",
		"log_search",
		"log_summarize",
		"log_list",
		"task_create",
		"task_read",
		"task_update_status",
		"git_commit_checkpoint",
		"git_branch_info",
	}
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("unexpected tool names:\nwant %#v\ngot  %#v", want, names)
	}
}
