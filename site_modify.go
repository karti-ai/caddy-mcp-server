package main

import "fmt"

func (c *CaddyClient) AddSite(domain, upstream string) error {
	route := map[string]interface{}{
		"handle": []interface{}{
			map[string]interface{}{
				"handler": "subroute",
				"routes": []interface{}{
					map[string]interface{}{
						"handle": []interface{}{
							map[string]interface{}{
								"handler":   "reverse_proxy",
								"upstreams": []interface{}{map[string]interface{}{"dial": upstream}},
							},
						},
					},
				},
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

func (c *CaddyClient) RemoveSite(domain string) error {
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

		var newRoutes []interface{}
		found := false
		for _, route := range routes {
			routeMap, _ := route.(map[string]interface{})
			matches, _ := routeMap["match"].([]interface{})

			for _, match := range matches {
				matchMap, _ := match.(map[string]interface{})
				hosts, _ := matchMap["host"].([]interface{})
				if len(hosts) > 0 {
					d, _ := hosts[0].(string)
					if d == domain {
						found = true
						continue
					}
				}
			}
			newRoutes = append(newRoutes, route)
		}

		if found {
			path := fmt.Sprintf("apps/http/servers/%s/routes", srvName)
			return c.SetConfig(path, newRoutes)
		}
	}
	return fmt.Errorf("site not found: %s", domain)
}

func (c *CaddyClient) UpdateUpstream(domain, upstream string) error {
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

		for i, route := range routes {
			routeMap, _ := route.(map[string]interface{})
			matches, _ := routeMap["match"].([]interface{})

			for _, match := range matches {
				matchMap, _ := match.(map[string]interface{})
				hosts, _ := matchMap["host"].([]interface{})
				if len(hosts) > 0 {
					d, _ := hosts[0].(string)
					if d == domain {
						newRoute := c.updateUpstreamInRoute(routeMap, upstream)
						routes[i] = newRoute
						path := fmt.Sprintf("apps/http/servers/%s/routes", srvName)
						return c.SetConfig(path, routes)
					}
				}
			}
		}
	}
	return fmt.Errorf("site not found: %s", domain)
}

func (c *CaddyClient) updateUpstreamInRoute(route map[string]interface{}, upstream string) map[string]interface{} {
	handle, _ := route["handle"].([]interface{})
	for i, h := range handle {
		handlerMap, ok := h.(map[string]interface{})
		if !ok {
			continue
		}
		if handlerMap["handler"] == "reverse_proxy" {
			handlerMap["upstreams"] = []interface{}{map[string]interface{}{"dial": upstream}}
			handle[i] = handlerMap
			route["handle"] = handle
			return route
		}
		if handlerMap["handler"] == "subroute" {
			subRoutes, _ := handlerMap["routes"].([]interface{})
			for j, sr := range subRoutes {
				srMap, ok := sr.(map[string]interface{})
				if !ok {
					continue
				}
				srHandlers, _ := srMap["handle"].([]interface{})
				for k, srH := range srHandlers {
					srHMap, ok := srH.(map[string]interface{})
					if !ok {
						continue
					}
					if srHMap["handler"] == "reverse_proxy" {
						srHMap["upstreams"] = []interface{}{map[string]interface{}{"dial": upstream}}
						srHandlers[k] = srHMap
						srMap["handle"] = srHandlers
						subRoutes[j] = srMap
						handlerMap["routes"] = subRoutes
						handle[i] = handlerMap
						route["handle"] = handle
						return route
					}
				}
			}
		}
	}
	return route
}
