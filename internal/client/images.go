package client

import "fmt"

func (c *Client) ListImages() ([]Image, error) {
	var out []Image
	return out, c.request("GET", "/api/v1/images", nil, &out)
}

func (c *Client) GetImage(id string) (*Image, error) {
	var out Image
	err := c.request("GET", fmt.Sprintf("/api/v1/images/%s", id), nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
