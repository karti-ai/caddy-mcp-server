package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// Site Management Handlers
func (s *MCPServer) handleListSites(id interface{}) {
	sites, err := s.client.ListSites()
	if err != nil {
		s.sendError(id, -32000, err.Error())
		return
	}

	var output string
	for _, site := range sites {
		domain, _ := site["domain"].(string)
		upstreams := s.client.extractUpstreams(site["route"].(map[string]interface{}))
		output += fmt.Sprintf("📍 %s\n", domain)
		if len(upstreams) > 0 {
			output += fmt.Sprintf("   → %v\n", upstreams)
		}
		if s.client.hasBasicAuth(site["route"].(map[string]interface{})) {
			output += "   🔒 Auth enabled\n"
		}
		output += "\n"
	}
	if len(sites) == 0 {
		output = "No sites configured\n"
	}
	s.sendTextResult(id, output)
}

func (s *MCPServer) handleGetSite(id interface{}, args map[string]interface{}) {
	domain, _ := args["domain"].(string)
	site, err := s.client.GetSite(domain)
	if err != nil {
		s.sendError(id, -32000, err.Error())
		return
	}

	output := fmt.Sprintf("Domain: %s\n", domain)
	output += fmt.Sprintf("Server: %s\n", site["server"])
	output += fmt.Sprintf("Upstreams: %v\n", site["upstreams"])
	output += fmt.Sprintf("Auth: %v\n", site["has_auth"])
	handlers, _ := json.MarshalIndent(site["handlers"], "", "  ")
	output += fmt.Sprintf("Handlers: %s\n", string(handlers))
	s.sendTextResult(id, output)
}

func (s *MCPServer) handleAddSite(id interface{}, args map[string]interface{}) {
	domain, _ := args["domain"].(string)
	if err := validateDomain(domain); err != nil {
		s.sendError(id, -32000, err.Error())
		return
	}
	upstream, _ := args["upstream"].(string)
	if err := s.client.AddSite(domain, upstream); err != nil {
		s.sendError(id, -32000, err.Error())
		return
	}
	s.sendTextResult(id, fmt.Sprintf("✅ Added %s → %s", domain, upstream))
}

func (s *MCPServer) handleRemoveSite(id interface{}, args map[string]interface{}) {
	domain, _ := args["domain"].(string)
	if err := s.client.RemoveSite(domain); err != nil {
		s.sendError(id, -32000, err.Error())
		return
	}
	s.sendTextResult(id, fmt.Sprintf("✅ Removed %s", domain))
}

func (s *MCPServer) handleUpdateUpstream(id interface{}, args map[string]interface{}) {
	domain, _ := args["domain"].(string)
	upstream, _ := args["upstream"].(string)
	if err := s.client.UpdateUpstream(domain, upstream); err != nil {
		s.sendError(id, -32000, err.Error())
		return
	}
	s.sendTextResult(id, fmt.Sprintf("✅ Updated %s → %s", domain, upstream))
}

// Path Route Handlers
func (s *MCPServer) handleAddPathRoute(id interface{}, args map[string]interface{}) {
	domain, _ := args["domain"].(string)
	if err := validateDomain(domain); err != nil {
		s.sendError(id, -32000, err.Error())
		return
	}
	path, _ := args["path"].(string)
	if err := validatePath(path); err != nil {
		s.sendError(id, -32000, err.Error())
		return
	}
	upstream, _ := args["upstream"].(string)
	if err := s.client.AddPathRoute(domain, path, upstream); err != nil {
		s.sendError(id, -32000, err.Error())
		return
	}
	s.sendTextResult(id, fmt.Sprintf("✅ Added %s%s/* → %s", domain, path, upstream))
}

func (s *MCPServer) handleRemovePathRoute(id interface{}, args map[string]interface{}) {
	domain, _ := args["domain"].(string)
	if err := validateDomain(domain); err != nil {
		s.sendError(id, -32000, err.Error())
		return
	}
	path, _ := args["path"].(string)
	if err := validatePath(path); err != nil {
		s.sendError(id, -32000, err.Error())
		return
	}
	if err := s.client.RemovePathRoute(domain, path); err != nil {
		s.sendError(id, -32000, err.Error())
		return
	}
	s.sendTextResult(id, fmt.Sprintf("✅ Removed %s%s/*", domain, path))
}

// Auth Handlers
func (s *MCPServer) handleAddBasicAuth(id interface{}, args map[string]interface{}) {
	domain, _ := args["domain"].(string)
	username, _ := args["username"].(string)
	password, _ := args["password"].(string)
	if err := s.client.AddBasicAuth(domain, username, password); err != nil {
		s.sendError(id, -32000, err.Error())
		return
	}
	s.sendTextResult(id, fmt.Sprintf("✅ Added basic auth for %s (user: %s)", domain, username))
}

func (s *MCPServer) handleRemoveBasicAuth(id interface{}, args map[string]interface{}) {
	domain, _ := args["domain"].(string)
	if err := s.client.RemoveBasicAuth(domain); err != nil {
		s.sendError(id, -32000, err.Error())
		return
	}
	s.sendTextResult(id, fmt.Sprintf("✅ Removed basic auth from %s", domain))
}

