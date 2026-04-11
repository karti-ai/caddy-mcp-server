package main

import "fmt"

func (c *CaddyClient) AddFileServer(domain, root string, browse bool) error {
	route := map[string]interface{}{
		"handle": []interface{}{
			map[string]interface{}{
				"handler": "vars",
				"root":    root,
			},
			map[string]interface{}{
				"handler":     "file_server",
				"browse":      browse,
				"hide":        []string{".git", ".env"},
			},
		},
		"match":    []interface{}{map[string]interface{}{"host": []string{domain}}},
		"terminal": true,
	}

	config, _ := c.GetConfig()
	apps, _ := config["apps"].(map[string]interface{})
	httpApps, _ := apps["http"].(map[string]interface{})
	servers, _ := httpApps["servers"].(map[string]interface{})

	for srvName := range servers {
		routesPath := fmt.Sprintf("apps/http/servers/%s/routes", srvName)
		routes, _ := c.GetConfigAtPath(routesPath)
		if routes != nil {
			routesList, _ := routes.([]interface{})
			routesList = append(routesList, route)
			return c.SetConfig(routesPath, routesList)
		}
	}

	return c.SetConfig("apps/http/servers/srv0/routes", []interface{}{route})
}

func (c *CaddyClient) UpdateFileServerRoot(domain, root string) error {
	config, err := c.GetConfig()
	if err != nil {
		return err
	}

	apps, _ := config["apps"].(map[string]interface{})
	httpApps, _ := apps["http"].(map[string]interface{})
	servers, _ := httpApps["servers"].(map[string]interface{})

	for srvName, server := range servers {
		srv, _ := server.(map[string]interface{})
		routes, _ := srv["routes"].([]interface{})

		for routeIdx, route := range routes {
			routeMap, _ := route.(map[string]interface{})
			matches, _ := routeMap["match"].([]interface{})

			for _, match := range matches {
				matchMap, _ := match.(map[string]interface{})
				hosts, _ := matchMap["host"].([]interface{})
				if len(hosts) > 0 {
					d, _ := hosts[0].(string)
					if d == domain {
						handle, _ := routeMap["handle"].([]interface{})
						for i, h := range handle {
							handlerMap, ok := h.(map[string]interface{})
							if !ok {
								continue
							}
							if handlerMap["handler"] == "vars" {
								handlerMap["root"] = root
								handle[i] = handlerMap
								routeMap["handle"] = handle
								routes[routeIdx] = routeMap

								configPath := fmt.Sprintf("apps/http/servers/%s/routes", srvName)
								return c.SetConfig(configPath, routes)
							}
						}
					}
				}
			}
		}
	}
	return fmt.Errorf("file server not found for: %s", domain)
}
