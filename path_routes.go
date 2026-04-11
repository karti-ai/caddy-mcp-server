package main

import (
	"fmt"
	"strings"
)

func (c *CaddyClient) AddPathRoute(domain, path, upstream string) error {
	config, err := c.GetConfig()
	if err != nil {
		return err
	}

	apps, _ := config["apps"].(map[string]interface{})
	httpApps, _ := apps["http"].(map[string]interface{})
	servers, _ := httpApps["servers"].(map[string]interface{})

	newRoute := map[string]interface{}{
		"group": "group1",
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
		"match": []interface{}{
			map[string]interface{}{
				"path": []string{path + "/*"},
			},
		},
	}

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
							if handlerMap["handler"] == "subroute" {
								subRoutes, _ := handlerMap["routes"].([]interface{})
								subRoutes = append(subRoutes, newRoute)
								handlerMap["routes"] = subRoutes
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
	return fmt.Errorf("site not found: %s", domain)
}

func (c *CaddyClient) RemovePathRoute(domain, path string) error {
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
							if handlerMap["handler"] == "subroute" {
								subRoutes, _ := handlerMap["routes"].([]interface{})
								var newSubRoutes []interface{}
								found := false
								for _, sr := range subRoutes {
									srMap, ok := sr.(map[string]interface{})
									if !ok {
										newSubRoutes = append(newSubRoutes, sr)
										continue
									}
									srMatches, _ := srMap["match"].([]interface{})
									for _, srm := range srMatches {
										srmMap, _ := srm.(map[string]interface{})
										paths, _ := srmMap["path"].([]interface{})
										for _, p := range paths {
											ps, _ := p.(string)
											if strings.HasPrefix(ps, path) {
												found = true
												break
											}
										}
									}
									if !found {
										newSubRoutes = append(newSubRoutes, sr)
									}
								}
								if found {
									handlerMap["routes"] = newSubRoutes
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
	}
	return fmt.Errorf("site or path not found: %s/%s", domain, path)
}
