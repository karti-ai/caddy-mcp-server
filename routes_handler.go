package main

import (
	"encoding/json"
	"fmt"
)

func (s *MCPServer) handleGetSiteRoutes(id interface{}, args map[string]interface{}) {
	domain, _ := args["domain"].(string)
	config, err := s.client.GetConfig()
	if err != nil {
		s.sendError(id, -32000, err.Error())
		return
	}

	apps, _ := config["apps"].(map[string]interface{})
	httpApps, _ := apps["http"].(map[string]interface{})
	servers, _ := httpApps["servers"].(map[string]interface{})

	var output string
	for _, server := range servers {
		srv, _ := server.(map[string]interface{})
		routes, _ := srv["routes"].([]interface{})

		for _, route := range routes {
			routeMap, _ := route.(map[string]interface{})
			matches, _ := routeMap["match"].([]interface{})

			for _, match := range matches {
				matchMap, _ := match.(map[string]interface{})
				hosts, _ := matchMap["host"].([]interface{})
				if len(hosts) > 0 {
					d, _ := hosts[0].(string)
					if d == domain {
						output = fmt.Sprintf("Routes for %s:\n\n", domain)
						data, _ := json.MarshalIndent(routeMap, "", "  ")
						output += string(data)
						s.sendTextResult(id, output)
						return
					}
				}
			}
		}
	}
	s.sendError(id, -32000, fmt.Sprintf("site not found: %s", domain))
}
