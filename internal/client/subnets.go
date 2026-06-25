package client

import "fmt"

func (c *Client) ListSubnets() ([]Subnet, error) {
	var out []Subnet
	return out, c.request("GET", "/api/v1/subnets", nil, &out)
}

func (c *Client) GetSubnet(id string) (*Subnet, error) {
	var out Subnet
	err := c.request("GET", fmt.Sprintf("/api/v1/subnets/%s", id), nil, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CreateSubnet(in Subnet) (*Subnet, error) {
	var out Subnet
	err := c.request("POST", "/api/v1/subnets", in, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) UpdateSubnet(id string, in Subnet) (*Subnet, error) {
	var out Subnet
	err := c.request("PUT", fmt.Sprintf("/api/v1/subnets/%s", id), in, &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) DeleteSubnet(id string) error {
	return c.request("DELETE", fmt.Sprintf("/api/v1/subnets/%s", id), nil, nil)
}
