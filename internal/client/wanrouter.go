package client

import "fmt"

func (c *Client) GetWanRouterConfig() (*WanRouterConfig, error) {
	var out WanRouterConfig
	if err := c.request("GET", "/api/v1/wanrouter/config", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateWanRouterConfig(in WanRouterConfig) (*WanRouterConfig, error) {
	var out WanRouterConfig
	if err := c.request("PUT", "/api/v1/wanrouter/config", in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetRouterConfig(id string) (*RouterConfig, error) {
	var out RouterConfig
	if err := c.request("GET", fmt.Sprintf("/api/v1/routers/%s/config", id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateRouterConfig(id string, in RouterConfig) (*RouterConfig, error) {
	var out RouterConfig
	if err := c.request("PUT", fmt.Sprintf("/api/v1/routers/%s/config", id), in, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
