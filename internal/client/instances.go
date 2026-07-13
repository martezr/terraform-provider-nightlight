package client

import "fmt"

func (c *Client) ListInstances() ([]Instance, error) {
	var out []Instance
	return out, c.request("GET", "/api/v1/instances", nil, &out)
}

func (c *Client) GetInstance(id string) (*Instance, error) {
	var out Instance
	err := c.request("GET", fmt.Sprintf("/api/v1/instances/%s", id), nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CreateInstance(in Instance) (*Instance, error) {
	var out Instance
	err := c.request("POST", "/api/v1/instances", in, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateInstance(id string, in Instance) (*Instance, error) {
	var out Instance
	err := c.request("PUT", fmt.Sprintf("/api/v1/instances/%s", id), in, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteInstance(id string) error {
	return c.request("DELETE", fmt.Sprintf("/api/v1/instances/%s", id), nil, nil)
}

func (c *Client) SendInstanceConsoleKeys(id string, cmd Command) error {
	return c.request("POST", fmt.Sprintf("/api/v1/instances/%s/sendkeys", id), cmd, nil)
}
