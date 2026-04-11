
package main

import "fmt"

func (s *MCPServer) handleToolCall(id interface{}, toolName string, args map[string]interface{}) {
	switch toolName {
	case "list_sites":
		s.handleListSites(id)
	case "get_site":
		s.handleGetSite(id, args)
	case "get_site_routes":
		s.handleGetSiteRoutes(id, args)
	case "add_site":
		s.handleAddSite(id, args)
	case "remove_site":
		s.handleRemoveSite(id, args)
	case "update_upstream":
		s.handleUpdateUpstream(id, args)
	case "add_path_route":
		s.handleAddPathRoute(id, args)
	case "remove_path_route":
		s.handleRemovePathRoute(id, args)
	case "add_basic_auth":
		s.handleAddBasicAuth(id, args)
	case "remove_basic_auth":
		s.handleRemoveBasicAuth(id, args)
	case "add_request_header":
		s.handleAddRequestHeader(id, args)
	case "add_response_header":
		s.handleAddResponseHeader(id, args)
	case "add_file_server":
		s.handleAddFileServer(id, args)
	case "update_file_root":
		s.handleUpdateFileRoot(id, args)
	case "add_wip_environment":
		s.handleAddWipEnvironment(id, args)
	case "export_config":
		s.handleExportConfig(id)
	case "adapt_caddyfile":
		s.handleAdaptCaddyfile(id, args)
	case "validate_config":
		s.handleValidateConfig(id, args)
	case "caddy_status":
		s.handleCaddyStatus(id)
	case "stop_caddy":
		s.handleStopCaddy(id)
	default:
		s.sendError(id, -32601, fmt.Sprintf("Unknown tool: %s", toolName))
	}
}
