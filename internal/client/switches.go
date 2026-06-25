package client

import "fmt"

func (c *Client) ListSwitches() ([]Switch, error) {
	var out []Switch
	return out, c.request("GET", "/api/v1/switches", nil, &out)
}

func (c *Client) GetSwitch(id string) (*Switch, error) {
	var out Switch
	err := c.request("GET", fmt.Sprintf("/api/v1/switches/%s", id), nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CreateSwitch(in Switch) (*Switch, error) {
	var out Switch
	err := c.request("POST", "/api/v1/switches", in, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateSwitch(id string, in Switch) (*Switch, error) {
	var out Switch
	err := c.request("PUT", fmt.Sprintf("/api/v1/switches/%s", id), in, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteSwitch(id string) error {
	return c.request("DELETE", fmt.Sprintf("/api/v1/switches/%s", id), nil, nil)
}
