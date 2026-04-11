#!/bin/bash
# Test script for caddy-mcp

echo 'Testing list_sites...'
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"list_sites","arguments":{}}}' | ./caddy-mcp

echo ''
echo 'Testing caddy_status...'
echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"caddy_status","arguments":{}}}' | ./caddy-mcp
