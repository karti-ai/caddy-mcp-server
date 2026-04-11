# Caddy MCP Server

A comprehensive Model Context Protocol (MCP) server for managing Caddy web server via the Admin API.

## Features

Complete Caddy management with 21 powerful tools:

### Site Management
| Tool | Description |
|------|-------------|
| `list_sites` | List all configured sites with upstreams & auth status |
| `get_site` | Get detailed configuration for a specific site |
| `get_site_routes` | Get all routes (including path-based) for a site |
| `add_site` | Add a new site with reverse proxy |
| `remove_site` | Remove a site configuration |
| `update_upstream` | Update the upstream address for a site |

### Path-Based Routing
| Tool | Description |
|------|-------------|
| `add_path_route` | Add path-based route (e.g., `/api/*` → localhost:3000) |
| `remove_path_route` | Remove a path-based route |

### Authentication
| Tool | Description |
|------|-------------|
| `add_basic_auth` | Add basic authentication to a site |
| `remove_basic_auth` | Remove basic authentication |

### Headers
| Tool | Description |
|------|-------------|
| `add_request_header` | Add header sent to upstream |
| `add_response_header` | Add header sent to client |

### File Server
| Tool | Description |
|------|-------------|
| `add_file_server` | Add static file server |
| `update_file_root` | Update file server root directory |

### Review Environments
| Tool | Description |
|------|-------------|
| `add_wip_environment` | Add WIP review env to review.example.com |

### Configuration
| Tool | Description |
|------|-------------|
| `export_config` | Export full Caddy JSON config |
| `adapt_caddyfile` | Convert Caddyfile to JSON |
| `validate_config` | Validate config without applying |

### Server Control
| Tool | Description |
|------|-------------|
| `caddy_status` | Check Caddy health |
| `stop_caddy` | Gracefully stop Caddy |

## Usage

```bash
# Build
go build -o caddy-mcp .

# Run locally
./caddy-mcp

# Run against remote Caddy (via SSH tunnel)
ssh -L 2019:localhost:2019 server-1 &
./caddy-mcp
```

## Environment

- `CADDY_HOST` - Caddy admin API host (default: `localhost:2019`)
- `REVIEW_DOMAIN` - Domain for WIP review environments (default: `review.example.com`)

## Architecture

- `client.go` - Caddy Admin API client
- `server.go` - MCP server setup
- `handlers.go` - Tool implementations
- `site_ops.go` - Site read operations
- `site_modify.go` - Site write operations
- `path_routes.go` - Path-based routing
- `auth_ops.go` - Authentication management
- `headers.go` - Header manipulation
- `fileserver.go` - Static file serving

## Security

### Password Transmission
**WARNING**: Passwords are transmitted in plaintext from the MCP client to this server.
This is an architectural limitation of the Model Context Protocol (MCP).
While Caddy stores passwords hashed (bcrypt), the password travels unencrypted
through the MCP communication channel.

### Caddy Admin API
The Caddy admin API (`localhost:2019` by default) uses HTTP, not HTTPS.
When running this MCP server remotely, use SSH port forwarding to secure the connection:
```bash
ssh -L 2019:localhost:2019 <remote-server> &
./caddy-mcp
```

## Examples

```bash
# List all sites
./caddy-mcp
# Send: {"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"list_sites","arguments":{}}}

# Add a WIP environment
# {"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"add_wip_environment","arguments":{"project":"my-project","port":3000}}}
```
