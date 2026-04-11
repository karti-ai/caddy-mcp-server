package main

import (
	"fmt"
)

func (c *CaddyClient) ListSites() ([]map[string]interface{}, error) {
	config, err := c.GetConfig()
	if err != nil {
		return nil, err
	}
	return c.extractSitesFromConfig(config)
}

func (c *CaddyClient) extractSitesFromConfig(config map[string]interface{}) ([]map[string]interface{}, error) {
	apps, _ := config["apps"].(map[string]interface{})
	httpApps, _ := apps["http"].(map[string]interface{})
	servers, _ := httpApps["servers"].(map[string]interface{})

	var sites []map[string]interface{}
	for serverName, server := range servers {
		srv, _ := server.(map[string]interface{})
		routes, _ := srv["routes"].([]interface{})

		for _, route := range routes {
			routeMap, _ := route.(map[string]interface{})
			matches, _ := routeMap["match"].([]interface{})

			for _, match := range matches {
				matchMap, _ := match.(map[string]interface{})
				hosts, _ := matchMap["host"].([]interface{})
				if len(hosts) > 0 {
					domain, _ := hosts[0].(string)
					site := map[string]interface{}{
						"domain": domain,
						"server": serverName,
						"route":  routeMap,
					}
					sites = append(sites, site)
				}
			}
		}
	}
	return sites, nil
}

func (c *CaddyClient) GetSite(domain string) (map[string]interface{}, error) {
	config, err := c.GetConfig()
	if err != nil {
		return nil, err
	}

	apps, _ := config["apps"].(map[string]interface{})
	httpApps, _ := apps["http"].(map[string]interface{})
	servers, _ := httpApps["servers"].(map[string]interface{})

	for serverName, server := range servers {
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
						site := map[string]interface{}{
							"domain":    domain,
							"server":    serverName,
							"route":     routeMap,
							"handlers":  c.extractHandlers(routeMap),
							"upstreams": c.extractUpstreams(routeMap),
							"has_auth":  c.hasBasicAuth(routeMap),
						}
						return site, nil
					}
				}
			}
		}
	}
	return nil, fmt.Errorf("site not found: %s", domain)
}

func (c *CaddyClient) extractHandlers(route map[string]interface{}) []map[string]interface{} {
	var handlers []map[string]interface{}
	handle, _ := route["handle"].([]interface{})
	for _, h := range handle {
		handlerMap, ok := h.(map[string]interface{})
		if !ok {
			continue
		}
		handlers = append(handlers, handlerMap)
	}
	return handlers
}

func (c *CaddyClient) extractUpstreams(route map[string]interface{}) []string {
	var upstreams []string
	handle, _ := route["handle"].([]interface{})
	for _, h := range handle {
		handlerMap, ok := h.(map[string]interface{})
		if !ok {
			continue
		}
		if handlerMap["handler"] == "reverse_proxy" {
			ups, _ := handlerMap["upstreams"].([]interface{})
			for _, u := range ups {
				upMap, ok := u.(map[string]interface{})
				if ok {
					dial, _ := upMap["dial"].(string)
					if dial != "" {
						upstreams = append(upstreams, dial)
					}
				}
			}
		}
		if handlerMap["handler"] == "subroute" {
			subRoutes, _ := handlerMap["routes"].([]interface{})
			for _, sr := range subRoutes {
				srMap, ok := sr.(map[string]interface{})
				if ok {
					srHandlers, _ := srMap["handle"].([]interface{})
					for _, srH := range srHandlers {
						srHMap, ok := srH.(map[string]interface{})
						if !ok {
							continue
						}
						if srHMap["handler"] == "reverse_proxy" {
							ups, _ := srHMap["upstreams"].([]interface{})
							for _, u := range ups {
								upMap, ok := u.(map[string]interface{})
								if ok {
									dial, _ := upMap["dial"].(string)
									if dial != "" {
										upstreams = append(upstreams, dial)
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return upstreams
}

func (c *CaddyClient) hasBasicAuth(route map[string]interface{}) bool {
	handlers := c.extractHandlers(route)
	for _, h := range handlers {
		if h["handler"] == "authentication" {
			return true
		}
		if h["handler"] == "subroute" {
			subRoutes, _ := h["routes"].([]interface{})
			for _, sr := range subRoutes {
				srMap, ok := sr.(map[string]interface{})
				if !ok {
					continue
				}
				srHandlers, _ := srMap["handle"].([]interface{})
				for _, srH := range srHandlers {
					srHMap, ok := srH.(map[string]interface{})
					if !ok {
						continue
					}
					if srHMap["handler"] == "authentication" {
						return true
					}
				}
			}
		}
	}
	return false
}
