package client

import "fmt"

func (c *Client) ListSites() ([]Site, error) {
	var out []Site
	return out, c.request("GET", "/api/v1/sites", nil, &out)
}

func (c *Client) GetSite(id string) (*Site, error) {
	var out Site
	err := c.request("GET", fmt.Sprintf("/api/v1/sites/%s", id), nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CreateSite(in Site) (*Site, error) {
	var out Site
	err := c.request("POST", "/api/v1/sites", in, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateSite(id string, in Site) (*Site, error) {
	var out Site
	err := c.request("PUT", fmt.Sprintf("/api/v1/sites/%s", id), in, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteSite(id string) error {
	return c.request("DELETE", fmt.Sprintf("/api/v1/sites/%s", id), nil, nil)
}
