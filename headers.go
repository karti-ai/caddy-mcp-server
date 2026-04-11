package main

import "fmt"

func (c *CaddyClient) AddHeaderUp(domain, header, value string) error {
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
						c.addHeaderToRoute(routeMap, header, value, "request")
						routes[routeIdx] = routeMap
						configPath := fmt.Sprintf("apps/http/servers/%s/routes", srvName)
						return c.SetConfig(configPath, routes)
					}
				}
			}
		}
	}
	return fmt.Errorf("site not found: %s", domain)
}

func (c *CaddyClient) AddHeaderDown(domain, header, value string) error {
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
						c.addHeaderToRoute(routeMap, header, value, "response")
						routes[routeIdx] = routeMap
						configPath := fmt.Sprintf("apps/http/servers/%s/routes", srvName)
						return c.SetConfig(configPath, routes)
					}
				}
			}
		}
	}
	return fmt.Errorf("site not found: %s", domain)
}

func (c *CaddyClient) addHeaderToRoute(route map[string]interface{}, header, value, direction string) {
	handle, _ := route["handle"].([]interface{})
	for i, h := range handle {
		handlerMap, ok := h.(map[string]interface{})
		if !ok {
			continue
		}
		if handlerMap["handler"] == "subroute" {
			subRoutes, _ := handlerMap["routes"].([]interface{})
			if len(subRoutes) > 0 {
				mainRoute, _ := subRoutes[0].(map[string]interface{})
				handlers, _ := mainRoute["handle"].([]interface{})
				
				headersHandler := map[string]interface{}{
					"handler": "headers",
				}
				if direction == "request" {
					headersHandler["request"] = map[string]interface{}{
						"set": map[string]interface{}{
							header: []string{value},
						},
					}
				} else {
					headersHandler["response"] = map[string]interface{}{
						"set": map[string]interface{}{
							header: []string{value},
						},
					}
				}
				
				// Find reverse_proxy and insert headers before it
				for j, handler := range handlers {
					hm, _ := handler.(map[string]interface{})
					if hm["handler"] == "reverse_proxy" {
						// Insert before reverse_proxy
						newHandlers := append(handlers[:j], headersHandler)
						newHandlers = append(newHandlers, handlers[j:]...)
						mainRoute["handle"] = newHandlers
						subRoutes[0] = mainRoute
						handlerMap["routes"] = subRoutes
						handle[i] = handlerMap
						route["handle"] = handle
						return
					}
				}
			}
		}
	}
}
