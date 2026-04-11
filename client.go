package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type CaddyClient struct {
	baseURL string
	client  *http.Client
}

func NewCaddyClient(host string) *CaddyClient {
	if host == "" {
		host = "localhost:2019"
	}
	return &CaddyClient{
		baseURL: "http://" + host,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *CaddyClient) doRequest(method, path string, body []byte) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.client.Do(req)
}

func (c *CaddyClient) GetConfig() (map[string]interface{}, error) {
	resp, err := c.doRequest("GET", "/config/", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	var config map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, err
	}
	return config, nil
}

func (c *CaddyClient) SetConfig(path string, value interface{}) error {
	data, _ := json.Marshal(value)
	resp, err := c.doRequest("PUT", "/config/"+path, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *CaddyClient) DeleteConfig(path string) error {
	resp, err := c.doRequest("DELETE", "/config/"+path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *CaddyClient) PostConfig(path string, value interface{}) error {
	data, _ := json.Marshal(value)
	resp, err := c.doRequest("POST", "/config/"+path, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *CaddyClient) LoadConfig(config interface{}) error {
	data, _ := json.Marshal(config)
	resp, err := c.doRequest("POST", "/load", data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

func (c *CaddyClient) Stop() error {
	resp, err := c.doRequest("GET", "/stop", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *CaddyClient) GetConfigAtPath(path string) (interface{}, error) {
	resp, err := c.doRequest("GET", "/config/"+path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 400 {
		return nil, nil
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}
	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *CaddyClient) AdaptCaddyfile(caddyfile string) (map[string]interface{}, error) {
	resp, err := c.doRequest("POST", "/adapt", []byte(caddyfile))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("adapt error %d: %s", resp.StatusCode, string(body))
	}
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *CaddyClient) ValidateConfig(config map[string]interface{}) error {
	data, _ := json.Marshal(config)
	resp, err := c.doRequest("POST", "/validate", data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("validation failed: %s", string(body))
	}
	return nil
}
