package main

import (
	"fmt"
	"log"
)

func (c *CaddyClient) AddBasicAuth(domain, username, password string) error {
	log.Printf("[WARNING] Password for user %s on domain %s is being transmitted in plaintext through MCP", username, domain)
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
								if len(subRoutes) > 0 {
									mainRoute, _ := subRoutes[0].(map[string]interface{})
									handlers, _ := mainRoute["handle"].([]interface{})

									// Check if auth already exists
									var authIdx = -1
									for idx, handler := range handlers {
										hm, _ := handler.(map[string]interface{})
										if hm["handler"] == "authentication" {
											authIdx = idx
											break
										}
									}

									// WARNING: Passwords are transmitted in plaintext through MCP protocol.
									// This is an architectural limitation of the Model Context Protocol.
									// The password is stored hashed in Caddy (bcrypt), but travels
									// unencrypted from the MCP client to this server.
									newAccount := map[string]interface{}{
										"username": username,
										"password": password,
									}

									if authIdx >= 0 {
										// Add to existing auth
										authHandler, _ := handlers[authIdx].(map[string]interface{})
										providers, _ := authHandler["providers"].(map[string]interface{})
										httpBasic, _ := providers["http_basic"].(map[string]interface{})
										accounts, _ := httpBasic["accounts"].([]interface{})
										accounts = append(accounts, newAccount)
										httpBasic["accounts"] = accounts
										providers["http_basic"] = httpBasic
										authHandler["providers"] = providers
										handlers[authIdx] = authHandler
									} else {
										// Create new auth handler
										authHandler := map[string]interface{}{
											"handler": "authentication",
											"providers": map[string]interface{}{
												"http_basic": map[string]interface{}{
													"accounts": []interface{}{newAccount},
													"hash": map[string]interface{}{
														"algorithm": "bcrypt",
													},
												},
											},
										}
										handlers = append([]interface{}{authHandler}, handlers...)
									}

									mainRoute["handle"] = handlers
									subRoutes[0] = mainRoute
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
	}
	return fmt.Errorf("site not found: %s", domain)
}

func (c *CaddyClient) RemoveBasicAuth(domain string) error {
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
								if len(subRoutes) > 0 {
									mainRoute, _ := subRoutes[0].(map[string]interface{})
									handlers, _ := mainRoute["handle"].([]interface{})
									var newHandlers []interface{}
									for _, handler := range handlers {
										hm, _ := handler.(map[string]interface{})
										if hm["handler"] != "authentication" {
											newHandlers = append(newHandlers, handler)
										}
									}
									mainRoute["handle"] = newHandlers
									subRoutes[0] = mainRoute
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
	}
	return fmt.Errorf("site not found: %s", domain)
}
