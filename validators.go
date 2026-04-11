package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func validateDomain(domain string) error {
	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}
	domainRegex := regexp.MustCompile(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)*[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`)
	if !domainRegex.MatchString(domain) {
		return fmt.Errorf("invalid domain format: %s", domain)
	}
	return nil
}

func validatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}
	if strings.Contains(path, "..") {
		return fmt.Errorf("path cannot contain parent directory references (..)")
	}
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("path must start with /")
	}
	return nil
}

func validatePort(port string) error {
	if port == "" {
		return fmt.Errorf("port cannot be empty")
	}
	p, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("invalid port number: %s", port)
	}
	if p < 1 || p > 65535 {
		return fmt.Errorf("port must be between 1 and 65535, got: %d", p)
	}
	return nil
}

func validateHeaderName(name string) error {
	if name == "" {
		return fmt.Errorf("header name cannot be empty")
	}
	headerRegex := regexp.MustCompile("^[a-zA-Z0-9!#$%&'*+\\-\\.^_`\\|~]+$")
	if !headerRegex.MatchString(name) {
		return fmt.Errorf("invalid header name format: %s (must be valid HTTP token characters only)", name)
	}
	return nil
}
