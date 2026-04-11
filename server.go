package main

import (
	"bufio"
	"encoding/json"
	"os"
)

type MCPServer struct {
	client *CaddyClient
}

func main() {
	caddyHost := os.Getenv("CADDY_HOST")
	if caddyHost == "" {
		caddyHost = "localhost:2019"
	}

	client := NewCaddyClient(caddyHost)
	server := &MCPServer{client: client}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var req map[string]interface{}
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			server.sendError(nil, -32700, "Parse error")
			continue
		}

		server.handleRequest(req)
	}
}

func (s *MCPServer) handleRequest(req map[string]interface{}) {
	method, _ := req["method"].(string)
	id := req["id"]

	switch method {
	case "initialize":
		s.sendResult(id, map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{"tools": map[string]interface{}{}},
			"serverInfo": map[string]interface{}{
				"name":    "caddy-mcp",
				"version": "2.0.0",
			},
		})

	case "tools/list":
		s.sendResult(id, map[string]interface{}{"tools": getAllTools()})

	case "tools/call":
		params, _ := req["params"].(map[string]interface{})
		toolName, _ := params["name"].(string)
		args, _ := params["arguments"].(map[string]interface{})
		s.handleToolCall(id, toolName, args)

	default:
		s.sendError(id, -32601, "Method not found")
	}
}

func getAllTools() []interface{} {
	return []interface{}{
		// Site management
		map[string]interface{}{
			"name":        "list_sites",
			"description": "List all configured Caddy sites with their upstreams",
			"inputSchema": map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
		},
		map[string]interface{}{
			"name":        "get_site",
			"description": "Get detailed configuration for a specific site",
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{"domain": map[string]interface{}{"type": "string", "description": "Domain name"}},
				"required":   []string{"domain"},
			},
		},
		map[string]interface{}{
			"name":        "get_site_routes",
			"description": "Get all routes (including path-based) for a site",
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{"domain": map[string]interface{}{"type": "string", "description": "Domain name"}},
				"required":   []string{"domain"},
			},
		},
		map[string]interface{}{
			"name":        "add_site",
			"description": "Add a new site with reverse proxy",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"domain":   map[string]interface{}{"type": "string", "description": "Domain name"},
					"upstream": map[string]interface{}{"type": "string", "description": "Upstream address (e.g., localhost:3000)"},
				},
				"required": []string{"domain", "upstream"},
			},
		},
		map[string]interface{}{
			"name":        "remove_site",
			"description": "Remove a site configuration",
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{"domain": map[string]interface{}{"type": "string", "description": "Domain to remove"}},
				"required":   []string{"domain"},
			},
		},
		map[string]interface{}{
			"name":        "update_upstream",
			"description": "Update the upstream address for a site",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"domain":   map[string]interface{}{"type": "string", "description": "Domain name"},
					"upstream": map[string]interface{}{"type": "string", "description": "New upstream address"},
				},
				"required": []string{"domain", "upstream"},
			},
		},
		// Path-based routing
		map[string]interface{}{
			"name":        "add_path_route",
			"description": "Add a path-based route to a site (e.g., /api/* → localhost:3000)",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"domain":   map[string]interface{}{"type": "string", "description": "Domain name"},
					"path":     map[string]interface{}{"type": "string", "description": "Path pattern (e.g., /api)"},
					"upstream": map[string]interface{}{"type": "string", "description": "Upstream address"},
				},
				"required": []string{"domain", "path", "upstream"},
			},
		},
		map[string]interface{}{
			"name":        "remove_path_route",
			"description": "Remove a path-based route from a site",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"domain": map[string]interface{}{"type": "string", "description": "Domain name"},
					"path":   map[string]interface{}{"type": "string", "description": "Path pattern to remove"},
				},
				"required": []string{"domain", "path"},
			},
		},
		// Authentication
		map[string]interface{}{
			"name":        "add_basic_auth",
			"description": "Add basic authentication to a site",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"domain":   map[string]interface{}{"type": "string", "description": "Domain name"},
					"username": map[string]interface{}{"type": "string", "description": "Username"},
					"password": map[string]interface{}{"type": "string", "description": "Password (will be hashed)"},
				},
				"required": []string{"domain", "username", "password"},
			},
		},
		map[string]interface{}{
			"name":        "remove_basic_auth",
			"description": "Remove basic authentication from a site",
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{"domain": map[string]interface{}{"type": "string", "description": "Domain name"}},
				"required":   []string{"domain"},
			},
		},
		// Headers
		map[string]interface{}{
			"name":        "add_request_header",
			"description": "Add a request header sent to upstream",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"domain": map[string]interface{}{"type": "string", "description": "Domain name"},
					"header": map[string]interface{}{"type": "string", "description": "Header name"},
					"value":  map[string]interface{}{"type": "string", "description": "Header value"},
				},
				"required": []string{"domain", "header", "value"},
			},
		},
		map[string]interface{}{
			"name":        "add_response_header",
			"description": "Add a response header sent to client",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"domain": map[string]interface{}{"type": "string", "description": "Domain name"},
					"header": map[string]interface{}{"type": "string", "description": "Header name"},
					"value":  map[string]interface{}{"type": "string", "description": "Header value"},
				},
				"required": []string{"domain", "header", "value"},
			},
		},
		// File server
		map[string]interface{}{
			"name":        "add_file_server",
			"description": "Add a static file server for a domain",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"domain": map[string]interface{}{"type": "string", "description": "Domain name"},
					"root":   map[string]interface{}{"type": "string", "description": "Root directory path"},
					"browse": map[string]interface{}{"type": "boolean", "description": "Enable directory browsing"},
				},
				"required": []string{"domain", "root"},
			},
		},
		map[string]interface{}{
			"name":        "update_file_root",
			"description": "Update the root directory for a file server",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"domain": map[string]interface{}{"type": "string", "description": "Domain name"},
					"root":   map[string]interface{}{"type": "string", "description": "New root directory"},
				},
				"required": []string{"domain", "root"},
			},
		},
		// Review environment helpers
		map[string]interface{}{
			"name":        "add_wip_environment",
			"description": "Add a WIP review environment (e.g., /project/wip/* → port)",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"project": map[string]interface{}{"type": "string", "description": "Project name (e.g., my-project, frontend)"},
					"port":    map[string]interface{}{"type": "number", "description": "Port number"},
				},
				"required": []string{"project", "port"},
			},
		},
		// Configuration
		map[string]interface{}{
			"name":        "export_config",
			"description": "Export full Caddy configuration as JSON",
			"inputSchema": map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
		},
		map[string]interface{}{
			"name":        "adapt_caddyfile",
			"description": "Convert Caddyfile format to JSON",
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{"caddyfile": map[string]interface{}{"type": "string", "description": "Caddyfile content"}},
				"required":   []string{"caddyfile"},
			},
		},
		// Server control
		map[string]interface{}{
			"name":        "caddy_status",
			"description": "Check Caddy server health and status",
			"inputSchema": map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
		},
		map[string]interface{}{
			"name":        "validate_config",
			"description": "Validate a configuration without applying it",
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{"config": map[string]interface{}{"type": "object", "description": "Configuration to validate"}},
				"required":   []string{"config"},
			},
		},
		map[string]interface{}{
			"name":        "stop_caddy",
			"description": "Stop the Caddy server gracefully",
			"inputSchema": map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
		},
	}
}
