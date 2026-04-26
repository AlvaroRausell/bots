package mcp

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Serve starts the MCP server on stdio.
func Serve() {
	ServeWithCatalog(DefaultCatalog())
}

func ServeWithCatalog(catalog Catalog) {
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
			handleToolsList(encoder, id, catalog)
		case "tools/call":
			params, ok := message["params"].(map[string]interface{})
			if !ok {
				sendError(encoder, -32602, "Invalid params", id)
				continue
			}
			handleToolsCall(encoder, id, params, catalog)
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

func handleToolsList(encoder *json.Encoder, id interface{}, catalog Catalog) {
	response := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result": map[string]interface{}{
			"tools": catalog.ToolSchemas(),
		},
	}
	encoder.Encode(response)
}

func handleToolsCall(encoder *json.Encoder, id interface{}, params map[string]interface{}, catalog Catalog) {
	name, ok := params["name"].(string)
	if !ok {
		sendError(encoder, -32602, "Invalid tool name", id)
		return
	}

	arguments, ok := params["arguments"].(map[string]interface{})
	if !ok {
		arguments = make(map[string]interface{})
	}

	result, err := catalog.Call(name, arguments)
	if err != nil {
		sendToolError(encoder, id, err.Error())
		return
	}

	sendSuccess(encoder, id, result)
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