// Header Handlers
func (s *MCPServer) handleAddRequestHeader(id interface{}, args map[string]interface{}) {
	domain, _ := args["domain"].(string)
	if err := validateDomain(domain); err != nil {
		s.sendError(id, -32000, err.Error())
		return
	}
	header, _ := args["header"].(string)
	if err := validateHeaderName(header); err != nil {
		s.sendError(id, -32000, err.Error())
		return
	}
	value, _ := args["value"].(string)
	if err := s.client.AddHeaderUp(domain, header, value); err != nil {
		s.sendError(id, -32000, err.Error())
		return
	}
	s.sendTextResult(id, fmt.Sprintf("✅ Added request header %s: %s for %s", header, value, domain))
}

func (s *MCPServer) handleAddResponseHeader(id interface{}, args map[string]interface{}) {
	domain, _ := args["domain"].(string)
	if err := validateDomain(domain); err != nil {
		s.sendError(id, -32000, err.Error())
		return
	}
	header, _ := args["header"].(string)
	if err := validateHeaderName(header); err != nil {
		s.sendError(id, -32000, err.Error())
		return
	}
	value, _ := args["value"].(string)
	if err := s.client.AddHeaderDown(domain, header, value); err != nil {
		s.sendError(id, -32000, err.Error())
		return
	}
	s.sendTextResult(id, fmt.Sprintf("✅ Added response header %s: %s for %s", header, value, domain))
}

// File Server Handlers
func (s *MCPServer) handleAddFileServer(id interface{}, args map[string]interface{}) {
	domain, _ := args["domain"].(string)
	root, _ := args["root"].(string)
	browse, _ := args["browse"].(bool)
	if err := s.client.AddFileServer(domain, root, browse); err != nil {
		s.sendError(id, -32000, err.Error())
		return
	}
	s.sendTextResult(id, fmt.Sprintf("✅ Added file server %s → %s (browse: %v)", domain, root, browse))
}

func (s *MCPServer) handleUpdateFileRoot(id interface{}, args map[string]interface{}) {
	domain, _ := args["domain"].(string)
	root, _ := args["root"].(string)
	if err := s.client.UpdateFileServerRoot(domain, root); err != nil {
		s.sendError(id, -32000, err.Error())
		return
	}
	s.sendTextResult(id, fmt.Sprintf("✅ Updated %s root → %s", domain, root))
}

// Review Environment Handler
func (s *MCPServer) handleAddWipEnvironment(id interface{}, args map[string]interface{}) {
	project, _ := args["project"].(string)
	portFloat, _ := args["port"].(float64)
	port := int(portFloat)

	reviewDomain := os.Getenv("REVIEW_DOMAIN")
	if reviewDomain == "" {
		reviewDomain = "review.example.com"
	}
	path := fmt.Sprintf("/%s/wip", project)
	upstream := fmt.Sprintf("localhost:%d", port)

	if err := s.client.AddPathRoute(reviewDomain, path, upstream); err != nil {
		s.sendError(id, -32000, err.Error())
		return
	}
	s.sendTextResult(id, fmt.Sprintf("✅ Added WIP https://%s%s/* → %s", reviewDomain, path, upstream))
}

// Config Handlers
func (s *MCPServer) handleExportConfig(id interface{}) {
	config, err := s.client.GetConfig()
	if err != nil {
		s.sendError(id, -32000, err.Error())
		return
	}
	data, _ := json.MarshalIndent(config, "", "  ")
	s.sendTextResult(id, string(data))
}

func (s *MCPServer) handleAdaptCaddyfile(id interface{}, args map[string]interface{}) {
	caddyfile, _ := args["caddyfile"].(string)
	result, err := s.client.AdaptCaddyfile(caddyfile)
	if err != nil {
		s.sendError(id, -32000, err.Error())
		return
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	s.sendTextResult(id, string(data))
}

func (s *MCPServer) handleValidateConfig(id interface{}, args map[string]interface{}) {
	configArg, _ := args["config"].(map[string]interface{})
	if err := s.client.ValidateConfig(configArg); err != nil {
		s.sendError(id, -32000, err.Error())
		return
	}
	s.sendTextResult(id, "✅ Configuration is valid")
}

// Control Handlers
func (s *MCPServer) handleCaddyStatus(id interface{}) {
	sites, err := s.client.ListSites()
	if err != nil {
		s.sendTextResult(id, fmt.Sprintf("❌ Caddy unreachable: %v", err))
		return
	}
	s.sendTextResult(id, fmt.Sprintf("✅ Caddy running (%d sites configured)", len(sites)))
}

func (s *MCPServer) handleStopCaddy(id interface{}) {
	if err := s.client.Stop(); err != nil {
		s.sendError(id, -32000, err.Error())
		return
	}
	s.sendTextResult(id, "✅ Caddy stopped")
}

// Helper functions
func (s *MCPServer) sendTextResult(id interface{}, text string) {
	s.sendResult(id, map[string]interface{}{
		"content": []interface{}{
			map[string]interface{}{"type": "text", "text": text},
		},
	})
}

func (s *MCPServer) sendResult(id interface{}, result interface{}) {
	resp := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"result":  result,
	}
	data, _ := json.Marshal(resp)
	fmt.Println(string(data))
}

func (s *MCPServer) sendError(id interface{}, code int, message string) {
	resp := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      id,
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	}
	data, _ := json.Marshal(resp)
	fmt.Println(string(data))
}
